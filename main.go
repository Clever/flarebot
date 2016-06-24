package main

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/Clever/flarebot/googledocs"
	"github.com/Clever/flarebot/jira"
)

//
// COMMANDS
//

// fire a flare
const fireFlareCommandRegexp string = "[fF]ire (?:a )?(?:retroactive )?[fF]lare [pP]([012]) *(.*)"

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

func currentTimeStringInTZ(tz string) string {
	tzLocation, _ := time.LoadLocation(tz)
	return time.Now().In(tzLocation).Format(time.RFC3339)
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

		// get a test google doc and update it
		googleDocID := "1Hd2T9hr4wYQZY6ZoZJG3y3yc6zuABjLumPQccHI1XXw"
		doc, _ := GoogleDocsServer.GetDoc(googleDocID)
		html, _ := GoogleDocsServer.GetDocContent(doc, "text/html")
		newHTML := strings.Replace(html, "Flare", "Booya", 1)
		GoogleDocsServer.UpdateDocContent(doc, newHTML)
	})

	client.Respond(fireFlareCommandRegexp, func(msg *Message, params [][]string) {
		// wrong channel?
		if msg.Channel != expectedChannel {
			return
		}

		// retroactive?
		isRetroactive := strings.Contains(msg.Text, "retroactive")

		if isRetroactive {
			client.Send("OK, let me quietly set up the Flare documents. Nobody freak out, this is retroactive.", msg.Channel)
		} else {
			client.Send("OK, let me get my flaregun", msg.Channel)
		}

		// for now matches are indexed
		priority, _ := strconv.Atoi(params[0][1])
		topic := params[0][2]

		author, _ := msg.AuthorUser()
		assigneeUser, _ := JiraServer.GetUserByEmail(author.Profile.Email)
		ticket, _ := JiraServer.CreateTicket(priority, topic, assigneeUser)

		if ticket == nil {
			panic("no JIRA ticket created")
		}

		// start progress on the ticket
		err = JiraServer.DoTicketTransition(ticket, "Start Progress")

		if err != nil {
			client.Send("JIRA ticket created, but couldn't mark it 'started'.", msg.Channel)
		}

		// retroactive Flares are immediately mitigated
		if isRetroactive {
			err = JiraServer.DoTicketTransition(ticket, "Mitigate")
		}

		docTitle := fmt.Sprintf("%s: %s", ticket.Key, topic)

		if isRetroactive {
			docTitle = fmt.Sprintf("%s - Retroactive", docTitle)
		}

		doc, err := GoogleDocsServer.CreateFromTemplate(docTitle, map[string]string{
			"jira_key": ticket.Key,
		})

		if err != nil {
			panic("No google doc created")
		}

		// update the google doc with some basic information
		html, err := GoogleDocsServer.GetDocContent(doc, "text/html")

		html = strings.Replace(html, "[FLARE-KEY]", ticket.Key, 1)
		html = strings.Replace(html, "[START-DATE]", currentTimeStringInTZ("US/Pacific"), 1)
		html = strings.Replace(html, "[SUMMARY]", topic, 1)

		GoogleDocsServer.UpdateDocContent(doc, html)

		err = GoogleDocsServer.SetDocPermissionTypeRole(doc, "domain", "writer")

		channel, _ := client.CreateChannel(strings.ToLower(ticket.Key))

		// set up the Flare room
		if isRetroactive {
			client.Send("This is a RETROACTIVE Flare. All is well.", channel.ID)
		}

		client.Send(fmt.Sprintf("JIRA ticket: %s", ticket.Url()), channel.ID)
		client.Send(fmt.Sprintf("Facts docs: %s", doc.File.AlternateLink), channel.ID)
		client.Send(fmt.Sprintf("Flare resources: %s", resources_url), channel.ID)

		// announce the specific Flare room in the overall Flares room
		target := "channel"

		if isRetroactive {
			author, _ := msg.AuthorUser()
			target = author.Name
		}

		client.Send(fmt.Sprintf("@%s: Flare fired. Please visit #%s", target, strings.ToLower(ticket.Key)), msg.Channel)
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

	// fallback response saying "I don't understand"
	client.Respond(".*", func(msg *Message, params [][]string) {
		// if not in the main Flares channel
		if msg.Channel != expectedChannel {
			_, err := GetTicketFromCurrentChannel(client, JiraServer, msg.Channel)

			// or in a flare-specific channel
			if err != nil {
				// bail
				return
			}
		}

		// should be taking commands here, and didn't understand
		client.Send("I'm sorry, I didn't understand that command.", msg.Channel)
	})

	panic(client.Run())
}
