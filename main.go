package main

import (
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	//	"encoding/json"
	"bytes"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v2"
	"google.golang.org/api/googleapi/transport"
	"net/http"
)

// the regexp for the fire a flare Slack message
const fireFlareCommandRegexp string = "^.*: fire (?:a )?flare [pP]([012]) *(.*)"

type jiraTicket struct {
	Url string
	Key string
}

type googleDoc struct {
	Url string
}

func createJiraTicket(priority int, topic string) *jiraTicket {
	// POST /rest/api/2/issue
	// https://docs.atlassian.com/jira/REST/latest/#api/2/issue-createIssue

	request := &map[string]interface{}{}

	fmt.Printf("%v", request)

	return &jiraTicket{
		Url: "http://example.com/foo",
		Key: "flare-test-4243",
	}
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

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{
		Transport: &transport.APIKey{Key: google_flarebot_access_token},
	})

	// instantiate the Google Drive client
	oauthClient := config.Client(ctx, token)
	service, err := drive.New(oauthClient)
	if err != nil {
		log.Fatalf("Unable to create Drive service: %v", err)
	}

	// copy the template doc to a new doc
	file, err := service.Files.Copy(google_template_doc_id, &drive.File{Title: topic}).Do()
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

func main() {
	token := decodeOAuthToken(os.Getenv("SLACK_FLAREBOT_ACCESS_TOKEN"))
	domain := os.Getenv("SLACK_DOMAIN")
	username := os.Getenv("SLACK_USERNAME")

	client, err := NewClient(token.AccessToken, domain, username)
	if err != nil {
		panic(err)
	}

	client.Respond(".*", func(msg *Message, params [][]string) {
		re := regexp.MustCompile(fireFlareCommandRegexp)

		// doesn't match?
		matches := re.FindStringSubmatch(msg.Text)
		if len(matches) == 0 {
			msg.Respond("I'm sorry Jim, I don't understand.")
			return
		}

		fmt.Printf("FROM: %v\n", msg)

		// for now matches are indexed
		priority, _ := strconv.Atoi(matches[1])
		topic := matches[2]

		// ok it matches
		client.Send("OK, let me get my matches...", msg.Channel)

		ticket := createJiraTicket(priority, topic)

		if ticket == nil {
			panic("no JIRA ticket created")
		}

		doc := createGoogleDoc(ticket.Url, ticket.Key, priority, topic)

		if doc == nil {
			panic("No google doc created")
		}

		// msg.Respond(fmt.Sprintf("Create a Google Doc for you: %s", doc.Url))

		channelId, _ := createSlackChannel(client, ticket.Key)

		// set up the Flare room
		client.Send(fmt.Sprintf("JIRA Ticket: %s", ticket.Url), channelId)
		client.Send(fmt.Sprintf("Facts Docs: %s", doc.Url), channelId)

		// announce the specific Flare room in the overall Flares room
		client.Send(fmt.Sprintf("@channel: Flare fired. Please visit #%s", ticket.Key), msg.Channel)
	})

	panic(client.Run())
}
