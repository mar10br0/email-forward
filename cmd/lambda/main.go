package main

import (
	"github.com/aws/aws-lambda-go/lambda"

	seshandler "github.com/mar10br0/email-forward/internal/ses/handler"
)

func main() {
	lambda.Start(seshandler.HandleRequest)
}
