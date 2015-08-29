package main

import (
	"os"
)

func main() {
	token := os.Getenv("SLACK_TOKEN")
	domain := os.Getenv("SLACK_DOMAIN")
	username := os.Getenv("SLACK_USERNAME")

	client, err := NewClient(token, domain, username)
	if err != nil {
		panic(err)
	}
	client.Respond("rack (me|one) up$", func(msg *Message, params [][]string) {
		author, err := msg.Author()
		if err != nil {
			panic(err)
		}
		redbull(author, "")
		msg.Respond("Thanks for telling me! Hope it gave you wings. :)")
	})
	client.Respond("rack (me|one) up for the (.*) team", func(msg *Message, params [][]string) {
		author, err := msg.Author()
		if err != nil {
			panic(err)
		}
		team := params[0][2]
		redbull(author, team)
		msg.Respond("Thanks for telling me! Hope it gave you wings. :)")
	})
	panic(client.Run())
}
