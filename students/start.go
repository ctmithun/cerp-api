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
	Email                 string `json:"email"`
	Id                    string `json:"id"`
	Name                  string `json:"name"`
	Mobile                string `json:"mobile"`
	Branch                string `json:"branch"`
	Batch                 string `json:"batch"`
	Year                  string `json:"year"`
	Yoa                   string `json:"yoa"`
	AdmissionCounselor    string `json:"admission_counselor"`
	FreeSeat              string `json:"free_seat"`
	Installments          string `json:"installments"`
	Reference             string `json:"reference"`
	PreviousQualification string `json:"previous_qualification"`
	YearOfPassing         string `json:"year_of_passing"`
	TotalMarks            string `json:"total_marks"`
	SecondLanguage        string `json:"second_lang"`
	StateOfPrevQual       string `json:"state_of_prev_qual"`
	Gender                string `json:"gender"`
	Dob                   string `json:"dob"`
	Nationality           string `json:"nationality"`
	Religion              string `json:"religion"`
	CasteCategory         string `json:"caste_category"`
	GovtScholarship       string `json:"govt_scholarship"`
	AnnualIncome          string `json:"annual_income"`
	BplCard               string `json:"bpl_card"`
	FatherName            string `json:"father_name"`
	FatherOccupation      string `json:"father_occupation"`
	FatherMobile          string `json:"father_mobile"`
	MotherName            string `json:"mother_name"`
	MotherOccupation      string `json:"mother_occupation"`
	MotherMobile          string `json:"mother_mobile"`
	SingleParent          string `json:"single_parent"`
	Aadhar                string `json:"aadhar"`
	Address               string `json:"address"`
	Passport              string `json:"passport"`
	BloodGroup            string `json:"blood_group"`
	Doj                   string `json:"doj"`
	Sid                   string `json:"sid"`
	Fees                  int    `json:"fees"`
	UniversitySeatNumber  string `json:"university_seat_number"`
	Photo                 string `json:"photo"`
	Signature             string `json:"signature"`
	MarksCard             string `json:"other"`
	TotalYearlyFees       string `json:"total_yearly_fees"`
	AdmissionFees         string `json:"admission_fees"`
	FeeReceipt            string `json:"fee_receipt"`
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
		"phone_number":   "+91" + data.Mobile,
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
	return "S-" + college + "|" + student.Email + "_" + student.Mobile
}

//go:deprecated
func OnboardStudent(college string, student *Student, uBy string) (string, error) {
	uId := cfg_details.GenerateUserId(getStudentIdKey(college, *student))
	PKKey := student.Batch + "-" + student.Branch
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
		Ts:        time.Now().Unix(),
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
	}
	if err != nil {
		fmt.Printf("Student onboard Failed...\n")
		return err
	}
	fmt.Printf("Updated table and waiting for rolesSet...\n")
	return err
}

func updateStudentRecord(college string, student *Student, PKKey string, SKKey int, uBy string) error {
	_, err := cfg_details.DynamoCfg.UpdateItem(
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
				":ts":   &types.AttributeValueMemberN{Value: strconv.FormatInt(time.Now().Unix(), 10)},
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
	if err != nil {
		fmt.Printf("Student update failed...\n")
		return err
	}
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
		ScanIndexForward: aws.Bool(true),
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
	PKKey := strings.Split(student.Batch, "-")[0] + "-" + student.Branch
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
	PKKey := student.Batch + "-" + student.Branch
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

func GetStudents(college string, course string, batch string, cs string) (types.AttributeValue, error) {
	key, err := attributevalue.Marshal(college)
	sKey, err := attributevalue.Marshal(strings.ToLower(course) + "_" + strings.ToLower(batch) + "_" + strings.ToLower(cs))
	if err != nil {
		return nil, err
	}
	ck := map[string]types.AttributeValue{
		"key":  key,
		"skey": sKey,
	}
	out, err := cfg_details.DynamoCfg.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String("college_metadata"),
		Key:       ck,
	})
	item := out.Item["students"]
	if item == nil {
		return nil, nil
	}
	return item, err
}
