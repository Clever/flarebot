package main

import (
	"fmt"

	"github.com/Clever/flarebot/Godeps/_workspace/src/github.com/nlopes/slack"
)

type Message struct {
	AuthorId   string
	AuthorName string
	Timestamp  string
	Text       string
	Channel    string
	api        *slack.Client
	sender     func(string, string)
}

func (m *Message) Author() (string, error) {
	user, err := m.api.GetUserInfo(m.AuthorId)
	if err != nil {
		return "", err
	}
	return user.Name, nil
}

func (m *Message) Respond(msg string) {
	m.sender(fmt.Sprintf("@%s: %s", m.AuthorName, msg), m.Channel)
}

func messageEventToMessage(msg *slack.MessageEvent, api *slack.Client, sender func(string, string)) *Message {
	return &Message{
		AuthorId:   msg.User,
		AuthorName: msg.Username,
		Timestamp:  msg.Timestamp,
		Text:       msg.Text,
		Channel:    msg.Channel,
		api:        api,
		sender:     sender,
	}
}
