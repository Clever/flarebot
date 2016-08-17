package main

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"html"
	"log"
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

type command struct {
	regexp      string
	example     string
	description string
}

var fireFlareCommand = &command{
	regexp:      "[fF]ire (?:a )?(?:retroactive )?[fF]lare [pP]([012]) *(.*)",
	example:     "fire a flare p2 there is still no hottub on the roof",
	description: "Fire a new Flare with the given priority and description",
}

var testCommand = &command{
	regexp:      "test *(.*)",
	example:     "",
	description: "",
}

var takingLeadCommand = &command{
	regexp:      "[iI] am (the )?incident lead",
	example:     "I am incident lead",
	description: "Declare yourself incident-lead",
}

var flareMitigatedCommand = &command{
	regexp:      "([Ff]lare )?(is )?mitigated",
	example:     "flare mitigated",
	description: "Mark the Flare mitigated",
}

// not a flare
var notAFlareCommand = &command{
	regexp:      "([Ff]lare )?(is )?not a [Ff]lare",
	example:     "not a flare",
	description: "Mark the Flare not-a-flare",
}

// help command
var helpCommand = &command{
	regexp:      "[Hh]elp *$",
	example:     "help",
	description: "display the list of commands available in this channel",
}

// help all command
var helpAllCommand = &command{
	regexp:      "[Hh]elp [Aa]ll",
	example:     "help all",
	description: "display the list of all commands and the channels where they're available",
}

var mainChannelCommands = []*command{helpCommand, helpAllCommand, fireFlareCommand}
var flareChannelCommands = []*command{helpCommand, takingLeadCommand, flareMitigatedCommand, notAFlareCommand}
var otherChannelCommands = []*command{helpAllCommand}

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

func sendCommandsHelpMessage(client *Client, channel string, commands []*command) {
	for _, c := range commands {
		client.Send(fmt.Sprintf("\"@%s: %s\" - %s", client.username, c.example, c.description), channel)
	}
}

func sendHelpMessage(client *Client, jiraServer *jira.JiraServer, channel string, inMainChannel bool) {
	var availableCommands []*command

	if inMainChannel {
		availableCommands = mainChannelCommands
	} else {
		_, err := GetTicketFromCurrentChannel(client, jiraServer, channel)

		if err == nil {
			availableCommands = flareChannelCommands
		} else {
			availableCommands = otherChannelCommands
		}
	}

	if len(availableCommands) == 0 {
		client.Send("no available commands in this channel.", channel)
	} else {
		client.Send("Available commands:", channel)
		sendCommandsHelpMessage(client, channel, availableCommands)
	}
}

func timeTillNextTopicChange(now time.Time) time.Duration {
	pt, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Fatal("couldn't load timezone for America/Los_Angeles: ", err)
	}
	now = now.In(pt)
	// Sunday = 0, Monday = 1, etc. so next Monday is 8
	// to get the difference between next Monday and today, subtract from 8 and mod 7
	// mod 7 ensures that if it's Sunday, then the next Monday is 1 instead of 8
	// if today == Monday (days == 0), then look at hours
	days := (8 - now.Weekday()) % 7
	if days == 0 {
		// special case for Monday
		// if it's not past noon, we can still change the topic today at noon,
		// so we can leave days == 0
		if now.Hour() > 11 {
			// it's past noon, so we need the next Monday
			days = 7
		}
	}
	year, month, day := now.Date()
	t := time.Date(year, month, day, 0, 0, 0, 0, pt) // see https://github.com/golang/go/issues/10894
	t = t.Add(24 * time.Duration(days) * time.Hour)  // next Monday 00:00:00
	t = t.Add(12 * time.Hour)                        // next Monday noon
	return t.Sub(now)
}

func swapNextTeam(topic string) string {
	teams := []string{
		"#oncall-apps",
		"#oncall-classrooms",
		"#oncall-districts",
		"#oncall-infra",
		"#oncall-ip",
	}
	for i, team := range teams {
		if strings.Contains(topic, team) {
			topic = strings.Replace(topic, team, teams[(i+1)%len(teams)], -1)
			break
		}
	}
	return topic
}

