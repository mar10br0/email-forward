package seshandler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/google/uuid"
	"google.golang.org/api/gmail/v1"

	gmailclient "github.com/mar10br0/email-forward/internal/gmail/client"
	"github.com/mar10br0/email-forward/internal/log"
	sesmessage "github.com/mar10br0/email-forward/internal/ses/message"
)

func HandleRequest(ctx context.Context, event events.SimpleEmailEvent) error {
	log.Progress("Handle SES-request for %s", log.Plurality(int64(len(event.Records)), "messages"))
	log.Debug("%v", event)

	s3Svc := s3.New(session.Must(session.NewSession()), aws.NewConfig().WithLogLevel(log.LOGLEVEL))

	var gmailSvc *gmail.Service
	var err error
	if gmailSvc, err = gmailclient.Connect(ctx); err != nil {
		return err
	}

	var deleteFailures []string
	for _, record := range event.Records {
		var recipient, id, data string
		recipient = record.SES.Receipt.Recipients[0]
		id = record.SES.Mail.MessageID
		log.Progress("Forwarding message %s for %s", id, recipient)

		if data, err = sesmessage.Fetch(s3Svc, recipient, id); err != nil {
			log.Failure(err, "")
		} else {
			if err = gmailclient.ImportMessage(gmailSvc, id, data); err != nil {
				log.Failure(err, "")
			} else {
				if err = sesmessage.Delete(s3Svc, recipient, id); err != nil {
					domain := strings.Split(recipient, "@")[1]
					user := strings.Split(recipient, "@")[0]
					deleteFailures = append(deleteFailures, domain+"-inbox/"+user+"/"+id)
					log.Failure(err, "Import succeeded, but deleting from S3 bucket failed")
				}
			}
		}
	}

	if len(deleteFailures) > 0 {
		profile, _ := gmailSvc.Users.GetProfile("me").Do()
		email := profile.EmailAddress
		notificationID := fmt.Sprintf("<%s@email-forward>", uuid.New().String())
		notification := append(
			[]string{
				"Message-Id: " + notificationID,
				"Date: " + time.Now().Format(time.RFC1123Z),
				fmt.Sprintf("From: \"Email Forwarder\" <%s>", email),
				"To: " + email,
				fmt.Sprintf("Subject: Failed to delete %s from S3!", log.Plurality(int64(len(deleteFailures)), "objects")),
				"Content-Type: text/plain; charset=utf-8",
				"",
			}[0:7], deleteFailures...)
		gmailclient.ImportMessage(gmailSvc, notificationID, strings.Join(notification, "\r\n"))
	}

	return nil
}
