package slack

import (
	"fmt"

	slk "github.com/nlopes/slack"
)

type Message struct {
	AuthorId   string
	AuthorName string
	Timestamp  string
	Text       string
	Channel    string
	api        *slk.Client
	sender     func(string, string)
}

func (m *Message) Author() (string, error) {
	user, err := m.AuthorUser()
	if err != nil {
		return "", err
	}
	return user.Name, nil
}

func (m *Message) AuthorUser() (*slk.User, error) {
	user, err := m.api.GetUserInfo(m.AuthorId)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (m *Message) Respond(msg string) {
	m.sender(fmt.Sprintf("@%s: %s", m.AuthorName, msg), m.Channel)
}

func messageEventToMessage(msg *slk.MessageEvent, api *slk.Client, sender func(string, string)) *Message {
	m := &Message{
		AuthorId:   msg.User,
		AuthorName: msg.Username,
		Timestamp:  msg.Timestamp,
		Text:       msg.Text,
		Channel:    msg.Channel,
		api:        api,
		sender:     sender,
	}
	fmt.Printf("Message: m=%+v\n\n", m)
	return m
}