func changeTopic(client *Client, channel string) {
	for {
		t := timeTillNextTopicChange(time.Now())
		time.Sleep(t)
		info, err := client.api.GetChannelInfo(channel)
		if err != nil {
			log.Fatal("error getting topic: ", err)
		}
		topic := html.UnescapeString(info.Topic.Value)
		topic = swapNextTeam(topic)
		_, err = client.api.SetChannelTopic(channel, topic)
		if err != nil {
			log.Fatal("error setting topic: ", err)
		}
	}
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

	go changeTopic(client, expectedChannel)

	client.Respond(testCommand.regexp, func(msg *Message, params [][]string) {
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

	client.Respond(fireFlareCommand.regexp, func(msg *Message, params [][]string) {
		// wrong channel?
		if msg.Channel != expectedChannel {
			return
		}

		client.api.SetUserAsActive()

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

		client.Send(topic, channel.ID)

		client.Send(fmt.Sprintf("JIRA ticket: %s", ticket.Url()), channel.ID)
		client.Send(fmt.Sprintf("Facts docs: %s", doc.File.AlternateLink), channel.ID)
		client.Send(fmt.Sprintf("Flare resources: %s", resources_url), channel.ID)

		// send room-specific help
		sendHelpMessage(client, JiraServer, channel.ID, false)

		// announce the specific Flare room in the overall Flares room
		target := "channel"

		if isRetroactive {
			author, _ := msg.AuthorUser()
			target = author.Name
		}

		client.Send(fmt.Sprintf("@%s: Flare fired. Please visit #%s", target, strings.ToLower(ticket.Key)), msg.Channel)
	})

	client.Respond(takingLeadCommand.regexp, func(msg *Message, params [][]string) {
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

	client.Respond(flareMitigatedCommand.regexp, func(msg *Message, params [][]string) {
		ticket, err := GetTicketFromCurrentChannel(client, JiraServer, msg.Channel)

		if err != nil {
			client.Send("Sorry, I can only assign incident leads in a channel that corresponds to a Flare issue in JIRA.", msg.Channel)
			return
		}

		client.Send("setting JIRA ticket to mitigated....", msg.Channel)

		if err := JiraServer.DoTicketTransition(ticket, "Mitigate"); err == nil {
			client.Send("... and the Flare was mitigated, and there was much rejoicing throughout the land.", msg.Channel)
		} else {
			client.Send("... couldn't do it :( The JIRA ticket might not be in the right state. Check it: "+ticket.Url(), msg.Channel)
		}

		// notify the main flares channel
		client.Send(fmt.Sprintf("@channel: #%s has been mitigated", strings.ToLower(ticket.Key)), expectedChannel)
	})

	client.Respond(notAFlareCommand.regexp, func(msg *Message, params [][]string) {
		ticket, err := GetTicketFromCurrentChannel(client, JiraServer, msg.Channel)

		if err != nil {
			client.Send("Sorry, I can't find the JIRA.", msg.Channel)
			return
		}

		client.Send("setting JIRA ticket to Not a Flare....", msg.Channel)

		if err := JiraServer.DoTicketTransition(ticket, "Not A Flare"); err == nil {
			client.Send("... and done.", msg.Channel)
		} else {
			client.Send("... couldn't do it :( The JIRA ticket might not be in the right state. Check it: "+ticket.Url(), msg.Channel)
		}

		// notify the main flares channel
		client.Send(fmt.Sprintf("@channel: turns out #%s is not a Flare", strings.ToLower(ticket.Key)), expectedChannel)
	})

	client.Respond(helpCommand.regexp, func(msg *Message, params [][]string) {
		sendHelpMessage(client, JiraServer, msg.Channel, (msg.Channel == expectedChannel))
	})

	client.Respond(helpAllCommand.regexp, func(msg *Message, param [][]string) {
		client.Send("Commands Available in the #flares channel:", msg.Channel)
		sendCommandsHelpMessage(client, msg.Channel, mainChannelCommands)
		client.Send("Commands Available in a single Flare channel:", msg.Channel)
		sendCommandsHelpMessage(client, msg.Channel, flareChannelCommands)
		client.Send("Commands Available in other channels:", msg.Channel)
		sendCommandsHelpMessage(client, msg.Channel, otherChannelCommands)
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
