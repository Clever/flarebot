package jira

import (
	//	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

//
// <TEST_DATA>
//

var mockOrigin = "https://mock-jira.com"
var mockUsername = "mockUsername"
var mockPassword = "mockPassword"
var mockProjectID = "mockProjectID"
var mockIssueTypeID = "1"
var mockPriorityIDs = []string{"mockPriority0", "mockPriority1", "mockPriority2"}

var mockIssueID = "MOCK-ISSUE-ID"

var mockIssueContent = `{"expand":"renderedFields,names,schema,operations,editmeta,changelog,versionedRepresentations","id":"28902","self":"https://mock.atlassian.net/rest/api/2/issue/28902","key":"MOCK-ISSUE-ID","fields":{"issuetype":{"self":"https://mock.atlassian.net/rest/api/2/issuetype/1","id":"1","description":"A problem which impairs or prevents the functions of the product.","iconUrl":"https://mock.atlassian.net/secure/viewavatar?size=xsmall&avatarId=10503&avatarType=issuetype","name":"Bug","subtask":false,"avatarId":10503},"timespent":null,"timeoriginalestimate":null,"description":"we broke verything","project":{"self":"https://mock.atlassian.net/rest/api/2/project/11701","id":"11701","key":"FLARE","name":"Flares"},"customfield_11100":null,"customfield_11200":"com.atlassian.servicedesk.plugins.approvals.internal.customfield.ApprovalsCFValue@ad0295","aggregatetimespent":null,"resolution":null,"customfield_11300":null,"timetracking":{},"customfield_10006":"9223372036854775807","customfield_10007":null,"customfield_10008":null,"attachment":[],"aggregatetimeestimate":null,"customfield_10605":null,"resolutiondate":null,"customfield_10606":null,"workratio":-1,"summary":"the world is collapsing","lastViewed":null,"watches":{"self":"https://mock.atlassian.net/rest/api/2/issue/MOCK-ISSUE-ID/watchers","watchCount":1,"isWatching":false},"creator":{"self":"https://mock.atlassian.net/rest/api/2/user?username=flarebot","name":"flarebot","key":"flarebot","emailAddress":"flarebot@example.com","displayName":"Flare Bot","active":true,"timeZone":"America/Los_Angeles"},"subtasks":[],"created":"2016-06-15T08:21:13.591-0700","reporter":{"self":"https://mock.atlassian.net/rest/api/2/user?username=flarebot","name":"flarebot","key":"flarebot","emailAddress":"flarebot@example.com","displayName":"Flare Bot","active":true,"timeZone":"America/Los_Angeles"},"aggregateprogress":{"progress":0,"total":0},"priority":{"self":"https://mock.atlassian.net/rest/api/2/priority/3","iconUrl":"https://mock.atlassian.net/images/icons/priorities/major.svg","name":"P2 - Major","id":"3"},"customfield_10300":"0|zgcta1:","labels":[],"customfield_10400":null,"timeestimate":null,"aggregatetimeoriginalestimate":null,"progress":{"progress":0,"total":0},"issuelinks":[],"comment":{"comments":[],"maxResults":0,"total":0,"startAt":0},"votes":{"self":"https://mock.atlassian.net/rest/api/2/issue/MOCK-ISSUE-ID/votes","votes":0,"hasVoted":false},"worklog":{"startAt":0,"maxResults":20,"total":0,"worklogs":[]},"assignee":{"self":"https://mock.atlassian.net/rest/api/2/user?username=alice.smith","name":"alice.smith","key":"alice.smith","emailAddress":"alice.smith@example.com","displayName":"Alice Smith","active":true,"timeZone":"America/Los_Angeles"},"updated":"2016-06-15T16:00:48.857-0700","status":{"self":"https://mock.atlassian.net/rest/api/2/status/10800","description":"","iconUrl":"https://mock.atlassian.net/images/icons/statuses/generic.png","name":"Mitigated","id":"10800","statusCategory":{"self":"https://mock.atlassian.net/rest/api/2/statuscategory/3","id":3,"key":"done","colorName":"green","name":"Done"}}}}`

var mockTransitionsContent = `{"expand":"transitions","transitions":[{"id":"161","name":"Mitigate","to":{"self":"https://mock.atlassian.net/rest/api/2/status/10800","description":"","iconUrl":"https://mock.atlassian.net/images/icons/statuses/generic.png","name":"Mitigated","id":"10800","statusCategory":{"self":"https://mock.atlassian.net/rest/api/2/statuscategory/3","id":3,"key":"done","colorName":"green","name":"Done"}},"hasScreen":false},{"id":"221","name":"Not A Flare","to":{"self":"https://mock.atlassian.net/rest/api/2/status/10801","description":"","iconUrl":"https://mock.atlassian.net/images/icons/statuses/generic.png","name":"NotAFlare","id":"10801","statusCategory":{"self":"https://mock.atlassian.net/rest/api/2/statuscategory/3","id":3,"key":"done","colorName":"green","name":"Done"}},"hasScreen":false}]}`

//
// </TEST_DATA>
//

func CreateTestJiraServer() JiraService {
	return &JiraServer{
		Origin:      mockOrigin,
		Username:    mockUsername,
		Password:    mockPassword,
		ProjectID:   mockProjectID,
		IssueTypeID: mockIssueTypeID,
		PriorityIDs: mockPriorityIDs,
	}
}

func TestGetUserByEmail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	jiraServiceCalled := false

	httpmock.RegisterResponder("GET", mockOrigin+"/rest/api/2/user/search",
		func(req *http.Request) (*http.Response, error) {
			jiraServiceCalled = true

			// TODO: add more checks that the call includes the right parameters

			return httpmock.NewStringResponse(200, `[{"self":"https://foobar.atlassian.net/rest/api/2/user?username=alice.smith","key":"alice.smith","name":"alice.smith","emailAddress":"alice.smith@example.com","displayName":"Alice Smith","active":true,"timeZone":"America/Los_Angeles","locale":"en_US"}]`), nil
		},
	)

	testServer := CreateTestJiraServer()

	theUser, err := testServer.GetUserByEmail("alice.smith@example.com")

	assert.True(t, jiraServiceCalled)

	assert.NoError(t, err)
	assert.Equal(t, theUser.Key, "alice.smith")
	assert.Equal(t, theUser.Name, "alice.smith")
	assert.Equal(t, theUser.EmailAddress, "alice.smith@example.com")
}

