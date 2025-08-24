package uv

import (
	"cerpApi/cfg_details"
	"cerpApi/notifications"
	"cerpApi/otp"
	"cerpApi/students"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var DOC_TYPE = []string{"MC", "HT"}

var DOC_TYPE_MAP = map[string]map[string]string{
	"MC": {
		"name":    "MARKS CARD",
		"content": "Marks Card {{batch}}",
	},
	"HT": {
		"name":    "HALL TICKET",
		"content": "Marks Card {{batch}}",
	},
}

var DOC_TYPE_EMPTY = []string{}

var DOC_BY_COLLEGE = map[string][]string{
	"ni": DOC_TYPE,
}

type OnboardUvMetadata struct {
	PK      string      `dynamodbav:"vkey1"`
	SK      string      `dynamodbav:"vkey2"`
	Value   interface{} `dynamodbav:"value"`
	Ts      int64       `dynamodbav:"ts"`
	Updater string      `dynamodbav:"uby"`
}

func OnboardUvDocs(colId string, batch string, docType string, studentList []string, uBy string) error {
	PKKey := batch
	SKKey := docType

	onF := OnboardUvMetadata{
		PK:      PKKey,
		SK:      SKKey,
		Value:   studentList,
		Ts:      time.Now().Unix(),
		Updater: uBy,
	}
	ctx, cancel := cfg_details.GetTimeoutContext()
	defer cancel()
	data, err := attributevalue.MarshalMap(onF)
	log.Printf("Error in OnboardUvDocs Marshal %v\n", err)
	_, err = cfg_details.DynamoCfg.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(colId + "_uv"),
		Item:      data,
	})
	log.Printf("Error in OnboardUvDocs in Dynamodb %v\n", err)
	return err
}

func ListUvDocsStudents(colId string, batch string, docType string) ([]string, error) {
	PKKey := batch
	SKKey := docType

	ctx, cancel := cfg_details.GetTimeoutContext()
	defer cancel()
	out, err := cfg_details.DynamoCfg.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(colId + "_uv"),
		Key: map[string]types.AttributeValue{
			"vkey1": &types.AttributeValueMemberS{Value: PKKey},
			"vkey2": &types.AttributeValueMemberS{Value: SKKey},
		},
	})
	if err != nil {
		log.Printf("Error in ListUvDocs while querying the DDB %v\n", err)
		return nil, err
	}
	if len(out.Item) == 0 {
		log.Printf("No data found for the college %s batch %s docType %s\n", colId, batch, docType)
		return nil, nil
	}
	var uvDocs []string
	err = attributevalue.Unmarshal(out.Item["value"], &uvDocs)
	if err != nil {
		log.Printf("Error unmarshaling UV docs %v\n", err)
		return nil, err
	}
	return uvDocs, nil
}

type CollectDocumentDetails struct {
	CollectedDate string `json:"collected_date"`
	Description   string `json:"description"`
	OtpVerified   bool   `json:"otp_verified"`
}

type CollectDocumentWrapper struct {
	Key1            string                 `dynamodbav:"vkey1"`
	Key2            string                 `dynamodbav:"vkey2"`
	Ts              int64                  `dynamodbav:"ts"`
	CollectDocument CollectDocumentDetails `dynamodbav:"collect_document_details"`
	Updater         string                 `dynamodbav:"uby"`
}

func CollectDocument(colId string, batch string, docType string, studentId string, description string, otp string, uBy string) error {
	cdWrap := getCollectDocumentWrapper(batch, docType, studentId, description, uBy)
	if ok, err := verifyOtp(colId, studentId, cdWrap, otp, docType); err != nil || !ok {
		log.Printf("OTP verification failed for UV %s %s %s\n", colId, batch, studentId)
		return errors.New("OTP verification failed")
	}
	cdWrap.Ts = cfg_details.GetCurrentTs()
	cdWrap.CollectDocument.OtpVerified = true
	ctx, cancel := cfg_details.GetTimeoutContext()
	defer cancel()
	data, err := attributevalue.MarshalMap(cdWrap)
	if err != nil {
		log.Printf("Error in Collecting the document MarshalMap %v\n", err)
		return err
	}
	_, err = cfg_details.DynamoCfg.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(colId + "_uv"),
		Item:                data,
		ConditionExpression: aws.String("attribute_not_exists(vkey1) AND attribute_not_exists(vkey2)"),
	})
	if err != nil {
		if strings.Contains(err.Error(), "ConditionalCheckFailedException") {
			log.Printf("Document already collected for %s %s %s\n", colId, batch, studentId)
			return errors.New("document already collected")
		}
		log.Printf("Error in Collecting the document in Dynamodb %v\n", err)
	}
	return err
}

