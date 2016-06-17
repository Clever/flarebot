package jira

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"io/ioutil"
)

type Ticket struct {
	Url           string
	Key           string
	ProjectID     string
	ProjectKey    string
	AssigneeEmail string
}

type User struct {
	Key   string
	Name  string
	Email string
}

type JiraService interface {
	GetUserByEmail(email string) (*User, error)
	GetTicketByKey(key string) (*Ticket, error)
	CreateTicket(priority int, topic string, assignee *User) (*Ticket, error)
	AssignTicketToUser(ticket *Ticket, user *User) error
	DoTicketTransition(ticket *Ticket, transitionName string) error
}

// tuned for a single project
type JiraServer struct {
	Origin      string
	Username    string
	Password    string
	ProjectID   string
	IssueTypeID string
	PriorityIDs []string
}

// always return an array of objects, oftentimes just one
func (server *JiraServer) DoRequest(method string, path string, body *map[string]interface{}) ([]map[string]interface{}, error) {
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
		return nil, err
	}

	defer resp.Body.Close()

	responseBody, _ := ioutil.ReadAll(resp.Body)

	if len(responseBody) == 0 {
		return make([]map[string]interface{}, 0), nil
	}

	// always an array
	var resultInterface interface{}
	json.Unmarshal(responseBody, &resultInterface)

	// this is ugly, but what are you gonna do when you don't know what to expect from JIRA?
	// introspect to figure out if it's an array or a map, and then do the right thing,
	// including deep typecasting to get an array of maps if needed.
	var result []map[string]interface{}
	if reflect.TypeOf(resultInterface).Kind() == reflect.Map {
		result = append(result, resultInterface.(map[string]interface{}))
	} else {
		resultWithInterfaces := resultInterface.([]interface{})
		for _, v := range resultWithInterfaces {
			result = append(result, v.(map[string]interface{}))
		}
	}

	return result, nil
}

func (server *JiraServer) GetTicketURL(ticketKey string) string {
	return fmt.Sprintf("%s/issues/%s", server.Origin, ticketKey)
}

func (server *JiraServer) GetUserByEmail(email string) (*User, error) {
	responseArray, err := server.DoRequest("GET", fmt.Sprintf("/rest/api/2/user/search?username=%s", email), nil)

	if err != nil {
		return nil, err
	}

	response := responseArray[0]

	return &User{
		Key:   response["key"].(string),
		Name:  response["name"].(string),
		Email: response["emailAddress"].(string),
	}, nil
}

func (server *JiraServer) GetTicketByKey(key string) (*Ticket, error) {
	responseArray, err := server.DoRequest("GET", fmt.Sprintf("/rest/api/2/issue/%s", key), nil)

	if err != nil {
		return nil, err
	}

	response := responseArray[0]

	if response["fields"] == nil {
		return nil, errors.New("no such ticket")
	}

	var fields map[string]interface{} = response["fields"].(map[string]interface{})
	var project map[string]interface{} = fields["project"].(map[string]interface{})

	var assignee map[string]interface{}
	var assigneeEmail string

	if assignee == nil {
		assignee = nil
	} else {
		assignee = fields["assignee"].(map[string]interface{})
		assigneeEmail = assignee["emailAddress"].(string)
	}

	ticketKey := response["key"].(string)
	return &Ticket{
		Key:           ticketKey,
		Url:           server.GetTicketURL(ticketKey),
		ProjectID:     project["id"].(string),
		ProjectKey:    project["key"].(string),
		AssigneeEmail: assigneeEmail,
	}, nil
}

func (server *JiraServer) CreateTicket(priority int, topic string, assignee *User) (*Ticket, error) {
	// request JSON
	request := &map[string]interface{}{
		"fields": &map[string]interface{}{
			"project": &map[string]interface{}{
				"id": server.ProjectID,
			},
			"issuetype": &map[string]interface{}{
				"id": server.IssueTypeID,
			},
			"assignee": &map[string]interface{}{
				"name": assignee.Name,
			},
			"summary": topic,
			"priority": &map[string]interface{}{
				"id": server.PriorityIDs[priority],
			},
		},
	}

	url := "/rest/api/2/issue"
	responseArray, _ := server.DoRequest("POST", url, request)
	response := responseArray[0]

	return &Ticket{
		Url: server.GetTicketURL(response["key"].(string)),
		Key: response["key"].(string),
	}, nil
}

func (server *JiraServer) UpdateTicket(ticket *Ticket, request *map[string]interface{}) error {
	url := "/rest/api/2/issue/" + ticket.Key
	_, err := server.DoRequest("PUT", url, request)

	// will be nil if no error
	return err
}

func (server *JiraServer) AssignTicketToUser(ticket *Ticket, user *User) error {
	// request JSON
	request := &map[string]interface{}{
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
	transitionsArray, _ := server.DoRequest("GET", transitionsURL, nil)

	transitions := transitionsArray[0]["transitions"].([]interface{})

	var transitionID string

	// find the right transition
	for _, v := range transitions {
		oneTransition := v.(map[string]interface{})
		if oneTransition["name"] == transitionName {
			transitionID = oneTransition["id"].(string)
			break
		}
	}

	if transitionID == "" {
		return errors.New("no transition named '" + transitionName + "'")
	}

	// request JSON
	request := &map[string]interface{}{
		"transition": &map[string]interface{}{
			"id": transitionID,
		},
	}

	_, err := server.DoRequest("POST", transitionsURL, request)

	return err
}
