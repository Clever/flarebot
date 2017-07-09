package slack

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"log"
	"regexp"
	"sync"

	"golang.org/x/oauth2"

	slk "github.com/nlopes/slack"
)

type Client struct {
	rtm              *slk.RTM
	API              *slk.Client
	Username, userId string

	mHandler sync.RWMutex
	handlers []*MessageHandler

	wg   sync.WaitGroup
	mCnt sync.Mutex
	// don't overflow plz
	cnt int

	outgoing chan slk.OutgoingMessage
	forever  chan interface{}
}

func (c *Client) Run() error {
	// Run forever... eventually have error handling, probably
	c.wg.Wait()
	return nil
}

func (c *Client) Stop() {
	// This disconnects from slack and closes the rtm channels
	err := c.rtm.Disconnect()
	if err != nil {
		// Non fatal error
		fmt.Printf("Failed to disconnect from RTM: %s\n", err.Error())
	}

	// Only close channels that have been created
	if c.outgoing != nil {
		fmt.Printf("Closing outgoing channel...\n")
		close(c.outgoing)
	}
	c.wg.Wait()
}

func (c *Client) CreateChannel(name string) (*slk.Channel, error) {
	channel, err := c.API.CreateChannel(name)
	if err != nil {
		return nil, err
	} else {
		return channel, nil
	}
}

func (c *Client) Hear(pattern string, fn func(*Message, [][]string)) {
	h := &MessageHandler{pattern: regexp.MustCompile(pattern), fn: fn}
	c.mHandler.Lock()
	c.handlers = append(c.handlers, h)
	c.mHandler.Unlock()
}

func (c *Client) Respond(pattern string, fn func(*Message, [][]string)) {
	c.Hear(fmt.Sprintf("<@%s|%s>:? %s", c.Username, c.userId, pattern), fn)
}

func (c *Client) Send(msg, channelId string) {
	c.mCnt.Lock()
	id := c.cnt
	c.cnt += 1
	c.mCnt.Unlock()

	c.outgoing <- slk.OutgoingMessage{
		ID:      id,
		Channel: channelId,
		Text:    msg,
		Type:    "message",
	}
}

func (c *Client) Pin(msg, channelId string) {
	c.outgoing <- slk.OutgoingMessage{
		Channel: channelId,
		Text:    msg,
		Type:    "pin",
	}
}

func (c *Client) handleMessage(msg *slk.MessageEvent) {
	m := messageEventToMessage(msg, c.API, c.Send)

	var theMatch *MessageHandler
	fmt.Println()

	c.mHandler.RLock()
	for _, h := range c.handlers {
		if h.Match(m) {
			theMatch = h
			break
		}
	}
	c.mHandler.RUnlock()

	if theMatch != nil {
		theMatch.Handle(m)
	}
}

func (c *Client) pinSlackMessage(channelId, msg string) {
	// The channel is brand new, so there shouldn't be more than 100 messages in
	// it, which is the default count returned
	params := slk.NewHistoryParameters()
	history, err := c.API.GetChannelHistory(channelId, params)
	if err != nil {
		log.Fatalf("Failed to lookup channel history for %s: %s", channelId, err)
	}
	for _, post := range history.Messages {
		if post.Text == msg {
			// There should only be one message matching this text
			err := c.API.AddPin(channelId, slk.ItemRef{
				Channel:   post.Channel,
				Timestamp: post.Timestamp,
			})
			if err != nil {
				log.Fatalf("Failed to pin message: %s", err)
			}
			log.Printf("Pinned message `%s` in channel: %s\n", msg, channelId)
			return
		}
	}
	log.Fatalf("Could not find message with text `%s` in channel: %s", msg, channelId)
}

func (c *Client) start() {
	c.outgoing = make(chan slk.OutgoingMessage)

	// parameters for all postings
	messageParameters := slk.NewPostMessageParameters()
	messageParameters.LinkNames = 1
	messageParameters.AsUser = true

	c.wg.Add(1)
	go func(ws *slk.RTM, chSender chan slk.OutgoingMessage) {
		for msg := range chSender {
			switch msg.Type {
			case "message":
				_, _, err := ws.PostMessage(msg.Channel, msg.Text, messageParameters)
				if err != nil {
					log.Fatalf("Failed to post message. %s\n", err.Error())
				}
			case "pin":
				c.pinSlackMessage(msg.Channel, msg.Text)
			default:
				log.Fatalf("Unknown outgoing message type: %s", msg.Type)
			}
		}
		c.wg.Done()
		fmt.Printf("Outgoing goroutine ended\n")
	}(c.rtm, c.outgoing)

	c.wg.Add(1)
	go func(rtm *slk.RTM) {
		// Not sure why, but MannageConnection() needs to be run as a goroutine
		// Otherwise it blocks forever
		go rtm.ManageConnection()
		for msg := range rtm.IncomingEvents {
			switch msg.Data.(type) {
			case *slk.HelloEvent:
				fmt.Println("Hello!")
			case *slk.MessageEvent:
				c.handleMessage(msg.Data.(*slk.MessageEvent))
			case *slk.RTMError:
				error := msg.Data.(*slk.RTMError)
				fmt.Printf("Error: %d - %s\n", error.Code, error.Msg)
			case *slk.UserTypingEvent, *slk.PresenceChangeEvent, slk.LatencyReport:
				// Do nothing
				continue
			case *slk.ConnectionErrorEvent:
				error := msg.Data.(*slk.ConnectionErrorEvent)
				fmt.Printf("Error: %v\n", error)
			case *slk.ChannelCreatedEvent:
				data := msg.Data.(*slk.ChannelCreatedEvent)
				fmt.Printf("Created channel: %s (%s)\n", data.Channel.Name, data.Channel.ID)
			case *slk.ChannelJoinedEvent:
				data := msg.Data.(*slk.ChannelJoinedEvent)
				fmt.Printf("Joined channel: %s (%s)\n", data.Channel.Name, data.Channel.ID)
			case *slk.ConnectedEvent:
				fmt.Printf("Connected to slack!\n")
			case *slk.ConnectingEvent:
				// Ignore the Connecting events
			case *slk.LatencyReport:
				// Ignore the Latency reports
			case *slk.ReconnectUrlEvent:
				// Ignore the reconnect URLS
			case *slk.DisconnectedEvent:
				// If the disconnect was intentional, exit the goroutine
				data := msg.Data.(*slk.DisconnectedEvent)
				if data.Intentional {
					fmt.Printf("Disconnected from slack (intentionally)\n")
					fmt.Printf("Incoming goroutine ended\n")
					c.wg.Done()
					return
				}
			default:
				fmt.Printf("Unexpected: %#v\n", msg.Data)
			}
		}
		c.wg.Done()
		fmt.Printf("Incoming goroutine ended\n")
	}(c.rtm)
}

func NewClient(token, domain, username string) (*Client, error) {
	api := slk.New(token)

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

	rtm := api.NewRTM()
	client := &Client{API: api, rtm: rtm, Username: username, userId: userId}
	client.start()
	return client, nil
}

func DecodeOAuthToken(tokenString string) *oauth2.Token {
	tokenBytes, _ := base64.StdEncoding.DecodeString(tokenString)
	tokenBytesBuffer := bytes.NewBuffer(tokenBytes)
	dec := gob.NewDecoder(tokenBytesBuffer)
	token := new(oauth2.Token)
	dec.Decode(token)

	return token
}
