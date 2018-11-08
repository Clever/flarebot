package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	// "github.com/Clever/flarebot/googledocs"
	"github.com/Clever/flarebot/jira"
	"github.com/Clever/flarebot/slack"
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
	regexp:      "[fF]ire (?:a )?(?:retroactive )?(?:.+emptive )?[fF]lare [pP]([012]) *(.*)",
	example:     "fire a flare p2 there is still no hottub on the roof",
	description: "Fire a new Flare with the given priority and description",
}

var testCommand = &command{
	regexp:      "test *(.*)",
	example:     "",
	description: "",
}

var takingLeadCommand = &command{
	regexp:      "[iI]('?m?| am?) (the )?incident lead",
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

// #flare-179-foo-bar --> #flare-179
var channelNameRegexp = regexp.MustCompile("^([^-]+-[^-]+)(?:-.+)")

func GetTicketFromCurrentChannel(client *slack.Client, JiraServer *jira.JiraServer, channelID string) (*jira.Ticket, error) {
	// first more info about the channel
	channel, err := client.API.GetChannelInfo(channelID)
	if err != nil {
		return nil, err
	}

	// we want to allow channel renaming as long as prefix remains #flare-<id>
	channelName := channelNameRegexp.ReplaceAllString(channel.Name, "$1")

	// then the ticket that matches
	ticket, err := JiraServer.GetTicketByKey(channelName)

	if err != nil || ticket.Fields.Project.ID != JiraServer.ProjectID {
		return nil, errors.New("no ticket for this channel")
	}

	return ticket, nil
}

func currentTimeStringInTZ(tz string) string {
	tzLocation, _ := time.LoadLocation(tz)
	return time.Now().In(tzLocation).Format(time.RFC3339)
}

func sendCommandsHelpMessage(client *slack.Client, channel string, commands []*command) {
	for _, c := range commands {
		client.Send(fmt.Sprintf("\"@%s: %s\" - %s", client.Username, c.example, c.description), channel)
	}
}

func sendHelpMessage(client *slack.Client, jiraServer *jira.JiraServer, channel string, inMainChannel bool) {
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

func sendReminderMessage(client *slack.Client, channel string, message string, delay time.Duration) {
	time.Sleep(delay)
	client.Send(message, channel)
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

	FlarebotUser, err := JiraServer.GetUserByEmail(JiraServer.Username)
	if err != nil {
		log.Fatalf("Failed to lookup flarebot user in jira: %s\n", err.Error())
	}

	// Google Docs service
	// googleDocsServer, err := googledocs.NewGoogleDocsServerWithServiceAccount(os.Getenv("GOOGLE_FLAREBOT_SERVICE_ACCOUNT_CONF"), os.Getenv("GOOGLE_TEMPLATE_DOC_ID"))
	// googleDomain := os.Getenv("GOOGLE_DOMAIN")

	// Link to flare resources
	resources_url := os.Getenv("FLARE_RESOURCES_URL")

	// Link to status page
	status_page_login_url := os.Getenv("STATUS_PAGE_LOGIN_URL")

	// Slack connection params
	var accessToken string
	if os.Getenv("SLACK_LEGACY_TOKEN") != "" {
		accessToken = os.Getenv("SLACK_FLAREBOT_ACCESS_TOKEN")
	} else {
		token := slack.DecodeOAuthToken(os.Getenv("SLACK_FLAREBOT_ACCESS_TOKEN"))
		accessToken = token.AccessToken
	}
	domain := os.Getenv("SLACK_DOMAIN")
	username := os.Getenv("SLACK_USERNAME")

	client, err := slack.NewClient(accessToken, domain, username)
	if err != nil {
		panic(err)
	}

	expectedChannel := os.Getenv("SLACK_CHANNEL")

	client.Respond(testCommand.regexp, func(msg *slack.Message, params [][]string) {
		author, err := msg.AuthorUser()
		if err != nil {
			client.Send("Unable to determine author of Slack message", msg.Channel)
			return
		}
		client.Send(fmt.Sprintf("I see you're using the test command. Excellent: %s", author.Profile.Email), msg.Channel)

		if len(params[0]) > 1 {
			client.Send(fmt.Sprintf("you told me: %s", params[0][1]), msg.Channel)
		}

		user, err := JiraServer.GetUserByEmail(author.Profile.Email)
		if err != nil {
			client.Send(fmt.Sprintf("Unable to determine JIRA user by email: %s", author.Profile.Email), msg.Channel)
			return
		}

		client.Send(fmt.Sprintf("JIRA username is %s", user.Name), msg.Channel)

		channel, err := client.API.GetChannelInfo(msg.Channel)
		if err != nil {
			client.Send("Unable to determine channel info", msg.Channel)
			return
		}

		client.Send(fmt.Sprintf("this channel is %s", channel.Name), msg.Channel)

		sampleTicketKey := "flare-165"
		ticket, err := JiraServer.GetTicketByKey(sampleTicketKey)
		if err != nil {
			client.Send(fmt.Sprintf("Unable to find JIRA ticket by key: %s", sampleTicketKey), msg.Channel)
			return
		}

		fmt.Println(ticket)

		// verify that we can open and parse the FLARE template
		// googleDocID := os.Getenv("GOOGLE_TEMPLATE_DOC_ID")
		// doc, err := googleDocsServer.GetDoc(googleDocID)
		// if err != nil {
		// 	client.Send(fmt.Sprintf("Unable to find the Google Doc Flare Template. ID: %s", googleDocID), msg.Channel)
		// 	return
		// }

		// _, err = googleDocsServer.GetDocContent(doc, "text/html")
		// if err != nil {
		// 	client.Send(fmt.Sprintf("Unable to get Google Doc Content for ID: %s", googleDocID), msg.Channel)
		// 	return
		// }
	})

	client.Respond(fireFlareCommand.regexp, func(msg *slack.Message, params [][]string) {
		// wrong channel?
		if msg.Channel != expectedChannel {
			return
		}

		client.API.SetUserAsActive()

		// retroactive?
		isRetroactive := strings.Contains(msg.Text, "retroactive")
		// preemptive?
		isPreemptive := strings.Contains(msg.Text, "emptive") // this could be pre-emptive, or preemptive

		if isRetroactive {
			client.Send("OK, let me quietly set up the Flare documents. Nobody freak out, this is retroactive.", msg.Channel)
		} else if isPreemptive {
			client.Send("OK, let me quietly set up the Flare documents. Nobody freak out, this is preemptive.", msg.Channel)
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
		// Kill Google Docs integration to fix flarebot
		// doc, err := googleDocsServer.CreateFromTemplate(docTitle, map[string]string{
		// 	"jira_key": ticket.Key,
		// })

		// if err != nil {
		// 	log.Fatalf("No google doc created: %s", err)
		// }

		// // update the google doc with some basic information
		// html, err := googleDocsServer.GetDocContent(doc, "text/html")

		// html = strings.Replace(html, "[FLARE-KEY]", ticket.Key, 1)
		// html = strings.Replace(html, "[START-DATE]", currentTimeStringInTZ("US/Pacific"), 1)
		// html = strings.Replace(html, "[SUMMARY]", topic, 1)

		// googleDocsServer.UpdateDocContent(doc, html)

		// // update permissions
		// if err = googleDocsServer.ShareDocWithDomain(doc, googleDomain, "writer"); err != nil {
		// 	log.Fatalf("Couldn't share google doc: %s", err)
		// }

		// // Add the doc to the Jira ticket
		// desc := fmt.Sprintf("[Facts Doc|%s]", doc.File.AlternateLink)
		// err = JiraServer.SetDescription(ticket, desc)
		// if err != nil {
		// 	fmt.Printf("Failed to set description for %s: %s\n", ticket.Key, err.Error())
		// }

		// set up the Flare room
		channel, err := client.CreateChannel(strings.ToLower(ticket.Key))

		if err != nil {
			log.Fatalf("Couldn't create Flare channel: %s", err)
		}

		if isRetroactive {
			client.Send("This is a RETROACTIVE Flare. All is well.", channel.ID)
		}

		client.API.SetChannelTopic(channel.ID, topic)

		client.Send(fmt.Sprintf("JIRA ticket: %s", ticket.Url()), channel.ID)
		// client.Send(fmt.Sprintf("Facts docs: %s", doc.File.AlternateLink), channel.ID)
		client.Send(fmt.Sprintf("Flare resources: %s", resources_url), channel.ID)
		client.Send(fmt.Sprintf("Manage status page: %s", status_page_login_url), channel.ID)
		client.Send(fmt.Sprintf("Remember: Rollback, Scale or Restart!"), channel.ID)

		// Pin the most important messages. NOTE: that this is based on text
		// matching, so the links need to be escaped to match
		client.Pin(fmt.Sprintf("JIRA ticket: <%s>", ticket.Url()), channel.ID)
		// client.Pin(fmt.Sprintf("Facts docs: <%s>", doc.File.AlternateLink), channel.ID)
		client.Pin(fmt.Sprintf("Manage status page: <%s>", status_page_login_url), channel.ID)
		client.Pin(fmt.Sprintf("Remember: Rollback, Scale or Restart!"), channel.ID)

		// send room-specific help
		sendHelpMessage(client, JiraServer, channel.ID, false)

		// let people know that they can rename this channel
		client.Send(fmt.Sprintf("NOTE: you can rename this channel as long as it starts with %s", channel.Name), channel.ID)

		go sendReminderMessage(client, channel.ID, "Are users affected? Consider creating an incident on the status page and updating the title.", 2*time.Minute)
		go sendReminderMessage(client, channel.ID, "Have you tried rolling back, scaling or restarting? (consider SSO version too)", 5*time.Minute)

		// announce the specific Flare room in the overall Flares room
		target := "channel"

		if isRetroactive || isPreemptive {
			author, _ := msg.AuthorUser()
			target = author.Name
		}

		client.Send(fmt.Sprintf("@%s: Flare fired. Please visit #%s -- %s", target, strings.ToLower(ticket.Key), topic), msg.Channel)
	})

	client.Respond(takingLeadCommand.regexp, func(msg *slack.Message, params [][]string) {
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

	client.Respond(flareMitigatedCommand.regexp, func(msg *slack.Message, params [][]string) {
		ticket, err := GetTicketFromCurrentChannel(client, JiraServer, msg.Channel)

		if err != nil {
			client.Send("Sorry, I can only assign incident leads in a channel that corresponds to a Flare issue in JIRA.", msg.Channel)
			return
		}

		client.Send("setting JIRA ticket to mitigated....", msg.Channel)

		// If the ticket is unassigned, attempt to assign it to the person
		// mitigating the flare. Since this is just for convenience, it
		// doesn't matter if it fails
		if ticket.Fields.Assignee.Name == FlarebotUser.Name {
			author, _ := msg.AuthorUser()
			assigneeUser, _ := JiraServer.GetUserByEmail(author.Profile.Email)
			JiraServer.AssignTicketToUser(ticket, assigneeUser)
		}

		if err := JiraServer.DoTicketTransition(ticket, "Mitigate"); err == nil {
			client.Send("... and the Flare was mitigated, and there was much rejoicing throughout the land.", msg.Channel)
		} else {
			client.Send("... couldn't do it :( The JIRA ticket might not be in the right state. Check it: "+ticket.Url(), msg.Channel)
		}

		// notify the main flares channel
		client.Send(fmt.Sprintf("#%s has been mitigated", strings.ToLower(ticket.Key)), expectedChannel)
	})

	client.Respond(notAFlareCommand.regexp, func(msg *slack.Message, params [][]string) {
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
		client.Send(fmt.Sprintf("turns out #%s is not a Flare", strings.ToLower(ticket.Key)), expectedChannel)
	})

	client.Respond(helpCommand.regexp, func(msg *slack.Message, params [][]string) {
		sendHelpMessage(client, JiraServer, msg.Channel, (msg.Channel == expectedChannel))
	})

	client.Respond(helpAllCommand.regexp, func(msg *slack.Message, param [][]string) {
		client.Send("Commands Available in the #flares channel:", msg.Channel)
		sendCommandsHelpMessage(client, msg.Channel, mainChannelCommands)
		client.Send("Commands Available in a single Flare channel:", msg.Channel)
		sendCommandsHelpMessage(client, msg.Channel, flareChannelCommands)
		client.Send("Commands Available in other channels:", msg.Channel)
		sendCommandsHelpMessage(client, msg.Channel, otherChannelCommands)
	})

	// fallback response saying "I don't understand"
	client.Respond(".*", func(msg *slack.Message, params [][]string) {
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
