package main

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v2"
	"google.golang.org/api/googleapi/transport"

	"github.com/Clever/flarebot/jira"
)

// the regexp for the fire a flare Slack message
const fireFlareCommandRegexp string = "^.* fire (?:a )?flare [pP]([012]) *(.*)"

// a regexp for testing without doing anything
const testCommandRegexp string = "^.* test *(.*)"

type googleDoc struct {
	Url string
}

func decodeOAuthToken(tokenString string) *oauth2.Token {
	tokenBytes, _ := base64.StdEncoding.DecodeString(tokenString)
	tokenBytesBuffer := bytes.NewBuffer(tokenBytes)
	dec := gob.NewDecoder(tokenBytesBuffer)
	token := new(oauth2.Token)
	dec.Decode(token)

	return token
}

func createGoogleDoc(jiraTicketURL string, flareKey string, priority int, topic string) *googleDoc {
	// https://github.com/google/google-api-go-client/blob/master/examples/drive.go#L33

	google_client_id := os.Getenv("GOOGLE_CLIENT_ID")
	google_client_secret := os.Getenv("GOOGLE_CLIENT_SECRET")
	google_flarebot_access_token := os.Getenv("GOOGLE_FLAREBOT_ACCESS_TOKEN")
	google_template_doc_id := os.Getenv("GOOGLE_TEMPLATE_DOC_ID")

	// decode the token back into a token
	token := decodeOAuthToken(google_flarebot_access_token)

	var config = &oauth2.Config{
		ClientID:     google_client_id,
		ClientSecret: google_client_secret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{drive.DriveScope},
	}

	// OAuth context with API key
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{
		Transport: &transport.APIKey{Key: google_flarebot_access_token},
	})

	// instantiate the Google Drive client
	oauthClient := config.Client(ctx, token)
	service, err := drive.New(oauthClient)
	if err != nil {
		log.Fatalf("Unable to create Drive service: %v", err)
	}

	google_doc_title := fmt.Sprintf("%s: %s", flareKey, topic)

	// copy the template doc to a new doc
	file, err := service.Files.Copy(google_template_doc_id, &drive.File{
		Title: google_doc_title,
	}).Do()

	if err != nil {
		fmt.Printf("An error occurred: %v\n", err)
		return nil
	}

	// make it editable by the entire organization
	permissions, err := service.Permissions.List(file.Id).Do()
	if err != nil {
		fmt.Printf("An error occurred: %v\n", err)
		return nil
	}

	// look for the domain permission and update it to "Writer"
	for _, perm := range permissions.Items {
		if perm.Type == "domain" {
			fmt.Println("found the permission")
			fmt.Println(perm)
			perm.Role = "writer"
			_, err = service.Permissions.Update(file.Id, perm.Id, perm).Do()
			if err != nil {
				fmt.Printf("error in permission: %v\n", err)
			}
		}
	}

	return &googleDoc{
		Url: file.AlternateLink,
	}
}

// returns the slack Channel ID
func createSlackChannel(client *Client, flareKey string) (string, error) {
	channel, err := client.api.CreateChannel(flareKey)
	if err != nil {
		return "", err
	} else {
		return channel.ID, nil
	}
}

func getSlackUserInfo(client *Client) (map[string]interface{}, error) {
	return nil, nil
}

func main() {
	// JIRA service
	var JiraServer jira.JiraService = &jira.JiraServer{
		Origin: os.Getenv("JIRA_ORIGIN"),
		Username: os.Getenv("JIRA_USERNAME"),
		Password: os.Getenv("JIRA_PASSWORD"),
		ProjectID: os.Getenv("JIRA_PROJECT_ID"),
		IssueTypeID: os.Getenv("JIRA_ISSUETYPE_ID"),
		PriorityIDs: strings.Split(os.Getenv("JIRA_PRIORITIES"), ","),
	}
	
	// Link to flare resources
	resources_url := os.Getenv("FLARE_RESOURCES_URL")

	// Slack connection params
	token := decodeOAuthToken(os.Getenv("SLACK_FLAREBOT_ACCESS_TOKEN"))
	domain := os.Getenv("SLACK_DOMAIN")
	username := os.Getenv("SLACK_USERNAME")

	client, err := NewClient(token.AccessToken, domain, username)
	if err != nil {
		panic(err)
	}

	// regular expressions
	re := regexp.MustCompile(fireFlareCommandRegexp)
	reTest := regexp.MustCompile(testCommandRegexp)

	expectedChannel := os.Getenv("SLACK_CHANNEL")

	client.Respond(".*", func(msg *Message, params [][]string) {
		// wrong channel?
		if msg.Channel != expectedChannel {
			// removing this because it doesn't really happen and it makes testing harder.
			// client.Send("I only respond in the #flares channel.", msg.Channel)
			return
		}

		// doesn't match?
		matches := re.FindStringSubmatch(msg.Text)

		fmt.Println("channel: ", msg.Channel)

		if len(matches) == 0 {
			// matches test?
			testMatches := reTest.FindStringSubmatch(msg.Text)
			
			if len(testMatches) > 0 {
				author, _ := msg.AuthorUser()
				client.Send(fmt.Sprintf("I see you're using the test command. Excellent: %s", author.Profile.Email), msg.Channel)

				user, _ := JiraServer.GetUserByEmail(author.Profile.Email)

				client.Send(fmt.Sprintf("JIRA username is %s", user.Name), msg.Channel)

				channel, _ := client.api.GetChannelInfo(msg.Channel)

				client.Send(fmt.Sprintf("this channel is %s", channel.Name), msg.Channel)

				ticket, _ := JiraServer.GetTicketByKey("flare-165")

				fmt.Println(ticket)
				
				return
			}

			client.Send("The only command I know is: fire a flare p0/p1/p2 <topic>", msg.Channel)
			return
		}

		// for now matches are indexed
		priority, _ := strconv.Atoi(matches[1])
		topic := matches[2]

		// ok it matches
		client.Send("OK, let me get my flaregun", msg.Channel)

		author, _ := msg.AuthorUser()
		assigneeUser, _ := JiraServer.GetUserByEmail(author.Profile.Email)
		ticket, _ := JiraServer.CreateTicket(priority, topic, assigneeUser)

		if ticket == nil {
			panic("no JIRA ticket created")
		}

		doc := createGoogleDoc(ticket.Url, ticket.Key, priority, topic)

		if doc == nil {
			panic("No google doc created")
		}

		channelId, _ := createSlackChannel(client, strings.ToLower(ticket.Key))

		// set up the Flare room
		client.Send(fmt.Sprintf("JIRA ticket: %s", ticket.Url), channelId)
		client.Send(fmt.Sprintf("Facts docs: %s", doc.Url), channelId)
		client.Send(fmt.Sprintf("Flare resources: %s", resources_url), channelId)

		// announce the specific Flare room in the overall Flares room
		client.Send(fmt.Sprintf("@channel: Flare fired. Please visit #%s", strings.ToLower(ticket.Key)), msg.Channel)
	})

	panic(client.Run())
}
