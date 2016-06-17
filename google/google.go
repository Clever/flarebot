// https://github.com/google/google-api-go-client/blob/master/examples/drive.go#L33

package google

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"io/ioutil"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v2"
	"google.golang.org/api/googleapi/transport"
)

type Doc struct {
	file *drive.File
}

type GoogleDocsService interface {
	CreateFromTemplate(templateDocID string, title string) (*Doc, err)
	SetDocPermissionTypeRole(doc *Doc, permissionType string, permissionRole string) err
}

type GoogleDocsServer struct {
	ClientID string
	ClientSecret string
	AccessToken string
	TemplateDocID string
}


func (server *GoogleDocsServer) CreateFromTemplate(templateDocID string, title string) (*Doc, err) {
}


func (server *GoogleDocsServer) SetDocPermissionTypeRole(doc *Doc, permissionType string, permissionRole string) err {
}
