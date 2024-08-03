package sesevent

import (
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mar10br0/email-forward/internal/log"
)

func Build(svc *s3.S3, recipient string) (event events.SimpleEmailEvent, err error) {
	domain := strings.Split(recipient, "@")[1]
	user := strings.Split(recipient, "@")[0]
	bucket := domain + "-inbox"
	prefix := user + "/"
	log.Progress("Building SES-event for %s in S3 %s/%s", recipient, bucket, prefix)

	var list *s3.ListObjectsV2Output
	if list, err = svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: &bucket, Prefix: &prefix}); err != nil {
		return
	}
	if len(list.Contents) == 0 {
		event.Records = []events.SimpleEmailRecord{}
		return
	}

	for _, item := range list.Contents {
		event.Records = append(event.Records, events.SimpleEmailRecord{
			SES: events.SimpleEmailService{
				Receipt: events.SimpleEmailReceipt{
					Recipients: []string{recipient},
				},
				Mail: events.SimpleEmailMessage{
					MessageID: strings.TrimPrefix(*item.Key, prefix),
				},
			},
		})
	}
	return
}
