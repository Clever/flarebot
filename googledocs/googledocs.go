// https://github.com/google/google-api-go-client/blob/master/examples/drive.go#L33

package googledocs

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v2"
)

type Doc struct {
	File *drive.File
}

type GoogleDocsService interface {
	CreateFromTemplate(title string, properties map[string]string) (*Doc, error)
	SetDocPermissionTypeRole(doc *Doc, permissionType string, permissionRole string) error
	GetDoc(fileID string) (*Doc, error)
	GetDocContent(doc *Doc, reltype string) (string, error)
	UpdateDocContent(doc *Doc, content string) error
}

type GoogleDocsServer struct {
	clientID      string
	clientSecret  string
	accessToken   *oauth2.Token
	client        *http.Client
	service       *drive.Service
	templateDocID string
}

func NewGoogleDocsServerWithServiceAccount(jsonConfigString string, templateDocID string) (*GoogleDocsServer, error) {
	jsonBytes := []byte(jsonConfigString)

	conf, err := google.JWTConfigFromJSON(jsonBytes, drive.DriveScope)
	if err != nil {
		log.Fatal(err)
	}

	// service client
	oauthClient := conf.Client(oauth2.NoContext)

	service, err := drive.New(oauthClient)

	if err != nil {
		return nil, err
	}

	return &GoogleDocsServer{
		client:        oauthClient,
		service:       service,
		templateDocID: templateDocID,
	}, nil
}

func NewGoogleDocsServer(clientID string, clientSecret string, accessToken *oauth2.Token, templateDocID string) (*GoogleDocsServer, error) {
	var config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{drive.DriveScope},
	}

	// instantiate the Google Drive client
	oauthClient := config.Client(context.TODO(), accessToken)

	service, err := drive.New(oauthClient)

	if err != nil {
		return nil, err
	}

	return &GoogleDocsServer{
		clientID:      clientID,
		clientSecret:  clientSecret,
		accessToken:   accessToken,
		client:        oauthClient,
		service:       service,
		templateDocID: templateDocID,
	}, nil
}

func (server *GoogleDocsServer) CreateFromTemplate(title string, properties map[string]string) (*Doc, error) {
	// add metadata properties
	propertiesArray := make([]*drive.Property, 0, len(properties))
	if properties != nil {
		for k, v := range properties {
			propertiesArray = append(propertiesArray, &drive.Property{
				Key:        k,
				Value:      v,
				Visibility: "public",
			})
		}
	}

	file := &drive.File{
		Title:      title,
		Properties: propertiesArray,
	}

	file, err := server.service.Files.Copy(server.templateDocID, &drive.File{
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
	permissions, err := server.service.Permissions.List(file.Id).Do()
	if err != nil {
		return err
	}

	// look for the right permission and update it to "Writer"
	for _, perm := range permissions.Items {
		if perm.Type == permissionType {
			perm.Role = permissionRole
			_, err = server.service.Permissions.Update(file.Id, perm.Id, perm).Do()
			return err
		}
	}

	return fmt.Errorf("could not find permission of type %s", permissionType)
}

func (server *GoogleDocsServer) ShareDocWithDomain(doc *Doc, domain string, permissionRole string) error {
	file := doc.File

	// make it editable by the entire organization
	newPermission := &drive.Permission{Value: domain, Type: "domain", Role: permissionRole}

	// create a new permission
	_, err := server.service.Permissions.Insert(file.Id, newPermission).Do()

	return err
}

func (server *GoogleDocsServer) GetDoc(fileID string) (*Doc, error) {
	f, err := server.service.Files.Get(fileID).Do()
	if err != nil {
		return nil, err
	}

	return &Doc{
		File: f,
	}, nil
}

func (server *GoogleDocsServer) GetDocContent(doc *Doc, reltype string) (string, error) {
	relLink := doc.File.ExportLinks[reltype]

	response, err := server.client.Get(relLink)

	if err != nil {
		return "", err
	}

	defer response.Body.Close()
	bytes, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func (server *GoogleDocsServer) UpdateDocContent(doc *Doc, content string) error {
	server.service.Files.Update(doc.File.Id, doc.File).Media(strings.NewReader(content)).Do()
	return nil
}
