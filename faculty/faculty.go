package faculty

import (
	"cerpApi/cfg_details"
	"cerpApi/iam"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"strings"
	"sync"
	"time"
)

var CFG, _ = config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-south-1"))

// var CFG, _ = config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("mumbai"), config.WithRegion("ap-south-1"))

type Faculty struct {
	Email       string `json:"email"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Doj         string `json:"doj"`
	Subjects    string `json:"subjects"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Roles       string `json:"roles"`
	Designation string `json:"designation"`
}

type OnboardFacultyMetadata struct {
	PK      string  `dynamodbav:"key"`
	SK      string  `dynamodbav:"email"`
	Value   Faculty `dynamodbav:"value"`
	Ts      int64   `dynamodbav:"ts"`
	Updater string  `dynamodbav:"uBy"`
}

func CreateFacultyMeta(college string, facultyData Faculty, uBy string) (bool, string) {
	uId := cfg_details.GenerateUserId(getFacultyIdKey(college, facultyData))
	user := iam.CreateAuth0User(mapFacultyData(facultyData, uId, college))
	fmt.Println("User created by the id - ", user)
	if user == "" {
		return false, uId
	}
	roles := strings.Split(facultyData.Roles, ",")
	var wg sync.WaitGroup
	wg.Add(1)
	var err error
	go func() {
		defer wg.Done()
		marshaledRoleTmp, err := json.Marshal(getRoles(roles))
		if err != nil {
			fmt.Printf("Error Marshaling the allRoles - %v", err)
			return
		}
		iam.SetUserRoles(user, &err, marshaledRoleTmp)
		fmt.Println("User roles updated")
	}()
	ok, err := updateFacultyData(college, facultyData, user, uBy)
	if !ok {
		return false, uId
	}
	fmt.Printf("Updated table and waiting for rolesSet...\n")
	wg.Wait()
	if err != nil {
		return false, uId
	}
	return true, uId
}

func getRoles(roles []string) map[string][]string {
	allRoles := make(map[string][]string)
	allRoles["roles"] = make([]string, 0)
	for _, role := range roles {
		switch role {
		case "faculty":
			allRoles["roles"] = append(allRoles["roles"], cfg_details.FACULTY_ROLE)
		case "counselor":
			allRoles["roles"] = append(allRoles["roles"], cfg_details.COUNSELOR_ROLE)
		case "admin":
			allRoles["roles"] = append(allRoles["roles"], cfg_details.ADMIN_ROLE)
		}
	}
	return allRoles
}

func updateFacultyData(college string, facultyData Faculty, user string, uBy string) (bool, error) {
	PKKey := user
	SKKey := facultyData.Email
	onF := OnboardFacultyMetadata{
		PK:      PKKey,
		SK:      SKKey,
		Value:   facultyData,
		Ts:      time.Now().Unix(),
		Updater: uBy,
	}
	facultyData.Id = PKKey
	data, err := attributevalue.MarshalMap(onF)
	_, err = cfg_details.DynamoCfg.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(college + "_faculty"),
		Item:      data,
	})
	if err != nil {
		fmt.Println("Error putting faculty data - ", err)
		return false, err
	}
	facultyBasicData := Faculty{
		Name: facultyData.Name,
	}
	allFacultyData := OnboardFacultyMetadata{
		PK:    college + "_faculty",
		SK:    user,
		Value: facultyBasicData,
	}
	allFacultyMarshalData, err := attributevalue.MarshalMap(allFacultyData)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	_, err = cfg_details.DynamoCfg.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(college + "_faculty"),
		Item:      allFacultyMarshalData,
	})
	if err != nil {
		fmt.Printf("Updating table failed...\n")
		return false, err
	}
	return true, nil
}

func mapFacultyData(data Faculty, uId string, college string) map[string]interface{} {
	body := map[string]interface{}{
		"email":          data.Email,
		"phone_number":   "+91" + data.PhoneNumber,
		"blocked":        false,
		"email_verified": false,
		"phone_verified": false,
		"name":           data.Name,
		"nickname":       data.Name,
		"user_id":        college + "|" + uId,
		"connection":     "Username-Password-Authentication",
		"verify_email":   true,
	}
	return body
}

func GetFacultyAssignedSubjects(colId string, userId string) string {
	data, err := cfg_details.DynamoCfg.Query(context.TODO(), &dynamodb.QueryInput{
		TableName: aws.String(colId + "_faculty"),
		Limit:     aws.Int32(1),
		ExpressionAttributeNames: map[string]string{
			"#pk": "key",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":hashKey": &types.AttributeValueMemberS{Value: userId},
		},
		KeyConditionExpression: aws.String("#pk = :hashKey"),
		ScanIndexForward:       aws.Bool(false),
	})
	if err != nil {
		fmt.Println("Error in fetching data - ", err)
		return ""
	}
	items := data.Items
	if len(items) == 0 {
		fmt.Println("No data found - ", userId)
		return ""
	}
	item := items[0]["value"]
	var res map[string]string
	_ = attributevalue.Unmarshal(item, &res)
	return fmt.Sprintf("%s", res["Subjects"])
}

