package main

import (
	"fmt"

	"github.com/nlopes/slack"
)

type Message struct {
	AuthorId  string
	Timestamp string
	Text      string
	ChannelId string
	api       *slack.Slack
	sender    func(string, string)
}

func (m *Message) Author() (string, error) {
	user, err := m.api.GetUserInfo(m.AuthorId)
	if err != nil {
		return "", err
	}
	return user.Name, nil
}

func (m *Message) Respond(msg string) {
	m.sender(fmt.Sprintf("<@%s> %s", m.AuthorId, msg), m.ChannelId)
}

func messageEventToMessage(msg *slack.MessageEvent, api *slack.Slack, sender func(string, string)) *Message {
	return &Message{
		AuthorId:  msg.UserId,
		Timestamp: msg.Timestamp,
		Text:      msg.Text,
		ChannelId: msg.ChannelId,
		api:       api,
		sender:    sender,
	}
}
