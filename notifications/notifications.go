package notifications

import (
	"cerpApi/cfg_details"
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"log"
)

var Q_URL = "https://sqs.ap-south-1.amazonaws.com/214672925111/attendance"

type AbsentNotificationWrapper struct {
	NotifyChannel string        `json:"notify_channel"`
	Data          []interface{} `json:"data"`
	Timeslot      string        `json:"time_slot"`
	Date          string        `json:"date"`
	Subject       string        `json:"class"`
}

func NotifyUsers(mes AbsentNotificationWrapper, channel string) {
	switch channel {
	case "Q":
		jsonStr, err := json.Marshal(mes)
		if err != nil {
			log.Printf("Error while adding message to the Q in NotifyUsers fn %v\n", err)
		} else {
			addQMes(string(jsonStr))
		}
		break
	}
}

func addQMes(mes string) {
	log.Printf("In Absent notification System - addQMes...")
	_, err := cfg_details.SqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:    &Q_URL,
		MessageBody: &mes,
	})
	if err != nil {
		log.Printf("Error sending message: %v\n", err)
	}
}