func verifyOtp(colId string, studentId string, cdWrap CollectDocumentWrapper, otp string, docType string) (bool, error) {
	hashStr, err := cfg_details.GenerateHash(cdWrap)
	if err != nil {
		return false, err
	}
	otpDetails, err := fetchSavedOtp(colId, studentId, docType)
	if err != nil {
		return false, err
	}
	currentTtl := cfg_details.GetCurrentTs()
	otpTtl := cfg_details.ParseStrToInt64(otpDetails["ttl"])
	if currentTtl > otpTtl {
		mes := "OTP expired"
		log.Printf("%s\n", mes)
		return false, errors.New(mes)
	}
	if otp != otpDetails["otp"] {
		mes := "OTP didn't match"
		log.Printf("%s\n", mes)
		return false, errors.New(mes)
	}
	if hashStr != otpDetails["hash"] {
		mes := "content changed otp didn't match"
		log.Printf("%s\n", mes)
		return false, errors.New(mes)
	}
	log.Printf("OTP verified successfully for %s %s %s\n", colId, studentId, docType)
	return true, nil
}

func fetchSavedOtp(colId string, sId string, docType string) (map[string]string, error) {
	key, err := attributevalue.Marshal(sId)
	if err != nil {
		log.Printf("Error in fetchSavedOtp while marshaling the key %s %v\n", sId, err)
		return nil, err
	}
	sKey, err := attributevalue.Marshal(docType)
	if err != nil {
		log.Printf("Error in fetchSavedOtp while marshaling vault skey - %v\n", err)
		return nil, err
	}
	ck := map[string]types.AttributeValue{
		"sid":      key,
		"otp_type": sKey,
	}
	ctx, cancel := cfg_details.GetTimeoutContext()
	defer cancel()
	out, err := cfg_details.DynamoCfg.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(colId + "_otp"),
		Key:       ck,
	})
	if err != nil {
		log.Printf("Error in fetchSavedOtp while querying the DDB %v\n", err)
		return nil, err
	}
	if out == nil || len(out.Item) == 0 {
		log.Printf("NoOTP data found for the user StudentId=%s colId=%s DocType=%s %v\n", sId, colId, docType, err)
		return nil, nil
	}
	mapOtpData := make(map[string]string)
	mapOtpData["hash"] = out.Item["hash"].(*types.AttributeValueMemberS).Value
	mapOtpData["otp"] = out.Item["otp"].(*types.AttributeValueMemberS).Value
	mapOtpData["ttl"] = out.Item["ttl"].(*types.AttributeValueMemberS).Value
	log.Printf("Fetched OTP data: %v\n", mapOtpData)
	return mapOtpData, nil
}

func SendOtpVerification(colId string, batch string, docType string, studentId string, description string, uBy string) (string, error) {
	cdWrap := getCollectDocumentWrapper(batch, docType, studentId, description, uBy)
	hashStr, err := cfg_details.GenerateHash(cdWrap)
	if err != nil {
		log.Printf("Error generating hash for otp %v\n", err)
	}
	otpData, err := otp.GenerateOtp(hashStr, "uv")
	if err != nil {
		log.Printf("Error generating in uv SendOtpVerification %v\n", err)
	}
	ttl := cfg_details.GenerateTtl(5)
	otp.PersistOtp(colId, studentId, "", hashStr, otpData, docType, ttl)
	email, err := students.GetStudentEmailById(colId, studentId)
	if err != nil {
		return "", err
	}
	content := getContent(docType)
	notifications.SendOtp(studentId, content["name"], ttl, content["content"], email, otpData)
	return otpData, err
}

func getContent(docType string) map[string]string {
	return DOC_TYPE_MAP[docType]
}

func getCollectDocumentWrapper(batch string, docType string, studentId string, description string, uBy string) CollectDocumentWrapper {
	collectDocumentDetails := CollectDocumentDetails{
		CollectedDate: cfg_details.GetCurrentDate(),
		Description:   description,
		OtpVerified:   false,
	}
	key1 := batch + "_" + docType
	cdWrap := CollectDocumentWrapper{
		Key1:            key1,
		Key2:            studentId,
		CollectDocument: collectDocumentDetails,
		Updater:         uBy,
	}
	return cdWrap
}

func GetUvDocs(colId string) []string {
	if val, ok := DOC_BY_COLLEGE[colId]; ok {
		return val
	}
	return DOC_TYPE_EMPTY
}
