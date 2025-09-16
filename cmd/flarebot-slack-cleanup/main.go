package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "embed"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/Clever/flarebot/jira"
	"github.com/Clever/kayvee-go/v7/logger"

	slk "github.com/slack-go/slack"
)

// generate launch config
//go:generate sh -c "$PWD/../../bin/launch-gen -o launch.go -p main $PWD/../../launch/flarebot-slack-cleanup.yml"

// generate kv config bytes for setting up log routing
//
//go:embed kvconfig.yml
var kvconfig []byte

type SlackClient interface {
	ArchiveConversation(channelID string) error
	JoinConversation(channelID string) (*slk.Channel, string, []string, error)
	GetConversations(input *slk.GetConversationsParameters) ([]slk.Channel, string, error)
}

type JiraClient interface {
	GetTicketByKey(key string) (*jira.Ticket, error)
	SetLabel(ticket *jira.Ticket, label string) error
}

// Handler encapsulates the external dependencies of the lambda function.
// The example here demonstrates the case where the handler logic involves communicating with S3.
type Handler struct {
	slackClient  SlackClient
	jiraClient   JiraClient
	launchConfig LaunchConfig
}

// Constants for the handler
const (
	defaultPageSize      = 200
	jiraArchivedLabel    = "archived"
	defaultRetryAttempts = 3
	defaultRetryDelay    = 1 * time.Second
)

// Handle is invoked by the Lambda runtime with the contents of the function input.
func (h Handler) Handle(ctx context.Context) error {
	// create a request-specific logger, attach it to ctx, and add the Lambda request ID.
	ctx = logger.NewContext(ctx, logger.New(os.Getenv("APP_NAME")))
	if lambdaContext, ok := lambdacontext.FromContext(ctx); ok {
		logger.FromContext(ctx).AddContext("aws-request-id", lambdaContext.AwsRequestID)
	}

	flareChannelPrefix := h.launchConfig.Env.FlareChannelPrefix
	threshold, err := strconv.Atoi(h.launchConfig.Env.ChannelAgeThreshold)
	if err != nil {
		return err
	}
	dryRun, err := strconv.ParseBool(h.launchConfig.Env.DryRun)
	if err != nil {
		return err
	}
	logger.FromContext(ctx).InfoD("starting-cleanup", logger.M{"flareChannelPrefix": flareChannelPrefix, "threshold": threshold, "dryRun": dryRun})

	var cursor string
	for {
		slkInput := &slk.GetConversationsParameters{
			ExcludeArchived: "true",
			Limit:           defaultPageSize,
		}

		if cursor != "" {
			slkInput.Cursor = cursor
		}

		response, err := retrySlack(ctx, defaultRetryAttempts, defaultRetryDelay, func() (interface{}, error) {
			channels, nextCursor, err := h.slackClient.GetConversations(slkInput)
			if err != nil {
				return nil, err
			}
			return &struct {
				Channels   []slk.Channel
				NextCursor string
			}{channels, nextCursor}, nil
		})
		if err != nil {
			return err
		}

		conversations := response.(*struct {
			Channels   []slk.Channel
			NextCursor string
		})

		for _, channel := range conversations.Channels {
			if (strings.HasPrefix(channel.Name, flareChannelPrefix)) && isOlderThanThreshold(int64(channel.Created), threshold) {
				logger.FromContext(ctx).DebugD("archiving-channel", logger.M{"channel": channel.Name})
				if !dryRun {
					err = h.cleanupSlackChannel(ctx, channel)
					if err != nil {
						logger.FromContext(ctx).ErrorD("error-archiving-channel", logger.M{"channelName": channel.Name, "channelID": channel.ID, "error": err.Error()})
						continue
					}
				}
			}
		}

		if conversations.NextCursor == "" {
			break
		}

		cursor = conversations.NextCursor
	}

	return nil
}

func (h Handler) cleanupSlackChannel(ctx context.Context, channel slk.Channel) error {
	_, err := retrySlack(ctx, defaultRetryAttempts, defaultRetryDelay, func() (interface{}, error) {
		err := h.slackClient.ArchiveConversation(channel.ID)

		if err != nil && err.Error() == "not_in_channel" {
			logger.FromContext(ctx).DebugD("joining-channel", logger.M{"channel": channel.Name})
			_, joinErr := retrySlack(ctx, defaultRetryAttempts, defaultRetryDelay, func() (interface{}, error) {
				_, _, _, joinErr := h.slackClient.JoinConversation(channel.ID)
				return nil, joinErr
			})
			if joinErr != nil {
				return nil, joinErr
			}
		}

		return nil, err
	})
	if err != nil {
		return err
	}

	ticket, err := h.jiraClient.GetTicketByKey(strings.ToUpper(channel.Name))
	if err != nil {
		return err
	}

	err = h.jiraClient.SetLabel(ticket, jiraArchivedLabel)
	if err != nil {
		return err
	}
	return nil
}

func isOlderThanThreshold(timestamp int64, threshold int) bool {
	creationTime := time.Unix(timestamp, 0)
	cutoffTime := time.Now().Add(-time.Duration(threshold) * 24 * time.Hour)
	return creationTime.Before(cutoffTime)
}

func retrySlack(ctx context.Context, attempts int, sleep time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	var err error
	for i := 0; i < attempts; i++ {
		res, err := fn()
		if err == nil {
			return res, nil
		}

		var te *slk.RateLimitedError
		if errors.As(err, &te) {
			logger.FromContext(ctx).InfoD("slack-ratelimit-error", logger.M{"error": err.Error()})
			time.Sleep(te.RetryAfter)
			continue
		}

		if i < attempts-1 {
			time.Sleep(sleep)
			sleep *= 2
		}
	}
	return nil, err
}

func main() {
	ctx := context.Background()
	if err := logger.SetGlobalRoutingFromBytes(kvconfig); err != nil {
		log.Fatalf("Error setting kvconfig: %v", err)
	}
	lg := logger.FromContext(ctx)

	launchConfig := InitLaunchConfig(nil)
	slackClient := slk.New(launchConfig.Env.SlackBotToken)
	jiraServer := jira.JiraServer{
		Origin:    launchConfig.Env.JiraOrigin,
		Username:  launchConfig.Env.JiraUsername,
		Password:  launchConfig.Env.JiraPassword,
		ProjectID: launchConfig.Env.JiraProjectID,
	}

	handler := Handler{
		slackClient:  slackClient,
		jiraClient:   &jiraServer,
		launchConfig: launchConfig,
	}

	if os.Getenv("IS_LOCAL") == "true" {
		lg.InfoD("running locally", logger.M{})
		err := handler.Handle(ctx)
		if err != nil {
			lg.ErrorD("error on handle", logger.M{"err": err.Error()})
			os.Exit(1)
		}
	} else {
		_, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			log.Fatalf("Error loading AWS config: %v", err)
		}
		lambda.Start(handler.Handle)
	}
}
