package faculty

import (
	"bytes"
	"cerpApi/cfg_details"
	"cerpApi/psw_generator"
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/oklog/ulid/v2"
	"io"
	"net/http"
	"time"
)

var CFG, _ = config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-south-1"))
var DynamoCfg = dynamodb.NewFromConfig(CFG)

var roleBody = map[string][]string{
	"roles": {cfg_details.FACULTY_ROLE},
}
var marshalledRole, _ = json.Marshal(roleBody)

type Faculty struct {
	Email       string `json:"email"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Doj         string `json:"doj"`
	Subjects    string `json:"subjects"`
	Description string `json:"description"`
}

type OnboardFacultyMetadata struct {
	PK      string `dynamodbav:"key"`
	SK      string `dynamodbav:"skey"`
	Value   string `dynamodbav:"value"`
	Ts      int64  `dynamodbav:"ts"`
	Updater string `dynamodbav:"uBy"`
}

func CreateFacultyMeta(college string, facultyData Faculty, userId string) (bool, string) {
	uId := generateUserId(college)
	user := createAuth0User(facultyData, uId, college)
	if user == "" {
		return false, uId
	}
	err := setUserRoles(uId)
	if err != nil {
		return false, uId
	}
	PKKey := college + "_" + "faculty"
	SKKey := facultyData.Email
	val := make(map[string]string)
	val["email"] = facultyData.Email
	val["name"] = facultyData.Name
	inputVal := map[string]string{
		"id":   uId,
		"name": facultyData.Name,
	}
	valByte, err := json.Marshal(inputVal)
	if err != nil {
		return false, uId
	}
	onF := OnboardFacultyMetadata{
		PK:      PKKey,
		SK:      SKKey,
		Value:   string(valByte),
		Ts:      time.Now().UTC().Unix(),
		Updater: userId,
	}
	data, err := attributevalue.MarshalMap(onF)
	_, err = DynamoCfg.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(college + "_faculty"),
		Item:      data,
	})

	if err != nil {
		return false, uId
	}
	return true, ""
}

func setUserRoles(uId string) error {
	url := cfg_details.API_URL + "/users/" + uId + "/roles"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(marshalledRole))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg_details.TOKEN)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resBody, err := io.ReadAll(res.Body)
	respMap := make(map[string]interface{})
	err = json.Unmarshal(resBody, &respMap)
	if val, ok := respMap["statusCode"]; err != nil || (ok && val.(int) > 300) {
		return errors.New("Roles setting failed, Try after sometime")
	}
	return nil
}

func generateUserId(col string) string {
	ulId := ulid.Make()
	return ulId.String()
}

func createAuth0User(data Faculty, uId string, college string) string {
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
		"password":       generateRandomPsw(),
		"verify_email":   true,
	}
	url := cfg_details.API_URL + "/users"
	marshalled, err := json.Marshal(body)
	if err != nil {
		return ""
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(marshalled))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg_details.TOKEN)
	if err != nil {
		return ""
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	resBody, err := io.ReadAll(res.Body)
	respMap := make(map[string]interface{})
	err = json.Unmarshal(resBody, &respMap)
	if val, ok := respMap["statusCode"]; err != nil || (ok && val.(int) > 300) {
		return ""
	}
	return respMap["user_id"].(string)
}

func generateRandomPsw() string {
	return psw_generator.GeneratePsw()
}

//func GetFaculties(college string) ([]faculty, error) {
//	return []faculty{}, nil
//key, err := attributevalue.Marshal(college)
//if err != nil {
//	return nil, err
//}
//ck := map[string]types.AttributeValue{
//	"key":  key,
//	"skey": sKey,
//}
//out, err := DynamoCfg.GetItem(context.Background(), &dynamodb.GetItemInput{
//	TableName: aws.String("college_metadata"),
//	Key:       ck,
//})
//fmt.Println(out)
//item := out.Item["students"]
//if item == nil {
//	return nil, nil
//}
//fmt.Println(item)
//var res string
//err = attributevalue.Unmarshal(item, &res)
//var parsedRes []student
//err = json.Unmarshal([]byte(res), &parsedRes)
//return parsedRes, err
//}
