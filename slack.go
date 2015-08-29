package main

import (
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/nlopes/slack"
)

type Client struct {
	rtm              *slack.SlackWS
	api              *slack.Slack
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
		Id:        id,
		ChannelId: channelId,
		Text:      msg,
		Type:      "message",
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
	chReceiver := make(chan slack.SlackEvent)

	go c.rtm.HandleIncomingEvents(chReceiver)
	go c.rtm.Keepalive(20 * time.Second)
	go func(ws *slack.SlackWS, chSender chan slack.OutgoingMessage) {
		for {
			select {
			case msg := <-chSender:
				ws.SendMessage(&msg)
			}
		}
	}(c.rtm, c.outgoing)
	for {
		select {
		case msg := <-chReceiver:
			switch msg.Data.(type) {
			case slack.HelloEvent:
				fmt.Println("Hello!")
			case *slack.MessageEvent:
				c.handleMessage(msg.Data.(*slack.MessageEvent))
			case *slack.SlackWSError:
				error := msg.Data.(*slack.SlackWSError)
				fmt.Printf("Error: %d - %s\n", error.Code, error.Msg)
			case *slack.UserTypingEvent, *slack.PresenceChangeEvent, slack.LatencyReport:
				// Do nothing
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
			userId = user.Id
			break
		}
	}

	rtm, err := api.StartRTM("", domain)
	if err != nil {
		return nil, err
	}

	client := &Client{api: api, rtm: rtm, username: username, userId: userId}
	go client.start()
	return client, nil
}
