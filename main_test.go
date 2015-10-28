package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJiraTicketRegex(t *testing.T) {
	for _, testCase := range []struct {
		intput, expected string
	}{
		{"flare-1", "flare-001"},
		{"flare-11", "flare-011"},
		{"flare-123", "flare-123"},
		{"flare-abc", "flare-abc"},             // bad numbers should be unchanged
		{"what-is-a-flare", "what-is-a-flare"}, // names not matching regex should be ignored
	} {
		newName := reformatSlackChannelName(testCase.intput)
		assert.Equal(t, testCase.expected, newName)
	}
}