func GetFacultiesData(college string, facultyId string) string {
	keyConditions := aws.String("#pk = :hashKey")
	expressionAttributeValues := map[string]types.AttributeValue{
		":hashKey": &types.AttributeValueMemberS{Value: college + "_faculty"},
	}
	expressionAttributeNames := map[string]string{
		"#pk":  "key",
		"#val": "value",
	}
	if facultyId != "" {
		keyConditions = aws.String("#pk = :hashKey")
		expressionAttributeValues[":hashKey"] = &types.AttributeValueMemberS{Value: facultyId}
	}
	cols := aws.String("email,#val")
	data, err := cfg_details.DynamoCfg.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:                 aws.String(college + "_faculty"),
		Limit:                     aws.Int32(50),
		KeyConditionExpression:    keyConditions,
		ProjectionExpression:      cols,
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
		ScanIndexForward:          aws.Bool(false),
	})
	if err != nil {
		fmt.Println(err)
	}
	res := make([]Faculty, len(data.Items))
	for i := 0; i < len(data.Items); i++ {
		var item Faculty
		err = attributevalue.Unmarshal(data.Items[i]["value"], &item)
		if facultyId != "" {
			err = attributevalue.Unmarshal(data.Items[i]["email"], &item.Email)
			item.Id = facultyId
		} else {
			err = attributevalue.Unmarshal(data.Items[i]["email"], &item.Id)
		}

		if err == nil {
			res[i] = item
		}
	}
	finalRes, err := json.Marshal(res)
	if err != nil {
		return ""
	}
	return string(finalRes)
}

func ModifyFacultyData(college string, facultyForm Faculty, uBy string, isRoleUpdate bool) (bool, string) {
	user := cfg_details.GenerateUserId(getFacultyIdKey(college, facultyForm))
	if !strings.Contains(facultyForm.Id, user) {
		return false, cfg_details.INVALID_DATA
	}
	_, err := updateFacultyData(college, facultyForm, facultyForm.Id, uBy)
	if isRoleUpdate {
		marshaledRoleTmp, err := json.Marshal(getRoles(strings.Split(facultyForm.Roles, ",")))
		if err != nil {
			fmt.Printf("Error Marshaling the allRoles - %v", err)
			return false, cfg_details.CODE_ERROR + err.Error()
		}
		iam.SetUserRoles(facultyForm.Id, &err, marshaledRoleTmp)
	}
	if err != nil {
		fmt.Println("Failed updating on the user - ", user)
		return false, facultyForm.Email
	}
	return true, facultyForm.Email
}

//func UploadFacultyFiles()

func DeactivateFaculty(college string, facultyForm Faculty, uBy string) (bool, string) {
	user := cfg_details.GenerateUserId(getFacultyIdKey(college, facultyForm))
	if !strings.Contains(facultyForm.Id, user) {
		return false, cfg_details.INVALID_DATA
	}
	err := iam.DeleteUser(facultyForm.Id, uBy)
	if err != nil {
		return false, err.Error()
	}
	return true, ""
}

func DeleteFaculty(college string, facultyForm Faculty, uBy string) (bool, string) {
	user := cfg_details.GenerateUserId(getFacultyIdKey(college, facultyForm))
	if !strings.Contains(facultyForm.Id, user) {
		return false, cfg_details.INVALID_DATA
	}
	err := iam.DeleteUser(facultyForm.Id, uBy)
	if err != nil {
		return false, err.Error()
	}
	item, err := cfg_details.DynamoCfg.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(college + "_faculty"),
		Key: map[string]types.AttributeValue{
			"key":   &types.AttributeValueMemberS{Value: facultyForm.Id},
			"email": &types.AttributeValueMemberS{Value: facultyForm.Email},
		},
	})
	if err != nil || item == nil {
		fmt.Println("Error in deactivating faculty from the table - ", facultyForm.Id)
		return false, ""
	}
	item, err = cfg_details.DynamoCfg.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(college + "_faculty"),
		Key: map[string]types.AttributeValue{
			"key":   &types.AttributeValueMemberS{Value: college + "_faculty"},
			"email": &types.AttributeValueMemberS{Value: facultyForm.Id},
		},
	})
	if err != nil || item == nil {
		fmt.Printf("Error in deactivating faculty from the table for %s_faculty partition key - %s\n", college, facultyForm.Id)
		return false, ""
	}
	return true, ""

}

func getFacultyIdKey(college string, facultyForm Faculty) string {
	return "F-" + college + "|" + facultyForm.Email + "_" + facultyForm.PhoneNumber
}
