package main

import (
	"fmt"
	"log"
	"os"
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

	var rootCmd = &cobra.Command{Use: "jira-cli"}
	rootCmd.AddCommand(cmdGetUserByEmail)
	rootCmd.AddCommand(cmdGetTicket)
	rootCmd.Execute()
}
