package main

import (
	"errors"
	"encoding/base64"
	"encoding/gob"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/oauth2"

	"github.com/Clever/flarebot/jira"
	"github.com/Clever/flarebot/googledocs"
)

//
// COMMANDS
//

// fire a flare
const fireFlareCommandRegexp string = "[fF]ire (?:a )?[fF]lare [pP]([012]) *(.*)"

// testing
const testCommandRegexp string = "test *(.*)"

// I am incident lead
const takingLeadCommandRegexp string = "[iI] am incident lead"

// flare mitigated
const flareMitigatedCommandRegexp string = "[Ff]lare (is )?mitigated"

func GetTicketFromCurrentChannel(client *Client, JiraServer *jira.JiraServer, channelID string) (*jira.Ticket, error) {
	// first more info about the channel
	channel, _ := client.api.GetChannelInfo(channelID)

	// then the ticket that matches
	ticket, err := JiraServer.GetTicketByKey(channel.Name)

	if err != nil || ticket.Fields.Project.ID != JiraServer.ProjectID {
		return nil, errors.New("no ticket for this channel")
	}

	return ticket, nil
}

func decodeOAuthToken(tokenString string) *oauth2.Token {
	tokenBytes, _ := base64.StdEncoding.DecodeString(tokenString)
	tokenBytesBuffer := bytes.NewBuffer(tokenBytes)
	dec := gob.NewDecoder(tokenBytesBuffer)
	token := new(oauth2.Token)
	dec.Decode(token)

	return token
}

func main() {
	// JIRA service
	var JiraServer *jira.JiraServer = &jira.JiraServer{
		Origin:      os.Getenv("JIRA_ORIGIN"),
		Username:    os.Getenv("JIRA_USERNAME"),
		Password:    os.Getenv("JIRA_PASSWORD"),
		ProjectID:   os.Getenv("JIRA_PROJECT_ID"),
		IssueTypeID: os.Getenv("JIRA_ISSUETYPE_ID"),
		PriorityIDs: strings.Split(os.Getenv("JIRA_PRIORITIES"), ","),
	}

	// Google Docs service
	GoogleDocsServer, err := googledocs.NewGoogleDocsServer(
		os.Getenv("GOOGLE_CLIENT_ID"),
		os.Getenv("GOOGLE_CLIENT_SECRET"),
		decodeOAuthToken(os.Getenv("GOOGLE_FLAREBOT_ACCESS_TOKEN")),
		os.Getenv("GOOGLE_TEMPLATE_DOC_ID"),
	)
	
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

	expectedChannel := os.Getenv("SLACK_CHANNEL")

	client.Respond(testCommandRegexp, func(msg *Message, params [][]string) {
		author, _ := msg.AuthorUser()
		client.Send(fmt.Sprintf("I see you're using the test command. Excellent: %s", author.Profile.Email), msg.Channel)

		if len(params[0]) > 1 {
			client.Send(fmt.Sprintf("you told me: %s", params[0][1]), msg.Channel)
		}

		user, _ := JiraServer.GetUserByEmail(author.Profile.Email)

		client.Send(fmt.Sprintf("JIRA username is %s", user.Name), msg.Channel)

		channel, _ := client.api.GetChannelInfo(msg.Channel)

		client.Send(fmt.Sprintf("this channel is %s", channel.Name), msg.Channel)

		ticket, _ := JiraServer.GetTicketByKey("flare-165")

		fmt.Println(ticket)
	})

	client.Respond(fireFlareCommandRegexp, func(msg *Message, params [][]string) {
		// wrong channel?
		if msg.Channel != expectedChannel {
			// removing this because it doesn't really happen and it makes testing harder.
			// client.Send("I only respond in the #flares channel.", msg.Channel)
			return
		}

		client.Send("OK, let me get my flaregun", msg.Channel)

		// for now matches are indexed
		priority, _ := strconv.Atoi(params[0][1])
		topic := params[0][2]

		author, _ := msg.AuthorUser()
		assigneeUser, _ := JiraServer.GetUserByEmail(author.Profile.Email)
		ticket, _ := JiraServer.CreateTicket(priority, topic, assigneeUser)

		if ticket == nil {
			panic("no JIRA ticket created")
		}

		docTitle := fmt.Sprintf("%s: %s", ticket.Key, topic)
		doc, err := GoogleDocsServer.CreateFromTemplate(docTitle)

		if err != nil {
			panic("No google doc created")
		}

		channel, _ := client.CreateChannel(strings.ToLower(ticket.Key))

		// set up the Flare room
		client.Send(fmt.Sprintf("JIRA ticket: %s", ticket.Url()), channel.ID)
		client.Send(fmt.Sprintf("Facts docs: %s", doc.File.AlternateLink), channel.ID)
		client.Send(fmt.Sprintf("Flare resources: %s", resources_url), channel.ID)

		// announce the specific Flare room in the overall Flares room
		client.Send(fmt.Sprintf("@channel: Flare fired. Please visit #%s", strings.ToLower(ticket.Key)), msg.Channel)
	})

	client.Respond(takingLeadCommandRegexp, func(msg *Message, params [][]string) {
		ticket, err := GetTicketFromCurrentChannel(client, JiraServer, msg.Channel)

		if err != nil {
			client.Send("Sorry, I can only assign incident leads in a channel that corresponds to a Flare issue in JIRA.", msg.Channel)
			return
		}

		author, _ := msg.AuthorUser()
		assigneeUser, _ := JiraServer.GetUserByEmail(author.Profile.Email)

		client.Send("working on assigning incident lead....", msg.Channel)

		err = JiraServer.AssignTicketToUser(ticket, assigneeUser)

		client.Send(fmt.Sprintf("Oh Captain My Captain! @%s is now incident lead. Please confirm all actions with them.", author.Name), msg.Channel)
	})

	client.Respond(flareMitigatedCommandRegexp, func(msg *Message, params [][]string) {
		ticket, err := GetTicketFromCurrentChannel(client, JiraServer, msg.Channel)

		if err != nil {
			client.Send("Sorry, I can only assign incident leads in a channel that corresponds to a Flare issue in JIRA.", msg.Channel)
			return
		}

		client.Send("setting JIRA ticket to mitigated....", msg.Channel)

		err = JiraServer.DoTicketTransition(ticket, "Mitigate")

		if err == nil {
			client.Send("... and the Flare was mitigated, and there was much rejoicing throughout the land.", msg.Channel)
		} else {
			client.Send("... couldn't do it :( The JIRA ticket might not be in the right state. Check it: "+ticket.Url(), msg.Channel)
		}
	})

	panic(client.Run())
}
