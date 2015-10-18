package main

import (
	"fmt"
	"regexp"
	"sync"
	//	"time"

	"github.com/nlopes/slack"
)

type Client struct {
	rtm              *slack.RTM
	api              *slack.Client
	username, userId string

	mHandler sync.RWMutex
	handlers []*MessageHandler

	mCnt sync.Mutex
	// don't overflow plz
	cnt int

	outgoing chan slack.OutgoingMessage
}

func (c *Client) Run() error {
	// Run forever... eventually have error handling, probably
	f := make(chan struct{})
	<-f
	return nil
}

func (c *Client) Hear(pattern string, fn func(*Message, [][]string)) {
	h := &MessageHandler{pattern: regexp.MustCompile(pattern), fn: fn}
	c.mHandler.Lock()
	c.handlers = append(c.handlers, h)
	c.mHandler.Unlock()
}

func (c *Client) Respond(pattern string, fn func(*Message, [][]string)) {
	c.Hear(fmt.Sprintf("<@%s|%s>:? %s", c.username, c.userId, pattern), fn)
}

func (c *Client) Send(msg, channelId string) {
	c.mCnt.Lock()
	id := c.cnt
	c.cnt += 1
	c.mCnt.Unlock()

	c.outgoing <- slack.OutgoingMessage{
		ID:      id,
		Channel: channelId,
		Text:    msg,
		Type:    "message",
	}
}

func (c *Client) handleMessage(msg *slack.MessageEvent) {
	m := messageEventToMessage(msg, c.api, c.Send)

	matches := []*MessageHandler{}
	fmt.Println()

	c.mHandler.RLock()
	for _, h := range c.handlers {
		if h.Match(m) {
			matches = append(matches, h)
		}
	}
	c.mHandler.RUnlock()

	for _, h := range matches {
		h.Handle(m)
	}
}

func (c *Client) start() {
	c.outgoing = make(chan slack.OutgoingMessage)

	// parameters for all postings
	messageParameters := slack.NewPostMessageParameters()
	messageParameters.LinkNames = 1
	messageParameters.AsUser = true

	go func(ws *slack.RTM, chSender chan slack.OutgoingMessage) {
		for {
			select {
			case msg := <-chSender:
				ws.PostMessage(msg.Channel, msg.Text, messageParameters)
			}
		}
	}(c.rtm, c.outgoing)
	for {
		select {
		//		case msg := <-chReceiver:
		case msg := <-c.rtm.IncomingEvents:
			switch msg.Data.(type) {
			case slack.HelloEvent:
				fmt.Println("Hello!")
			case *slack.MessageEvent:
				c.handleMessage(msg.Data.(*slack.MessageEvent))
			case *slack.RTMError:
				error := msg.Data.(*slack.RTMError)
				fmt.Printf("Error: %d - %s\n", error.Code, error.Msg)
			case *slack.UserTypingEvent, *slack.PresenceChangeEvent, slack.LatencyReport:
				// Do nothing
				continue
			case *slack.ConnectionErrorEvent:
				error := msg.Data.(*slack.ConnectionErrorEvent)
				fmt.Printf("Error: %v\n", error)
			default:
				fmt.Printf("Unexpected: %#v\n", msg.Data)
			}
		}
	}

}

func NewClient(token, domain, username string) (*Client, error) {
	api := slack.New(token)

	users, err := api.GetUsers()
	if err != nil {
		return nil, err
	}
	var userId string
	for _, user := range users {
		if user.Name == username {
			userId = user.ID
			break
		}
	}

	/*
		rtm, err := api.StartRTM("", domain)
		if err != nil {
			return nil, err
		}*/
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	client := &Client{api: api, rtm: rtm, username: username, userId: userId}
	go client.start()
	return client, nil
}
