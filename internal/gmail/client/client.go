package gmailclient

import (
	"context"
	"encoding/base64"

	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	gmailconfig "github.com/mar10br0/email-forward/internal/gmail/config"
	"github.com/mar10br0/email-forward/internal/log"
)

func Connect(ctx context.Context) (svc *gmail.Service, err error) {
	log.Progress("Connecting to Gmail")

	var config *oauth2.Config
	var token *oauth2.Token
	if config, token, err = gmailconfig.Load(); err != nil {
		return
	}

	client := config.Client(ctx, token)
	if svc, err = gmail.NewService(ctx, option.WithHTTPClient(client)); err != nil {
		return
	}

	return
}

func ImportMessage(gmailSvc *gmail.Service, messageID string, messageData string) (err error) {
	log.Debug("Gmail-import message %s as unread into INBOX", messageID)
	var message gmail.Message
	message.Raw = base64.URLEncoding.EncodeToString([]byte(messageData))
	message.LabelIds = []string{"INBOX", "UNREAD"}
	var response *gmail.Message
	response, err = gmailSvc.Users.Messages.Import("me", &message).Do()
	log.Response(response.ServerResponse.Header)

	return
}
