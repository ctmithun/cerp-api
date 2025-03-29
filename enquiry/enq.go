package enquiry

import (
	"cerpApi/cfg_details"
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"strconv"
	"time"
)

const ENQ_TABLE_NAME = "_enquiry"

type EnquiryList struct {
	EnqList  []EnquiryData `json:"enq_list"`
	NextPage string        `json:"next_page"`
}

type EnquiryData struct {
	Enq EnquiryForm `dynamodbav:"value"`
	Id  string      `dynamodbav:"uid"`
}
type EnquiryForm struct {
	Name           string   `json:"name"`
	Course         string   `json:"course"`
	Mobile         string   `json:"mobile"`
	ParentMobile   string   `json:"parent_mobile"`
	Reference      string   `json:"reference"`
	Location       string   `json:"location"`
	Comments       []string `json:"comments"`
	Status         string   `json:"status"`
	CouncillorId   string   `json:"councillor_id"`
	CouncillorName string   `json:"councillor_name"`
}

type enquiryFormOnboard struct {
	Uid     string      `dynamodbav:"uid"`
	User    string      `dynamodbav:"user"`
	Value   EnquiryForm `dynamodbav:"value"`
	Ts      int64       `dynamodbav:"ts"`
	Updater string      `dynamodbav:"uBy"`
}

const ACTIVE_ENQ = "active"

func DeactivateActiveEnquiry(college string, enqId string) error {
	return deactivateActiveEnquiry(college, enqId)
}

func deactivateActiveEnquiry(college string, enqId string) error {
	if enqId == "" {
		return errors.New(cfg_details.INPUT_ERROR)
	}
	_, err := cfg_details.DynamoCfg.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		Key: map[string]types.AttributeValue{
			"uid":  &types.AttributeValueMemberS{Value: ACTIVE_ENQ},
			"user": &types.AttributeValueMemberS{Value: enqId},
		},
		TableName: aws.String(college + ENQ_TABLE_NAME),
	})
	if err != nil {
		return errors.New("Error deleting enquiry item: - " + enqId + " - " + err.Error())
	}
	return nil
}

func UpdateActiveEnquiry(college string, key string, comments string, uBy string) error {
	commentsList := make([]types.AttributeValue, 0)
	commentsList = append(commentsList, &types.AttributeValueMemberS{Value: comments})
	_, err := cfg_details.DynamoCfg.UpdateItem(
		context.TODO(),
		&dynamodb.UpdateItemInput{
			TableName: aws.String(college + ENQ_TABLE_NAME),
			Key: map[string]types.AttributeValue{
				"uid":  &types.AttributeValueMemberS{Value: key},
				"user": &types.AttributeValueMemberS{Value: key},
			},
			UpdateExpression: aws.String("SET #val.#Comments = list_append(#val.#Comments, :comments), ts = :ts, uBy = :uBy"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":comments": &types.AttributeValueMemberL{Value: commentsList},
				":ts":       &types.AttributeValueMemberN{Value: strconv.FormatInt(time.Now().Unix(), 10)},
				":uBy":      &types.AttributeValueMemberS{Value: uBy},
			},
			ExpressionAttributeNames: map[string]string{
				"#val":      "value",
				"#Comments": "Comments",
			},
		},
	)
	return err
}

func GetActiveEnquiry(college string, enqId string) (EnquiryData, error) {
	pKey, err := attributevalue.Marshal(enqId)
	if err != nil {
		return EnquiryData{}, errors.New("invalid key")
	}
	key := map[string]types.AttributeValue{
		"uid":  pKey,
		"user": pKey,
	}
	item, err := cfg_details.DynamoCfg.GetItem(context.TODO(), &dynamodb.GetItemInput{
		Key:       key,
		TableName: aws.String(college + ENQ_TABLE_NAME),
	})
	if err != nil {
		return EnquiryData{}, err
	}
	var eqFrm EnquiryData
	err = attributevalue.UnmarshalMap(item.Item, &eqFrm)
	if err != nil {
		return EnquiryData{}, errors.New(cfg_details.DATA_ERROR + err.Error())
	}
	return eqFrm, nil
}

