package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"io/ioutil"
)

type User struct {
	AccountId    string `json:"accountId"`
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
	Key    string       `json:"key"`
	Fields TicketFields `json:"fields"`
}

// tuned for a single project
type JiraServer struct {
	Origin   string
	Username string
	Password string
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

func (server *JiraServer) GetTicketByKey(key string) (*Ticket, error) {
	var ticket Ticket
	err := server.DoRequest("GET", fmt.Sprintf("/rest/api/2/issue/%s", key), nil, &ticket)

	if err != nil {
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

func (server *JiraServer) SetLabel(ticket *Ticket, label string) error {
	request := map[string]interface{}{
		"update": &map[string]interface{}{
			"labels": []map[string]interface{}{
				{"add": label},
			},
		},
	}
	return server.UpdateTicket(ticket, request)
}
