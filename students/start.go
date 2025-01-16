package students

import (
	"cerpApi/cfg_details"
	"cerpApi/iam"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Student struct {
	Email                string `json:"email"`
	Id                   string `json:"id"`
	Name                 string `json:"name"`
	PhoneNumber          string `json:"phone_number"`
	Doj                  string `json:"doj"`
	Sid                  string `json:"sid"`
	Batch                string `json:"batch"`
	Stream               string `json:"class"`
	Fees                 int    `json:"fees"`
	Type                 string `json:"type"`
	UniversitySeatNumber string `json:"university_seat_number"`
}

type OnboardStudentBasicData struct {
	BatchYear string  `dynamodbav:"pk"`
	SK        int     `dynamodbav:"row_num"`
	Sid       string  `dynamodbav:"student_id"`
	Value     Student `dynamodbav:"value"`
	Ts        int64   `dynamodbav:"ts"`
	Updater   string  `dynamodbav:"uBy"`
}

var roleBody = map[string][]string{
	"roles": {cfg_details.STUDENT_ROLE},
}
var marshalledRole, _ = json.Marshal(roleBody)

func getStudentData(data *Student, uId string, college string) map[string]interface{} {
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

func getStudentIdKey(college string, student Student) string {
	return "S-" + college + "|" + student.Email + "_" + student.PhoneNumber
}

func OnboardStudent(college string, student *Student, uBy string) (string, error) {
	uId := cfg_details.GenerateUserId(getStudentIdKey(college, *student))
	PKKey := student.Batch + "-" + student.Stream
	SKKey := getRowNumber(PKKey, college)
	user := iam.CreateAuth0User(getStudentData(student, uId, college))
	fmt.Println("User created by the id - ", user)
	student.Id = user
	if user == "" {
		return "", errors.New("user creation failed")
	}
	var wg sync.WaitGroup
	wg.Add(1)
	var err error
	go func() {
		defer wg.Done()
		iam.SetUserRoles(user, &err, marshalledRole)
		fmt.Println("User roles updated")
	}()
	//inputVal := structToMap(student)
	student.Sid = PKKey + "-" + strconv.Itoa(SKKey)
	err = persistStudentRecord(college, student, PKKey, SKKey, uBy, false)
	wg.Wait()
	return student.Sid, err
}

func persistStudentRecord(college string, student *Student, PKKey string, SKKey int, uBy string, isUpdate bool) error {
	onF := OnboardStudentBasicData{
		BatchYear: PKKey,
		SK:        SKKey,
		Sid:       student.Sid,
		Value:     *student,
		Ts:        time.Now().UTC().Unix(),
		Updater:   uBy,
	}
	data, err := attributevalue.MarshalMap(onF)
	if !isUpdate {
		_, err = cfg_details.DynamoCfg.PutItem(context.TODO(), &dynamodb.PutItemInput{
			TableName:           aws.String(college + "_students"),
			Item:                data,
			ConditionExpression: aws.String("pk <> :pk AND row_num <> :row"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk":  &types.AttributeValueMemberS{Value: PKKey},
				":row": &types.AttributeValueMemberN{Value: strconv.Itoa(SKKey)},
			},
		})
	} else {
		_, err = cfg_details.DynamoCfg.UpdateItem(
			context.TODO(),
			&dynamodb.UpdateItemInput{
				TableName: aws.String(college + "_students"),
				Key: map[string]types.AttributeValue{
					"pk":      &types.AttributeValueMemberS{Value: PKKey},
					"row_num": &types.AttributeValueMemberN{Value: strconv.Itoa(SKKey)},
				},
				UpdateExpression: aws.String("SET #val.#Fees = :Fees, #val.#Name = :Name, #val.#Doj = :Doj, ts = :ts, uBy = :uBy"),
				//ConditionExpression: aws.String("pk <> :pk AND row_num <> :row"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":Fees": &types.AttributeValueMemberN{Value: strconv.Itoa(student.Fees)},
					":Name": &types.AttributeValueMemberS{Value: student.Name},
					":Doj":  &types.AttributeValueMemberS{Value: student.Doj},
					":ts":   &types.AttributeValueMemberN{Value: strconv.FormatInt(time.Now().UTC().Unix(), 10)},
					":uBy":  &types.AttributeValueMemberS{Value: uBy},
				},
				ExpressionAttributeNames: map[string]string{
					"#val":  "value",
					"#Name": "Name",
					"#Doj":  "Doj",
					"#Fees": "Fees",
				},
			},
		)
	}
	if err != nil {
		fmt.Printf("Student onboard Failed...\n")
		return err
	}
	fmt.Printf("Updated table and waiting for rolesSet...\n")
	return err
}

func GetRowNumber(pKey string, college string) int {
	return getRowNumber(pKey, college)
}

func getRowNumber(pKey string, college string) int {

	data, err := cfg_details.DynamoCfg.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(college + "_students"),
		Limit:                  aws.Int32(1),
		KeyConditionExpression: aws.String("pk = :hashKey"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":hashKey": &types.AttributeValueMemberS{Value: pKey},
		},
		ScanIndexForward: aws.Bool(false),
	})
	if err != nil {
		fmt.Println(err)
	}
	if data.Count == 0 {
		return 1
	}
	a, err := strconv.Atoi(data.Items[0]["row_num"].(*types.AttributeValueMemberN).Value)
	if err != nil {
		return 0
	}
	nextRow := a + 1
	return nextRow
}

