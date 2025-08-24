package otp

import (
	"cerpApi/cfg_details"
	"encoding/base32"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func GenerateOtp(secret string, purpose string) (string, error) {
	secretBase32 := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte(secret + purpose))
	otp, err := totp.GenerateCodeCustom(secretBase32, time.Now(), totp.ValidateOpts{
		Period:    300,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		log.Printf("error while generating otp for %s %s %v\n", secret, purpose, err)
		return "", err
	}
	return otp, nil
}

func PersistOtp(colId string, sId string, content string, hashVal string, otpStr string, otpType string, ttl string) error {
	otpData := make(map[string]string)
	otpData["sid"] = sId
	otpData["otp_type"] = otpType
	otpData["otp"] = otpStr
	otpData["hash"] = hashVal
	otpData["ttl"] = ttl
	data, err := attributevalue.MarshalMap(otpData)
	log.Println(otpData)
	if err != nil {
		log.Printf("Error in insertToDb while marshaling the data %v\n", err)
		return err
	}
	ctx, cancel := cfg_details.GetTimeoutContext()
	defer cancel()
	_, err = cfg_details.DynamoCfg.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(colId + "_otp"),
		Item:      data,
	})
	if err != nil {
		log.Printf("Error while writing to DB for the otp %s %v\n", sId, err)
		return err
	}
	return nil
}