func GetActiveEnquiries(college string, nextPage string) ([]EnquiryForm, error) {
	keyConditions := aws.String("uid = :hashKey")
	expressionAttributeValues := map[string]types.AttributeValue{
		":hashKey": &types.AttributeValueMemberS{Value: ACTIVE_ENQ},
	}
	expressionAttributeNames := map[string]string{
		"#val":  "value",
		"#user": "user",
		"#Name": "Name",
	}
	cols := aws.String("uid,#user,#val.#Name,#val.Mobile,#val.Course,#val.CouncillorName")
	if nextPage != "" {
		keyConditions = aws.String(*keyConditions + " AND #user > :nextKey")
		expressionAttributeValues[":nextKey"] = &types.AttributeValueMemberS{Value: nextPage}
	}
	query := &dynamodb.QueryInput{
		TableName:                 aws.String(college + ENQ_TABLE_NAME),
		Limit:                     aws.Int32(250),
		KeyConditionExpression:    keyConditions,
		ProjectionExpression:      cols,
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
		ScanIndexForward:          aws.Bool(true),
	}
	res, err := cfg_details.DynamoCfg.Query(context.TODO(), query)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, errors.New(cfg_details.DATA_ERROR + err.Error())
	}
	itemsTmp := res.Items
	items := make([]EnquiryForm, len(itemsTmp))
	for i, item := range itemsTmp {
		err := attributevalue.Unmarshal(item["value"], &items[i])
		if err != nil {
			return nil, errors.New(cfg_details.DATA_ERROR + err.Error())
		}
	}
	return items, nil
}

func CreateEnquiry(college string, eqFrm *EnquiryForm, uBy string) error {
	uId := cfg_details.GenerateUlid(time.Now())
	err := createNewEnquiry(uId, college, eqFrm, uBy)
	if err != nil {
		return err
	}
	err = updateActiveEnqDirectory(college, eqFrm, uId, uBy)
	if err != nil {
		return err
	}
	return nil
}

func createNewEnquiry(uId string, college string, frm *EnquiryForm, uBy string) error {
	onF := enquiryFormOnboard{
		Uid:     uId,
		User:    uId,
		Value:   *frm,
		Ts:      time.Now().Unix(),
		Updater: uBy,
	}
	data, err := attributevalue.MarshalMap(onF)
	if err != nil {
		return errors.New("Code error - " + err.Error())
	}
	_, err = cfg_details.DynamoCfg.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName:           aws.String(college + ENQ_TABLE_NAME),
		Item:                data,
		ConditionExpression: aws.String("uid <> :pk AND #user <> :row"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":  &types.AttributeValueMemberS{Value: onF.Uid},
			":row": &types.AttributeValueMemberS{Value: onF.User},
		},
		ExpressionAttributeNames: map[string]string{
			"#user": "user",
		},
	})
	if err != nil {
		return errors.New("DP error - " + err.Error())
	}
	return err
}

func updateActiveEnqDirectory(college string, enqFm *EnquiryForm, uId string, uBy string) error {
	data, err := attributevalue.MarshalMap(enquiryFormOnboard{
		Uid:     "active",
		User:    uId,
		Ts:      time.Now().Unix(),
		Updater: uBy,
		Value:   *enqFm,
	})
	if err != nil {
		return errors.New("Code error - " + err.Error())
	}
	_, err = cfg_details.DynamoCfg.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName:           aws.String(college + ENQ_TABLE_NAME),
		ConditionExpression: aws.String("uid <> :pk AND #user <> :row"),
		Item:                data,
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":  &types.AttributeValueMemberS{Value: "active"},
			":row": &types.AttributeValueMemberS{Value: uId},
		},
		ExpressionAttributeNames: map[string]string{
			"#user": "user",
		},
	})
	if err != nil {
		return errors.New("DP error Active Directory - " + err.Error())
	}
	return nil
}
