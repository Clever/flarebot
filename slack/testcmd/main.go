package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Clever/flarebot/slack"
	"github.com/spf13/cobra"
)

type slackCommand struct {
	client *slack.Client
}

func (sc *slackCommand) getChannelID(name string) string {
	channels, err := sc.client.API.GetChannels(true)
	if err != nil {
		log.Fatal(err)
	}
	for _, channel := range channels {
		if channel.Name == name {
			return channel.ID
		}
	}
	log.Fatalf("Failed to find channel '%s'", name)
	return ""
}

func (sc *slackCommand) CreateChannelCommand(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		log.Fatalf("A channel name must be provided\n")
	}

	channel, err := sc.client.CreateChannel(args[0])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Channel `%s` created with id: %s\n", channel.Name, channel.ID)
}

func (sc *slackCommand) PostMessageCommand(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		log.Fatalf("A channel name and quoted message must be provided\n")
	}

	sc.client.Send(args[1], sc.getChannelID(args[0]))
}

func (sc *slackCommand) PinMessageCommand(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		log.Fatalf("A channel name and quoted message must be provided\n")
	}

	sc.client.Pin(args[1], sc.getChannelID(args[0]))
}

func getEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("Failed to get '%s' from the environment", name)
	}
	return value
}

func main() {
	// Slack connection params
	var accessToken string
	if os.Getenv("SLACK_LEGACY_TOKEN") != "" {
		accessToken = getEnv("SLACK_FLAREBOT_ACCESS_TOKEN")
	} else {
		token := slack.DecodeOAuthToken(getEnv("SLACK_FLAREBOT_ACCESS_TOKEN"))
		fmt.Printf("token=%+v accessToken=%s
\n", token, getEnv("SLACK_FLAREBOT_ACCESS_TOKEN"))
		accessToken = token.AccessToken
	}
	domain := getEnv("SLACK_DOMAIN")
	username := getEnv("SLACK_USERNAME")
	fmt.Printf("accessToken=%s domain=%s username=%s\n", accessToken, domain, username)
	client, err := slack.NewClient(accessToken, domain, username)
	if err != nil {
		panic(err)
	}
	sc := slackCommand{
		client: client,
	}

	var cmdCreateChannel = &cobra.Command{
		Use:   "createChannel <channel_name>",
		Short: "create a channel",
		Run:   sc.CreateChannelCommand,
	}

	var cmdPostMessage = &cobra.Command{
		Use:   "postMessage <channel_name> <text>",
		Short: "post a message to a channel",
		Run:   sc.PostMessageCommand,
	}

	var cmdPinMessage = &cobra.Command{
		Use:   "pinMessage <channel_name> <message text>",
		Short: "pin the message matching the text",
		Run:   sc.PinMessageCommand,
	}

	var rootCmd = &cobra.Command{Use: "slack-cli"}
	rootCmd.AddCommand(cmdCreateChannel)
	rootCmd.AddCommand(cmdPostMessage)
	rootCmd.AddCommand(cmdPinMessage)
	rootCmd.Execute()

	// Hacky. Call sleep so that the goroutines in slack.NewClient() have an
	// opportunity to run.
	time.Sleep(1 * time.Second)
	client.Stop()
	fmt.Printf("Goodbye\n")
}