func GetStudentsData(college string, batch string, stream string) []Student {
	keyConditions := aws.String("pk = :hashKey")
	pk := batch + "-" + stream
	expressionAttributeValues := map[string]types.AttributeValue{
		":hashKey": &types.AttributeValueMemberS{Value: pk},
	}
	cols := aws.String("row_num,#val")
	data, err := cfg_details.DynamoCfg.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:                 aws.String(college + "_students"),
		Limit:                     aws.Int32(300),
		KeyConditionExpression:    keyConditions,
		ProjectionExpression:      cols,
		ExpressionAttributeValues: expressionAttributeValues,
		ExpressionAttributeNames: map[string]string{
			"#val": "value",
		},
		ScanIndexForward: aws.Bool(false),
	})
	if err != nil {
		fmt.Println(err)
	}
	if data.Count == 0 {
		return make([]Student, 0)
	}
	res := make([]Student, data.Count)
	items := data.Items
	for i := 0; i < len(items); i++ {
		var student Student
		err = attributevalue.Unmarshal(items[i]["value"], &student)
		res[i] = student
	}
	fmt.Println("Collected data is - ", data)
	return res
}

func UpdateStudentRecord(college string, student *Student, uBy string) (bool, string) {
	user := cfg_details.GenerateUserId(getStudentIdKey(college, *student))
	if !strings.Contains(student.Id, user) {
		return false, cfg_details.INVALID_DATA
	}
	PKKey := student.Batch + "-" + student.Stream
	SKKey, err := extractRowNum(student.Sid)
	if err != nil {
		return false, cfg_details.INVALID_DATA
	}
	err = persistStudentRecord(college, student, PKKey, SKKey, uBy, true)
	if err != nil {
		return false, err.Error()
	}
	return true, ""
}

func extractRowNum(sid string) (int, error) {
	sArr := strings.Split(sid, "-")
	return strconv.Atoi(sArr[len(sArr)-1])
}

func DeactivateStudent(college string, student Student, uBy string) (bool, string) {
	userId := cfg_details.GenerateUserId(getStudentIdKey(college, student))

	if !strings.Contains(student.Id, userId) {
		return false, cfg_details.INVALID_DATA
	}
	fmt.Printf("Deactivating the user %s by %s\n", student.Id, uBy)
	err := iam.DeactivateUser(student.Id, uBy)
	if err != nil {
		fmt.Println("User deactivation failed in Auth0 - ", student.Id)
		return false, cfg_details.AUTH0_UNAVAILABLE
	}
	PKKey := student.Batch + "-" + student.Stream
	SKKey, err := extractRowNum(student.Sid)
	if err != nil {
		fmt.Println("Error deactivating the student - ", student.Id)
		return false, err.Error()
	}
	item, err := cfg_details.DynamoCfg.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(college + "_students"),
		Key: map[string]types.AttributeValue{
			"pk":      &types.AttributeValueMemberS{Value: PKKey},
			"row_num": &types.AttributeValueMemberN{Value: strconv.Itoa(SKKey)},
		},
	})
	if err != nil || item == nil {
		fmt.Println("Error in deactivating student from the table - ", student.Id)
		return false, ""
	}
	return true, ""
}
