package main

import (
	"PaaS-deployment-sqs-trigger/db"
	"PaaS-deployment-sqs-trigger/environment"
	"PaaS-deployment-sqs-trigger/services"
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	log.SetFlags(log.Flags() | log.Lshortfile)
	lambda.Start(handler)
}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {

	environment.InitEnv()
	db.InitializePostgres()

	var e error

	for _, message := range sqsEvent.Records {

		appId := message.MessageAttributes["APP_UUID"].StringValue

		if appId == nil {
			return nil
		}

		e = services.DeployApp(*appId)

	}

	return e
}
