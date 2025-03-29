package subject

import (
	"cerpApi/cfg_details"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log"
	"strings"
)

func GetSubjects(collegeId string, class string, batch string) map[string]string {
	key, err := attributevalue.Marshal(collegeId + "_" + strings.ToLower(class))
	sKey, err := attributevalue.Marshal(batch)
	if err != nil {
		log.Printf("Error Fetching subjects %v\n", err)
		return nil
	}
	ck := map[string]types.AttributeValue{
		"key":  key,
		"skey": sKey,
	}
	fmt.Printf("key %s, skey %s\n", key, sKey)
	out, err := cfg_details.DynamoCfg.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String("college_metadata"),
		Key:       ck,
	})
	item := out.Item["subjects"]
	var res map[string]string
	err = attributevalue.Unmarshal(item, &res)
	if err != nil {
		log.Printf("Error Fetching subjects unmarshalling %v\n", err)
		return nil
	}
	return res
}
