package gmailconfig

import (
	"encoding/json"
	"os"

	"github.com/mar10br0/email-forward/internal/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"google.golang.org/api/gmail/v1"
)

var CREDENTIALS_FILE = "gmail-credentials.json"
var TOKEN_FILE = "gmail-token.json"
var SCOPES = []string{gmail.GmailMetadataScope, gmail.GmailInsertScope}

func Load() (config *oauth2.Config, token *oauth2.Token, err error) {
	log.Debug("Loading Gmail Credentials from %s", CREDENTIALS_FILE)

	var credentials []byte
	if credentials, err = os.ReadFile(CREDENTIALS_FILE); err != nil {
		return
	}

	// If modifying these scopes, delete your previously saved token.json.
	if config, err = google.ConfigFromJSON(credentials, SCOPES...); err != nil {
		return
	}

	log.Debug("Loading Gmail Token from %s", TOKEN_FILE)

	var f *os.File
	if f, err = os.Open(TOKEN_FILE); err != nil {
		return
	}
	defer f.Close()

	token = &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return
}
