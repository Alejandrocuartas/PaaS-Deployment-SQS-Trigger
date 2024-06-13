package main

import (
	"context"
	"encoding/json"
	"log"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/joho/godotenv"
)

func Test(t *testing.T) {
	godotenv.Load()
	event := `{
		"Records": [
		 {
			"messageAttributes": {
				"APP_UUID": {
					"stringValue": "5c67ee2f-380a-49a6-9f41-0646d5ef6903",
					"stringListValues": [],
					"binaryListValues": [],
					"dataType": "String"
				}
			},
			"eventSourceARN": "arn:aws:sqs:us-east-1:696144082470:NotificationQueue-Testing",
			"eventSource": "aws:sqs",
			"awsRegion": "us-east-1"
		}
	   ]
	}`

	sqsEvent := events.SQSEvent{}
	e := json.Unmarshal([]byte(event), &sqsEvent)
	if e != nil {
		log.Println(e)
	}
	handler(context.TODO(), sqsEvent)
}
