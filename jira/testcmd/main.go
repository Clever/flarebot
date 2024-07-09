package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Clever/flarebot/jira"
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
)

type jiraCommand struct {
	jiraServer jira.JiraServer
}

func (jc *jiraCommand) GetTicketCommand(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		log.Fatalf("A single jira ticket name must be provided\n")
	}

	ticket, err := jc.jiraServer.GetTicketByKey(args[0])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Ticket details:\n")
	spew.Dump(ticket)
}

func (jc *jiraCommand) GetUserByEmail(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		log.Fatalf("A single email address must be provided\n")
	}

	user, err := jc.jiraServer.GetUserByEmail(args[0])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("User object:\n")
	spew.Dump(user)
}

func (jc *jiraCommand) SetDescriptionCommand(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		log.Fatalf("A jira ticket and description must be provided\n")
	}

	ticket, err := jc.jiraServer.GetTicketByKey(args[0])
	if err != nil {
		log.Fatal(err)
	}

	err = jc.jiraServer.SetDescription(ticket, args[1])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Set description to '%s'\n", args[1])
}

func (jc *jiraCommand) CreateTicketCommand(cmd *cobra.Command, args []string) {
	if len(args) != 3 {
		log.Fatalf("A priority, description and assignee must be provided\n")
	}

	assignee, err := jc.jiraServer.GetUserByEmail(args[2])
	fmt.Printf("User object:\n")
	spew.Dump(assignee)

	if err != nil {
		log.Fatal(err)
	}

	priority, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatal(err)
	}
	ticket, err := jc.jiraServer.CreateTicket(priority, args[1], assignee)

	fmt.Printf("Ticket created:\n")
	spew.Dump(ticket)
}

func main() {
	jc := jiraCommand{
		jiraServer: jira.JiraServer{
			Origin:      os.Getenv("JIRA_ORIGIN"),
			Username:    os.Getenv("JIRA_USERNAME"),
			Password:    os.Getenv("JIRA_PASSWORD"),
			ProjectID:   os.Getenv("JIRA_PROJECT_ID"),
			IssueTypeID: os.Getenv("JIRA_ISSUETYPE_ID"),
			PriorityIDs: strings.Split(os.Getenv("JIRA_PRIORITIES"), ","),
		},
	}

	var cmdGetUserByEmail = &cobra.Command{
		Use:   "getUserByEmail <email address>",
		Short: "print the user record for a jira user",
		Run:   jc.GetUserByEmail,
	}

	var cmdGetTicket = &cobra.Command{
		Use:   "getTicket <ticket-id>",
		Short: "print the captured jira ticket attributes",
		Run:   jc.GetTicketCommand,
	}

	var cmdSetDescription = &cobra.Command{
		Use:   "setDescription <ticket-id> <description>",
		Short: "replace the description with the text",
		Run:   jc.SetDescriptionCommand,
	}

	var cmdCreateTicket = &cobra.Command{
		Use:   "createTicket <priority> <topic> <assignee-email>",
		Short: "creates ticket with given assignee",
		Run:   jc.CreateTicketCommand,
	}

	var rootCmd = &cobra.Command{Use: "jira-cli"}
	rootCmd.AddCommand(cmdGetUserByEmail)
	rootCmd.AddCommand(cmdGetTicket)
	rootCmd.AddCommand(cmdSetDescription)
	rootCmd.AddCommand(cmdCreateTicket)
	rootCmd.Execute()
}
