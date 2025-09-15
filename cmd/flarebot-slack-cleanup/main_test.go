package main

import (
	"context"
	"errors"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	slk "github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

//go:generate mockgen -package main -destination mock.go -source main.go SlackClient,JiraClient

type handleInput struct {
	ctx        context.Context
	testConfig TestConfig
}

type TestConfig struct {
	DryRun bool `json:"dryRun"`
}

type handleOutput struct {
	err error
}

type handleTest struct {
	description      string
	input            handleInput
	output           handleOutput
	mockExpectations func(slackClient *MockSlackClient, jiraClient *MockJiraClient)
}

func TestHandle(t *testing.T) {
	channel := slk.Channel{
		GroupConversation: slk.GroupConversation{
			Name: "flaretest-old-channel",
		},
	}
	channel.Conversation.Created = slk.JSONTime(1234567890)

	tests := []handleTest{
		{
			description: "test channels archived > no dry run",
			input: handleInput{
				ctx:        context.Background(),
				testConfig: TestConfig{DryRun: false},
			},
			output: handleOutput{
				err: nil,
			},
			mockExpectations: func(slackClient *MockSlackClient, jiraClient *MockJiraClient) {
				slackClient.EXPECT().GetConversations(gomock.Any()).Return([]slk.Channel{channel}, "", nil).Times(1)
				slackClient.EXPECT().ArchiveConversation(channel.ID).Return(nil).Times(1)
				jiraClient.EXPECT().GetTicketByKey("FLARETEST-OLD-CHANNEL").Return(nil, nil).Times(1)
				jiraClient.EXPECT().SetLabel(gomock.Any(), "archived").Return(nil).Times(1)
			},
		},
		{
			description: "test channels archived > dry run",
			input: handleInput{
				ctx:        context.Background(),
				testConfig: TestConfig{DryRun: true},
			},
			output: handleOutput{
				err: nil,
			},
			mockExpectations: func(slackClient *MockSlackClient, jiraClient *MockJiraClient) {
				slackClient.EXPECT().GetConversations(gomock.Any()).Return([]slk.Channel{channel}, "", nil).Times(1)
				slackClient.EXPECT().ArchiveConversation(channel.ID).Return(nil).Times(0)
			},
		},
		{
			description: "test channels archived > flare bot not in channel",
			input: handleInput{
				ctx:        context.Background(),
				testConfig: TestConfig{DryRun: false},
			},
			output: handleOutput{
				err: nil,
			},
			mockExpectations: func(slackClient *MockSlackClient, jiraClient *MockJiraClient) {
				slackClient.EXPECT().GetConversations(gomock.Any()).Return([]slk.Channel{channel}, "", nil).Times(1)
				slackClient.EXPECT().ArchiveConversation(channel.ID).Return(errors.New("not_in_channel")).Times(1)
				slackClient.EXPECT().JoinConversation(channel.ID).Return(&channel, "", []string{}, nil).Times(1)
				slackClient.EXPECT().ArchiveConversation(channel.ID).Return(nil).Times(1)
				jiraClient.EXPECT().GetTicketByKey("FLARETEST-OLD-CHANNEL").Return(nil, nil).Times(1)
				jiraClient.EXPECT().SetLabel(gomock.Any(), "archived").Return(nil).Times(1)
			},
		},
		{
			description: "test channels archived > rate limited",
			input: handleInput{
				ctx:        context.Background(),
				testConfig: TestConfig{DryRun: false},
			},
			output: handleOutput{
				err: nil,
			},
			mockExpectations: func(slackClient *MockSlackClient, jiraClient *MockJiraClient) {
				slackClient.EXPECT().GetConversations(gomock.Any()).Return([]slk.Channel{channel}, "", nil).Times(1)
				rlErr := &slk.RateLimitedError{
					RetryAfter: 1 * time.Second,
				}
				slackClient.EXPECT().ArchiveConversation(channel.ID).Return(rlErr).Times(1)
				slackClient.EXPECT().ArchiveConversation(channel.ID).Return(nil).Times(1)
				jiraClient.EXPECT().GetTicketByKey("FLARETEST-OLD-CHANNEL").Return(nil, nil).Times(1)
				jiraClient.EXPECT().SetLabel(gomock.Any(), "archived").Return(nil).Times(1)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			mockController := gomock.NewController(t)
			defer mockController.Finish()
			mockSlackClient := NewMockSlackClient(mockController)
			mockJiraClient := NewMockJiraClient(mockController)
			test.mockExpectations(mockSlackClient, mockJiraClient)
			os.Setenv("DRY_RUN", strconv.FormatBool(test.input.testConfig.DryRun))
			err := Handler{slackClient: mockSlackClient, jiraClient: mockJiraClient}.Handle(test.input.ctx)
			assert.Equal(t, test.output.err, err)
		})
	}
}
