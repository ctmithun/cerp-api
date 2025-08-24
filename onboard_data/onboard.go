package onboard_data

import (
	"cerpApi/cfg_details"
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const VAULT_META = "vault-meta"

type onboardS2SData struct {
	PK      string            `dynamodbav:"key"`
	SK      string            `dynamodbav:"skey"`
	Value   map[string]string `dynamodbav:"value"`
	Ts      int64             `dynamodbav:"ts"`
	Updater string            `dynamodbav:"uBy"`
}

func GetS2SPerSub(colId string, course string, batch string, cs string, sub string) (map[string]interface{}, error) {
	key, err := attributevalue.Marshal(getS2SKeyStr(colId, course, batch, cs))
	if err != nil {
		return nil, err
	}
	sKey, err := attributevalue.Marshal(sub)
	if err != nil {
		log.Printf("Error in GetS2SPerSub while marshaling - %v\n", err)
		return nil, err
	}
	ck := map[string]types.AttributeValue{
		"key":  key,
		"skey": sKey,
	}
	out, err := cfg_details.DynamoCfg.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName:            aws.String("college_metadata"),
		Key:                  ck,
		ProjectionExpression: aws.String("#val"),
		ExpressionAttributeNames: map[string]string{
			"#val": "value",
		},
	})
	if out == nil || len(out.Item) == 0 {
		return nil, nil
	}
	item := out.Item["value"].(*types.AttributeValueMemberM)
	studentsData := item.Value["students"].(*types.AttributeValueMemberS)
	if studentsData.Value == "" {
		return nil, nil
	}
	mapStudentsData := make(map[string]interface{})
	for _, s := range strings.Split(studentsData.Value, ",") {
		mapStudentsData[s] = nil
	}
	return mapStudentsData, nil
}

func GetS2S(colId string, course string, batch string, cs string) (string, error) {
	resp, err := cfg_details.DynamoCfg.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(cfg_details.COLLEGE_METADATA_TABLE),
		KeyConditionExpression: aws.String("#key = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: getS2SKeyStr(colId, course, batch, cs)},
		},
		ProjectionExpression: aws.String("skey, #val"),
		ExpressionAttributeNames: map[string]string{
			"#val": "value",
			"#key": "key",
		},
	})
	if err != nil {
		log.Printf("Query Error For GetS2S - %v\n", err)
		return "", err
	}
	s2sMaps := make(map[string]string)
	err = attributevalue.UnmarshalListOfMaps(resp.Items, &s2sMaps)
	for _, v := range resp.Items {
		sKey, _ := v["skey"].(*types.AttributeValueMemberS)
		val, _ := v["value"].(types.AttributeValue)
		val2 := val.(*types.AttributeValueMemberM)
		val3 := val2.Value["students"].(*types.AttributeValueMemberS)
		s2sMaps[sKey.Value] = val3.Value
	}
	respStr, err := json.Marshal(s2sMaps)
	if err != nil {
		log.Printf("Marshal Error For GetS2S - %v\n", err)
		return "", err
	}
	return string(respStr), err
}

func OnboardS2S(colId string, course string, batch string, cs string, uBy string, reqBody []byte) (string, error) {
	reqBodyParsed := make(map[string]string)
	err := json.Unmarshal(reqBody, &reqBodyParsed)
	if err != nil {
		log.Printf("OnboardS2S: unmarshal reqBody err: %v", err)
		return "", err
	}
	return batchWrite(colId, course, batch, cs, uBy, reqBodyParsed)
}

func OnboardVaultMeta(colId string, course string, uBy string, reqBody []byte) (string, error) {
	var writeRequests []types.WriteRequest
	// reqBodyParsed := make(map[string]string)
	// err := json.Unmarshal(reqBody, &reqBodyParsed)
	// if err != nil {
	// 	log.Printf("OnboardVaultMeta: unmarshal reqBody err: %v", err)
	// 	return "", err
	// }
	ts := time.Now().Unix()
	val := map[string]string{VAULT_META: string(reqBody)}
	dataItem := onboardS2SData{
		PK:      cfg_details.GetPk(colId, course),
		SK:      VAULT_META,
		Value:   val,
		Ts:      ts,
		Updater: uBy,
	}
	item, err := attributevalue.MarshalMap(dataItem)
	if err != nil {
		log.Printf("Error marshalling map OnboardVaultMeta %v\n", err)
		return "", err
	}
	writeRequests = append(writeRequests, types.WriteRequest{
		PutRequest: &types.PutRequest{Item: item},
	})
	_, err = cfg_details.DynamoCfg.BatchWriteItem(context.TODO(), &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			cfg_details.COLLEGE_METADATA_TABLE: writeRequests,
		},
	})
	if err != nil {
		log.Printf("Failed to batch write items in OnboardVaultMeta: %v\n", err)
		return "", err
	}

	return "", nil
}

func GetVaultMeta(colId string, course string) (string, error) {
	key, err := attributevalue.Marshal(cfg_details.GetPk(colId, course))
	if err != nil {
		return "", err
	}
	sKey, err := attributevalue.Marshal(VAULT_META)
	if err != nil {
		log.Printf("Error in GetVaultMeta while marshaling - %v\n", err)
		return "", err
	}
	ck := map[string]types.AttributeValue{
		"key":  key,
		"skey": sKey,
	}
	out, err := cfg_details.DynamoCfg.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName:            aws.String(cfg_details.COLLEGE_METADATA_TABLE),
		Key:                  ck,
		ProjectionExpression: aws.String("#val"),
		ExpressionAttributeNames: map[string]string{
			"#val": "value",
		},
	})
	if err != nil {
		log.Printf("Error in fetching data from Dynamo - %v\n", err)
		return "", err
	}
	if out == nil || len(out.Item) == 0 {
		return "", nil
	}
	item := out.Item["value"].(*types.AttributeValueMemberM)
	log.Println(item)
	vaultData := item.Value[VAULT_META].(*types.AttributeValueMemberS)
	if vaultData.Value == "" {
		return "", nil
	}
	return vaultData.Value, nil
}

func batchWrite(colId string, course string, batch string, cs string, uBy string, s2sData map[string]string) (string, error) {
	var writeRequests []types.WriteRequest
	ts := time.Now().Unix()
	for k, v := range s2sData {
		val := map[string]string{"students": v}
		dataItem := onboardS2SData{
			PK:      getS2SKeyStr(colId, course, batch, cs),
			SK:      k,
			Value:   val,
			Ts:      ts,
			Updater: uBy,
		}
		item, err := attributevalue.MarshalMap(dataItem)
		if err != nil {
			log.Printf("Error marshalling map onboardS2SData %v\n", err)
			return "", err
		}
		writeRequests = append(writeRequests, types.WriteRequest{
			PutRequest: &types.PutRequest{Item: item},
		})
	}
	_, err := cfg_details.DynamoCfg.BatchWriteItem(context.TODO(), &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			cfg_details.COLLEGE_METADATA_TABLE: writeRequests,
		},
	})

	if err != nil {
		log.Printf("Failed to batch write items: %v\n", err)
		return "", err
	}
	return strconv.FormatInt(ts, 10), err
}

func getS2SKeyStr(colId string, course string, batch string, cs string) string {
	return cfg_details.GetPk(strings.ToLower(colId), strings.ToLower(course), strings.ToLower(batch), strings.ToLower(cs))
}
