package faculty

import (
	"bytes"
	"cerpApi/cfg_details"
	"cerpApi/iam"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
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
	err := iam.DeactivateUser(facultyForm.Id, uBy)
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
	// item, err := cfg_details.DynamoCfg.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
	// 	TableName: aws.String(college + "_faculty"),
	// 	Key: map[string]types.AttributeValue{
	// 		"key":   &types.AttributeValueMemberS{Value: facultyForm.Id},
	// 		"email": &types.AttributeValueMemberS{Value: facultyForm.Email},
	// 	},
	// })
	// if err != nil || item == nil {
	// 	fmt.Println("Error in deactivating faculty from the table - ", facultyForm.Id)
	// 	return false, ""
	// }
	item, err := cfg_details.DynamoCfg.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
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

func DeleteFacultyFile(s3Client *s3.Client, colId string, fId string, fileKey string, uBy string) (string, error) {
	tableName := colId + "_files"
	key := map[string]types.AttributeValue{
		"uid": &types.AttributeValueMemberS{Value: fId}, // Change as needed
	}
	setVal := " SET"
	expr := map[string]types.AttributeValue{}
	setVal = setVal + " ts = :ts, uBy = :uBy"
	expr[":ts"] = &types.AttributeValueMemberN{Value: fmt.Sprint(time.Now().Unix())}
	expr[":uBy"] = &types.AttributeValueMemberS{Value: uBy}
	aliasFileKey := strings.ReplaceAll(fileKey, ".", "")
	exprNames := map[string]string{
		"#fil":             "values",
		"#" + aliasFileKey: fileKey,
	}
	updateInput := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(tableName),
		Key:                       key,
		UpdateExpression:          aws.String("REMOVE #fil.#" + aliasFileKey + setVal),
		ExpressionAttributeNames:  exprNames,
		ExpressionAttributeValues: expr,
	}
	_, err := cfg_details.DynamoCfg.UpdateItem(context.TODO(), updateInput)
	if err != nil {
		log.Printf("Error removing the key from dynamo %s %v\n", fileKey, err)
		return "", err
	}
	err = RemoveFacultyFileFromS3(s3Client, colId, fId, fileKey)
	if err != nil {
		log.Printf("Error Removing file from S3 %s %s %v\n", fileKey, fId, err)
		return "Not Removed", err
	}
	return "Removed!", nil
}

func UpdateFileMeta(colId string, formMap map[string]string, fId string, uBy string) (string, error) {
	tableName := colId + "_files"
	key := map[string]types.AttributeValue{
		"uid": &types.AttributeValueMemberS{Value: fId}, // Change as needed
	}
	setVal := "SET"
	expr := map[string]types.AttributeValue{}
	exprNames := map[string]string{
		"#fil": "values",
	}
	for k, v := range formMap {
		if setVal != "SET" {
			setVal = setVal + ","
		}
		aliasKey := strings.ReplaceAll(k, ".", "")
		setVal = setVal + " #fil.#" + aliasKey + " = :" + aliasKey
		expr[":"+aliasKey] = &types.AttributeValueMemberS{Value: v}
		exprNames["#"+aliasKey] = k
	}
	setVal = setVal + ", ts = :ts, uBy = :uBy"
	expr[":ts"] = &types.AttributeValueMemberN{Value: fmt.Sprint(time.Now().Unix())}
	expr[":uBy"] = &types.AttributeValueMemberS{Value: uBy}
	ind := 0
	for {
		updateInput := &dynamodb.UpdateItemInput{
			TableName:                 aws.String(tableName),
			Key:                       key,
			UpdateExpression:          aws.String(setVal),
			ExpressionAttributeNames:  exprNames,
			ExpressionAttributeValues: expr,
		}
		_, err := cfg_details.DynamoCfg.UpdateItem(context.TODO(), updateInput)
		if err != nil {
			log.Printf("Failed to update item: %v", err)
			err2 := insertItem(colId, fId, uBy)
			if err2 != nil || ind == 2 {
				return err.Error(), err
			}
			ind = ind + 1
			log.Printf("Retrying for the %d times\n", ind+1)
			continue
		} else {
			break
		}
	}
	return "Updated!!", nil
}

