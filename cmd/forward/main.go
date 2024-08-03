package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	sesevent "github.com/mar10br0/email-forward/internal/ses/event"
	seshandler "github.com/mar10br0/email-forward/internal/ses/handler"
)

func main() {
	if len(os.Args) < 2 {
		log.Failure(nil, "Usage: %s <recipient-email>", os.Args[0])
		return
	}

	var event events.SimpleEmailEvent
	var err error

	s3Svc := s3.New(session.Must(session.NewSession()), aws.NewConfig().WithLogLevel(log.LEVEL))

	if event, err = sesevent.Build(s3Svc, os.Args[1]); err != nil {
		log.Failure(err, "")
		return
	}

	if len(event.Records) > 0 {
		if err = seshandler.HandleRequest(context.Background(), event); err != nil {
			log.Failure(err, "")
		} else {
			log.Success("All messages forwarded successfully")
		}
	} else {
		log.Success("Nothing to do")
	}
}
