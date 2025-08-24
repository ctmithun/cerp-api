package u_by_service

import (
	"cerpApi/cfg_details"
	"log"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func GetUpdatedBy(colId string, ids map[string]string) {
	table := cfg_details.GetPk(colId, "faculty")
	err := getUByItems(table, table, ids)
	if err != nil {
		log.Fatalf("Error fetching items: %v", err)
	}
}

type UserInfo struct {
	Name string `json:"name"`
}

func getUByItems(tableName, pk string, idMaps map[string]string) error {
	keys := make([]map[string]types.AttributeValue, len(idMaps))
	i := 0
	for k, _ := range idMaps {
		keys[i] = map[string]types.AttributeValue{
			"key":   &types.AttributeValueMemberS{Value: pk},
			"email": &types.AttributeValueMemberS{Value: k},
		}
		i++
	}

	input := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			tableName: {
				Keys: keys,
			},
		},
	}
	ctx, cancel := cfg_details.GetTimeoutContext()
	defer cancel()
	resp, err := cfg_details.DynamoCfg.BatchGetItem(ctx, input)
	if err != nil {
		return err
	}
	// Process results
	for _, item := range resp.Responses[tableName] {
		var updater UserInfo
		err := attributevalue.Unmarshal(item["value"], &updater)
		if err != nil {
			log.Printf("Error in unmarshal %v\n", err)
			continue
		}
		emailAttr, ok := item["email"].(*types.AttributeValueMemberS)
		if !ok {
			log.Println("email is not a string")
			continue
		}

		idMaps[emailAttr.Value] = updater.Name
	}

	return nil
}
