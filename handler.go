package main

import "regexp"

type MessageHandler struct {
	pattern *regexp.Regexp
	fn      func(*Message, [][]string)
}

func (h *MessageHandler) Match(msg *Message) bool {
	return h.pattern.Match([]byte(msg.Text))
}

func (h *MessageHandler) Handle(msg *Message) {
	h.fn(msg, h.pattern.FindAllStringSubmatch(msg.Text, -1))
}
