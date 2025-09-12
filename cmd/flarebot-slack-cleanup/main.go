package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
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

// generate kv config bytes for setting up log routing
//
//go:embed kvconfig.yml
var kvconfig []byte

// Handler encapsulates the external dependencies of the lambda function.
// The example here demonstrates the case where the handler logic involves communicating with S3.
type Handler struct {
	slackClient *slk.Client
	jiraClient  *jira.JiraServer
}

const (
	DEFAULT_PAGE_SIZE      = 200
	ARCHIVED_LABEL         = "archived"
	DEFAULT_RETRY_ATTEMPTS = 3
	DEFAULT_RETRY_DELAY    = 1 * time.Second
)

// Handle is invoked by the Lambda runtime with the contents of the function input.
func (h Handler) Handle(ctx context.Context) error {
	// create a request-specific logger, attach it to ctx, and add the Lambda request ID.
	ctx = logger.NewContext(ctx, logger.New(os.Getenv("APP_NAME")))
	if lambdaContext, ok := lambdacontext.FromContext(ctx); ok {
		logger.FromContext(ctx).AddContext("aws-request-id", lambdaContext.AwsRequestID)
	}
	logger.FromContext(ctx).InfoD("starting-cleanup", logger.M{})

	var cursor string

	for {
		input := &slk.GetConversationsParameters{
			ExcludeArchived: "true",
			Limit:           DEFAULT_PAGE_SIZE,
		}

		if cursor != "" {
			input.Cursor = cursor
		}

		response, err := retrySlack(ctx, DEFAULT_RETRY_ATTEMPTS, DEFAULT_RETRY_DELAY, func() (interface{}, error) {
			channels, nextCursor, err := h.slackClient.GetConversations(input)
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

		flareChannelRegex := flareChannelRegex()
		for _, channel := range conversations.Channels {
			if flareChannelRegex.MatchString(channel.Name) && isOlderThan180Days(int64(channel.Created)) {
				logger.FromContext(ctx).InfoD("archiving-channel", logger.M{"channel": channel.Name})
				fmt.Println("archiving channel", channel.Name)
				// err = h.CleanupSlackChannel(channel)
				// if err != nil {
				// 	return err
				// }
			}
			fmt.Println("channel", channel.Name)
		}

		if conversations.NextCursor == "" {
			break
		}

		cursor = conversations.NextCursor
	}

	return nil
}

func (h Handler) CleanupSlackChannel(channel slk.Channel) error {
	_, err := retrySlack(context.Background(), DEFAULT_RETRY_ATTEMPTS, DEFAULT_RETRY_DELAY, func() (interface{}, error) {
		err := h.slackClient.ArchiveConversation(channel.ID)
		return nil, err
	})
	if err != nil {
		return err
	}
	ticket, err := h.jiraClient.GetTicketByKey(strings.ToUpper(channel.Name))
	if err != nil {
		return err
	}
	err = h.jiraClient.SetLabel(ticket, ARCHIVED_LABEL)
	if err != nil {
		return err
	}
	return nil
}

func flareChannelRegex() *regexp.Regexp {
	prefix := os.Getenv("FLARE_CHANNEL_PREFIX")
	escapedPrefix := regexp.QuoteMeta(prefix)
	pattern := "^" + escapedPrefix + "\\d+$"
	return regexp.MustCompile(pattern)
}

func isOlderThan180Days(timestamp int64) bool {
	creationTime := time.Unix(timestamp, 0)
	cutoffTime := time.Now().Add(-180 * 24 * time.Hour)
	return creationTime.Before(cutoffTime)
}

func retrySlack(ctx context.Context, attempts int, sleep time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	var err error
	for i := 0; i < attempts; i++ {
		res, err := fn()
		if err == nil {
			return res, nil
		}

		var te slk.RateLimitedError
		if errors.As(err, &te) {
			logger.FromContext(ctx).InfoD("ratelimit-error", logger.M{"error": err.Error()})
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

	_, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("Error loading AWS config: %v", err)
	}
	token := os.Getenv("SLACK_BOT_TOKEN")
	slackClient := slk.New(token)
	jiraServer := jira.JiraServer{
		Origin:    os.Getenv("JIRA_ORIGIN"),
		Username:  os.Getenv("JIRA_USERNAME"),
		Password:  os.Getenv("JIRA_PASSWORD"),
		ProjectID: os.Getenv("JIRA_PROJECT_ID"),
	}

	handler := Handler{
		slackClient: slackClient,
		jiraClient:  &jiraServer,
	}

	if os.Getenv("IS_LOCAL") == "true" {
		// Update input as needed to debug
		lg.InfoD("running locally", logger.M{})
		err := handler.Handle(ctx)
		if err != nil {
			lg.ErrorD("error on handle", logger.M{"err": err.Error()})
			os.Exit(1)
		}
	} else {
		lambda.Start(handler.Handle)
	}
}
