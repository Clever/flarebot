// https://github.com/google/google-api-go-client/blob/master/examples/drive.go#L33

package google

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"errors"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v2"
)

type Doc struct {
	File *drive.File
}

type GoogleDocsService interface {
	CreateFromTemplate(title string) (*Doc, error)
	SetDocPermissionTypeRole(doc *Doc, permissionType string, permissionRole string) error
}

type GoogleDocsServer struct {
	ClientID string
	ClientSecret string
	AccessToken *oauth2.Token
	Service *drive.Service
	TemplateDocID string
}


func decodeOAuthToken(tokenString string) *oauth2.Token {
	tokenBytes, _ := base64.StdEncoding.DecodeString(tokenString)
	tokenBytesBuffer := bytes.NewBuffer(tokenBytes)
	dec := gob.NewDecoder(tokenBytesBuffer)
	token := new(oauth2.Token)
	dec.Decode(token)

	return token
}

func NewGoogleDocsServer(clientID string, clientSecret string, accessToken string, templateDocID string) (*GoogleDocsServer, error) {
	token := decodeOAuthToken(accessToken)

	var config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{drive.DriveScope},
	}

	// instantiate the Google Drive client
	oauthClient := config.Client(context.Background(), token)
	service, err := drive.New(oauthClient)

	if err != nil {
		return nil, err
	}
	
	return &GoogleDocsServer{
		ClientID: clientID,
		ClientSecret: clientSecret,
		AccessToken: token,
		Service: service,
		TemplateDocID: templateDocID,
	}, nil
}

func (server *GoogleDocsServer) CreateFromTemplate(title string) (*Doc, error) {
	file, err := server.Service.Files.Copy(server.TemplateDocID, &drive.File{
		Title: title,
	}).Do()

	if err != nil {
		return nil, err
	}
	
	return &Doc{
		File: file,
	}, nil
}


func (server *GoogleDocsServer) SetDocPermissionTypeRole(doc *Doc, permissionType string, permissionRole string) error {
	file := doc.File
	
	// make it editable by the entire organization
	permissions, err := server.Service.Permissions.List(file.Id).Do()
	if err != nil {
		return err
	}

	// look for the right permission and update it to "Writer"
	for _, perm := range permissions.Items {
		if perm.Type == permissionType {
			perm.Role = permissionRole
			_, err = server.Service.Permissions.Update(file.Id, perm.Id, perm).Do()
			return err
		}
	}

	return errors.New("could not find permission of type " + permissionType)
}