type FileUpdater struct {
	PK      string            `dynamodbav:"uid"`
	Values  map[string]string `dynamodbav:"values"`
	Ts      int64             `dynamodbav:"ts"`
	Updater string            `dynamodbav:"uBy"`
}

func insertItem(colId string, uId string, uBy string) error {
	fileBook := FileUpdater{
		PK:      uId,
		Values:  make(map[string]string),
		Ts:      time.Now().Unix(),
		Updater: uBy,
	}
	data, err := attributevalue.MarshalMap(fileBook)
	if err != nil {
		return err
	}
	_, err = cfg_details.DynamoCfg.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(colId + "_files"),
		Item:      data,
	})
	return err
}

func UploadFacultyToS3(s3Client *s3.Client, fileName string, fileData []byte, id string, colId string) error {
	_, err := s3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(cfg_details.BUCKET_STUDENTS_FACULTIES),
		Key:    aws.String(getFacultyS3Key(colId, id, fileName)),
		Body:   bytes.NewReader(fileData),
	})
	if err != nil {
		return fmt.Errorf("failed to upload %s to S3: %w", fileName, err)
	}
	return nil
}

var tags = cfg_details.ExpireTags()

func RemoveFacultyFileFromS3(s3Client *s3.Client, colId string, id string, fileName string) error {
	oldKey := getFacultyS3Key(colId, id, fileName)
	newKey := getFacultyS3Key(colId, id, "rem_"+strconv.FormatInt(time.Now().Unix(), 10)+"_"+fileName)

	// 1. Copy the object to new key
	_, err := s3Client.CopyObject(context.TODO(), &s3.CopyObjectInput{
		Bucket:           aws.String(cfg_details.BUCKET_STUDENTS_FACULTIES),
		CopySource:       aws.String(cfg_details.BUCKET_STUDENTS_FACULTIES + "/" + oldKey),
		Key:              aws.String(newKey),
		Tagging:          aws.String(tags),
		TaggingDirective: s3types.TaggingDirectiveReplace,
	})
	if err != nil {
		log.Printf("Failed to backup the deleting file %s %v\n", oldKey, err)
	}
	_, err = s3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(cfg_details.BUCKET_STUDENTS_FACULTIES),
		Key:    aws.String(oldKey),
	})
	if err != nil {
		log.Printf("failed to Delete key %s to S3: %v\n", oldKey, err)
		return err
	}
	return nil
}

func getFacultyS3Key(colId string, id string, fileName string) string {
	return colId + "/faculty/" + id + "/" + fileName
}

func DownloadFacultyFile(colId string, uId string, fName string, s3Client *s3.Client) (string, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(cfg_details.BUCKET_STUDENTS_FACULTIES),
		Key:    aws.String(getFacultyS3Key(colId, uId, fName)),
	}

	presignedURL, err := cfg_details.Presigner.PresignGetObject(context.TODO(), input, s3.WithPresignExpires(60*time.Second))
	if err != nil {
		log.Printf("Error in dowloading the faculty file for s3 read operation - %s/%s err-%v\n", uId, fName, err)
		return "", err
	}
	enc := url.QueryEscape(presignedURL.URL)
	body, _ := json.Marshal(cfg_details.FileResponse{URL: enc})
	return string(body), err
}

func FetchFilesMetadata(colId string, uId string) (map[string]string, error) {
	ck := map[string]types.AttributeValue{
		"uid": &types.AttributeValueMemberS{Value: uId},
	}
	out, err := cfg_details.DynamoCfg.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String(colId + "_files"),
		Key:       ck,
	})
	if err != nil {
		log.Printf("Error in FetchFilesMetadata while reading data from DDB %v\n", err)
		return nil, err
	}
	item := out.Item["values"]
	var res map[string]string
	err = attributevalue.Unmarshal(item, &res)
	if err != nil {
		log.Printf("Error in FetchFilesMetadata while unmarshaling the data from DDB %v\n", err)
		return nil, err
	}
	return res, nil
}
