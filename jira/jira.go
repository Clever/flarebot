package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"io/ioutil"
)

type status struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type transition struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	To        status `json:"to"`
	HasScreen bool   `json:"hasScreen"`
}

type transitionResponse struct {
	Transitions []transition `json:"transitions"`
}

type User struct {
	Key          string `json:"key"`
	Name         string `json:"name"`
	EmailAddress string `json:"emailAddress"`
}

type Project struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

type TicketFields struct {
	Project  Project `json:"project"`
	Creator  User    `json:"creator"`
	Reporter User    `json:"reporter"`
	Assignee User    `json:"assignee"`
}

type Ticket struct {
	service JiraService
	Key     string       `json:"key"`
	Fields  TicketFields `json:"fields"`
}

type JiraService interface {
	TicketUrl(ticketKey string) string
	GetUserByEmail(email string) (*User, error)
	GetTicketByKey(key string) (*Ticket, error)
	CreateTicket(priority int, topic string, assignee *User) (*Ticket, error)
	AssignTicketToUser(ticket *Ticket, user *User) error
	DoTicketTransition(ticket *Ticket, transitionName string) error
	SetDescription(ticket *Ticket, description string) error
}

// tuned for a single project
type JiraServer struct {
	Origin      string
	Username    string
	Password    string
	ProjectKey  string
	IssueType   string
	PriorityIDs []string
}

func (ticket *Ticket) Url() string {
	if ticket.service == nil {
		return ""
	}

	return ticket.service.TicketUrl(ticket.Key)
}

// unmarshalls into the provided data structure
func (server *JiraServer) DoRequest(method string, path string, body map[string]interface{}, response interface{}) error {
	fullURL := fmt.Sprintf("%s%s", server.Origin, path)

	var req *http.Request

	if body != nil {
		jsonStr, _ := json.Marshal(body)
		req, _ = http.NewRequest(method, fullURL, bytes.NewBuffer(jsonStr))
		req.Header.Add("Content-Type", "application/json")
	} else {
		req, _ = http.NewRequest(method, fullURL, nil)
	}

	req.SetBasicAuth(server.Username, server.Password)
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Printf("got an error: %s\n", err)
		return err
	}

	defer resp.Body.Close()

	responseBody, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode > 299 {
		return fmt.Errorf("Status-code:%d, error: %s", resp.StatusCode, responseBody)
	}

	if len(responseBody) == 0 || response == nil {
		return nil
	}

	// return err, should be nil if no problem
	// result is unmarshalled into response
	return json.Unmarshal(responseBody, response)
}

func (server *JiraServer) TicketUrl(ticketKey string) string {
	return fmt.Sprintf("%s/browse/%s", server.Origin, ticketKey)
}

func (server *JiraServer) GetUserByEmail(email string) (*User, error) {
	var users []User
	err := server.DoRequest("GET", fmt.Sprintf("/rest/api/2/user/search?username=%s", email), nil, &users)

	if err != nil {
		return nil, err
	}
	return &users[0], nil
}

func (server *JiraServer) GetTicketByKey(key string) (*Ticket, error) {
	var ticket Ticket
	err := server.DoRequest("GET", fmt.Sprintf("/rest/api/2/issue/%s", key), nil, &ticket)

	if err != nil {
		return nil, err
	}

	ticket.service = server
	return &ticket, nil
}

func (server *JiraServer) CreateTicket(priority int, topic string, assignee *User) (*Ticket, error) {
	// request JSON
	request := map[string]interface{}{
		"fields": &map[string]interface{}{
			"project": &map[string]interface{}{
				"key": server.ProjectKey,
			},
			"issuetype": &map[string]interface{}{
				"name": server.IssueType,
			},
			"assignee": &map[string]interface{}{
				"name": assignee.Name,
			},
			"summary": topic,
			"priority": &map[string]interface{}{
				"id": fmt.Sprintf("%d", priority),
			},
		},
	}

	var ticket Ticket

	url := "/rest/api/2/issue"
	err := server.DoRequest("POST", url, request, &ticket)

	ticket.service = server

	if err != nil {
		fmt.Printf("\nERROR: %s\n", err)
		return nil, err
	}

	return &ticket, nil
}

func (server *JiraServer) UpdateTicket(ticket *Ticket, request map[string]interface{}) error {
	url := "/rest/api/2/issue/" + ticket.Key
	err := server.DoRequest("PUT", url, request, nil)

	// will be nil if no error
	return err
}

func (server *JiraServer) AssignTicketToUser(ticket *Ticket, user *User) error {
	// request JSON
	request := map[string]interface{}{
		"fields": &map[string]interface{}{
			"assignee": &map[string]interface{}{
				"name": user.Name,
			},
		},
	}

	return server.UpdateTicket(ticket, request)
}

func (server *JiraServer) DoTicketTransition(ticket *Ticket, transitionName string) error {
	// get the transitions that are allowed and find the right one.
	transitionsURL := fmt.Sprintf("/rest/api/2/issue/%s/transitions", ticket.Key)

	var response transitionResponse
	err := server.DoRequest("GET", transitionsURL, nil, &response)

	// find the right transition
	var theTransition *transition
	for _, v := range response.Transitions {
		if v.Name == transitionName {
			theTransition = &v
			break
		}
	}

	if theTransition == nil {
		return fmt.Errorf("no transition named %s", transitionName)
	}

	// request JSON
	request := map[string]interface{}{
		"transition": theTransition,
	}

	err = server.DoRequest("POST", transitionsURL, request, nil)

	return err
}

func (server *JiraServer) SetDescription(ticket *Ticket, description string) error {
	editURL := fmt.Sprintf("/rest/api/2/issue/%s", ticket.Key)
	request := map[string]interface{}{
		"fields": &map[string]interface{}{
			"description": description,
		},
	}
	return server.DoRequest("PUT", editURL, request, nil)
}
