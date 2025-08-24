package notifications

import (
	"cerpApi/cfg_details"
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

var Q_URL = "https://sqs.ap-south-1.amazonaws.com/214672925111/attendance"

type AbsentNotificationWrapper struct {
	NotifyChannel string        `json:"notify_channel"`
	Data          []interface{} `json:"data"`
	Timeslot      string        `json:"time_slot"`
	Date          string        `json:"date"`
	Subject       string        `json:"class"`
}

type OtpType struct {
	OtpType string `json:"otp_type"`
	Ts      string `json:"ts"`
	Otp     string `json:"otp"`
}

type StudentOtpWrapper struct {
	Sid     string  `json:"sid"`
	Type    OtpType `json:"otp_type"`
	Content string  `json:"content"`
	Email   string  `json:"email"`
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
	}
}

func SendOtp(sId string, purpose string, ts string, content string, email string, otp string) error {
	otpType := OtpType{
		OtpType: purpose,
		Ts:      ts,
		Otp:     otp,
	}
	studentOtpWrap := StudentOtpWrapper{
		Sid:     sId,
		Type:    otpType,
		Content: content,
		Email:   email,
	}
	studentOtpWrapJsonStr, err := json.Marshal(studentOtpWrap)
	if err != nil {
		log.Printf("Error while sending the otp for the student - %s", sId)
		return err
	}
	return addQMes(string(studentOtpWrapJsonStr))
}

func addQMes(mes string) error {
	log.Printf("In notification System - addQMes...")
	_, err := cfg_details.SqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:    &Q_URL,
		MessageBody: &mes,
	})
	if err != nil {
		log.Printf("Error sending message: %v\n", err)
		return err
	}
	return nil
}