func TestGetTicketByKey(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	jiraServiceCalled := false

	httpmock.RegisterResponder("GET", mockOrigin+"/rest/api/2/issue/"+mockIssueID,
		func(req *http.Request) (*http.Response, error) {
			jiraServiceCalled = true

			return httpmock.NewStringResponse(200, mockIssueContent), nil
		},
	)

	testServer := CreateTestJiraServer()

	theTicket, err := testServer.GetTicketByKey(mockIssueID)

	assert.True(t, jiraServiceCalled)

	assert.NoError(t, err)
	if err == nil {
		assert.Equal(t, theTicket.Key, mockIssueID)
	}
}

func TestCreateTicket(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	jiraServiceCalled := false

	httpmock.RegisterResponder("POST", mockOrigin+"/rest/api/2/issue",
		func(req *http.Request) (*http.Response, error) {
			jiraServiceCalled = true

			return httpmock.NewStringResponse(200, mockIssueContent), nil
		},
	)

	testServer := CreateTestJiraServer()

	theTicket, err := testServer.CreateTicket(0, "It's a problem", &User{
		Name:         "alice.smith",
		EmailAddress: "alice.smith@example.com",
	})

	assert.True(t, jiraServiceCalled)

	assert.NoError(t, err)
	if err == nil {
		assert.Equal(t, theTicket.Key, mockIssueID)
	}
}

func TestAssignTicketToUser(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	jiraServiceCalled := false

	httpmock.RegisterResponder("PUT", mockOrigin+"/rest/api/2/issue/MOCK-ISSUE-ID",
		func(req *http.Request) (*http.Response, error) {
			jiraServiceCalled = true

			return httpmock.NewStringResponse(200, ""), nil
		})

	testServer := CreateTestJiraServer()

	err := testServer.AssignTicketToUser(&Ticket{
		//		Url:        "https://mock.atlassian.net/issues/MOCK-ISSUE-ID",
		Key: "MOCK-ISSUE-ID",
		Fields: TicketFields{
			Project: Project{
				ID:  "MOCK-PROJECT-ID",
				Key: "MOCK-PROJECT-KEY",
			},
		},
	},
		&User{
			Key:          "alice.smith",
			Name:         "alice.smith",
			EmailAddress: "alice.smith@example.com",
		})

	assert.True(t, jiraServiceCalled)

	assert.NoError(t, err)
}

func TestDoTicketTransition(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	jiraServiceCalled := false

	transitionsURL := "/rest/api/2/issue/MOCK-ISSUE-ID/transitions"

	// mock returning the transitions
	httpmock.RegisterResponder("GET", mockOrigin+transitionsURL,
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(200, mockTransitionsContent), nil
		})

	httpmock.RegisterResponder("POST", mockOrigin+transitionsURL,
		func(req *http.Request) (*http.Response, error) {
			jiraServiceCalled = true

			return httpmock.NewStringResponse(200, mockIssueContent), nil
		},
	)

	testServer := CreateTestJiraServer()

	err := testServer.DoTicketTransition(&Ticket{
		Key: "MOCK-ISSUE-ID",
		Fields: TicketFields{
			Project: Project{
				ID:  "MOCK-PROJECT-ID",
				Key: "MOCK-PROJECT-KEY",
			},
		},
	}, "Mitigate")

	assert.True(t, jiraServiceCalled)

	assert.NoError(t, err)
}
