package main

import (
	"cerpApi/cfg_details"
	"cerpApi/faculty"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MicahParks/keyfunc"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	_ "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/golang-jwt/jwt/v4"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//const (
//	CLAIM_ISS       = "https://dev-q0ywv1av1mdo8z4n.us.auth0.com/"
//	CLAIM_CLIENT_ID = "t4pVw4sPvFWgvq4n3DgvFGnavMatqYwv"
//)

var SECRET = []byte("This is the way -- Mando")

//var CFG, _ = config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("mumbai"), config.WithRegion("ap-south-1"))

var CFG, _ = config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-south-1"))
var DYNAMO_CFG = dynamodb.NewFromConfig(CFG)
var ResponseHeaders = map[string]string{"Access-Control-Allow-Origin": "*", "X-Frame-Options": "SAMEORIGIN", "Strict-Transport-Security": "max-age=31557600; includeSubDomains"}
var AUTH_500 = getProxyResponse("Invalid Input!", 500)
var AUTH_504 = getProxyResponse("Service Unavailable Input!", 504)
var AUTH_404 = getProxyResponse("Service Not Found!", 404)
var AUTH_403 = getProxyResponse("Access Denied!", 403)

func getProxyResponse(body string, statusCode int) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{Body: body, StatusCode: statusCode, IsBase64Encoded: true, Headers: ResponseHeaders}
}

func main() {

	ch := make(chan bool)

	go func() {
		// Do some work
		ch <- false
	}()

	// Wait for the goroutine to finish
	res := <-ch
	print(res)

	//log.Println("Running lambda-authorizer")
	////lambda.Start(handler)
	//req := events.APIGatewayProxyRequest{
	//	Headers: map[string]string{},
	//}
	//req.Headers["Authorization"] = "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IlNwb3ZQcXFBUFVKbWljZjdMRFkxViJ9.eyJpc3MiOiJodHRwczovL2Rldi1xMHl3djFhdjFtZG84ejRuLnVzLmF1dGgwLmNvbS8iLCJzdWIiOiJhdXRoMHw2NzA2OTIyMDk1NGUxYmQ0NTFkMjJiZmMiLCJhdWQiOlsiaHR0cHM6Ly9kZXYtcTB5d3YxYXYxbWRvOHo0bi51cy5hdXRoMC5jb20vYXBpL3YyLyIsImh0dHBzOi8vZGV2LXEweXd2MWF2MW1kbzh6NG4udXMuYXV0aDAuY29tL3VzZXJpbmZvIl0sImlhdCI6MTcyOTUwNTA1NywiZXhwIjoxNzI5NTA2ODU3LCJzY29wZSI6Im9wZW5pZCBwcm9maWxlIGVtYWlsIiwiYXpwIjoidDRwVnc0c1B2RldndnE0bjNEZ3ZGR25hdk1hdHFZd3YifQ.ckL7dyxUqfFEuc83eauc4K3JtjD50UWWWAuq1m-zcSaYaSHfTP-YgTx_WgrG19rr_pWPlS0YIGrRhEvTsvTcmSg7qtsPx66EAIJnazfPa8vUsBqXFBPH5YW7RC-nyVChcxvhf5XwiUK_lr32l4LNVz8DS-v5t4ERUF71Maa6yefaTMdXGvXGD1K2zdhUwMBLNP-r3xudu_BmMX8xBoNwJHBHDqCVWAYVNffY0XEGmC0m3Wb5w8nf4ANxatUSA99Ta_al-W5yrwIbWIql9Rn-hNeAj_0-ZmY3Bj2ZCVnZgaMotdGezhSOlCNJcfPIyLxF3ReeV38IhcSAS19UnRJi5g"
	//_, err := handler(req)
	//if err != nil {
	//	panic(err)
	//}

	//lambda.Start(handler)
	//students, err := getStudents("ni", "BCA", "SEM-1")
	//attendanceStudents, err := getAttendanceStudents("ni", "BCA", "SEM-1", "CA-C1T", "31-10-2024")
	//if err != nil {
	//	return
	//}
	//if err != nil {
	//	return
	//}
	//fmt.Println(attendanceStudents)

	//form, err := getAttendanceForm("ni", "BCA", "SEM-1", "CA-C1T", "30-10-2024")
	//if err != nil {
	//	return
	//}
	//fmt.Println(form)

	//err := updateAttendanceForm("ni", "BCA", "SEM-1", "CA-C1T", "2024-10-28", 23213)
	//if err != nil {
	//	return
	//}
	//OnboardSubjects("ni", "BCA", "SEM-1", "mithun", nil)

	//getStudents("ni", "BCA", "SEM-2")

	data := faculty.Faculty{
		Email:       "pavancis09@gmail.com",
		Id:          "",
		Name:        "Pavan C T",
		PhoneNumber: "9844990216",
		Doj:         "2024-12-9",
		Subjects:    "",
		Description: "New adjunct faculty",
	}
	faculty.CreateFacultyMeta("ni", data, "mithun")

}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	claims, err := verifyToken(request.Headers["Authorization"])
	if err != nil {
		return AUTH_500, err
	}
	userId := claims["sub"].(string)
	//resPath := request.Resource
	fmt.Printf("Resource path requested is %s", claims)
	if request.Resource == "/roles" {
		res := fetchRoles(strings.Split(userId, "|")[1])
		if res == "" {
			return AUTH_403, errors.New("no roles set")
		}
		tkn := getHmacData(int64(int(claims["exp"].(float64))), res, userId)
		return respondData(claims, tkn, err)
	} else if request.Resource == "/userRoles" {
		res := fetchRoles(strings.Split(userId, "|")[1])
		if res == "" {
			return AUTH_500, errors.New("no roles set")
		}
		return respondData(claims, res, err)
	}
	roles := parseHmacToken(request.Headers["cerp-api-token"])
	fmt.Printf("Found roles %s\n", roles)
	if request.Resource == "/attendance/students" {
		if strings.Contains(roles, "faculty") || strings.Contains(roles, "admin") {
			colId := request.QueryStringParameters["college_id"]
			course := request.QueryStringParameters["class"]
			batch := request.QueryStringParameters["batch"]
			subject := request.QueryStringParameters["subject"]
			date := request.QueryStringParameters["date"]
			fmt.Println(colId, course, batch, subject, date)
			res, err := getAttendanceStudents(colId, course, batch, subject, date)
			if err != nil {
				return AUTH_500, err
			}
			return respondData(claims, res, err)
		}
		return AUTH_500, errors.New("no roles set")
	} else if request.Resource == "/subject" {
		fmt.Println("Inisde the subject API...")
		if strings.Contains(roles, "faculty") || strings.Contains(roles, "admin") {
			fmt.Println("Getting the subjects ...")
			res := getSubjects(request.QueryStringParameters["college_id"], request.QueryStringParameters["class"], request.QueryStringParameters["batch"])
			return respondData(claims, res, err)
		}
		return AUTH_500, errors.New("no roles set")
	} else if request.Resource == "/attendance/update" {
		fmt.Println("Inisde the attendance Update API...")
		if strings.Contains(roles, "faculty") || strings.Contains(roles, "admin") {
			colId := request.QueryStringParameters["college_id"]
			course := request.QueryStringParameters["class"]
			batch := request.QueryStringParameters["batch"]
			subject := request.QueryStringParameters["subject"]
			date := request.QueryStringParameters["date"]
			var attendanceForm attendanceForm
			err = json.Unmarshal([]byte(request.Body), &attendanceForm)
			fmt.Println(attendanceForm)
			attendanceForm.UBy = userId
			attendanceForm.Ts = time.Now().UTC().Unix()
			if err != nil {
				return events.APIGatewayProxyResponse{}, err
			}
			err = updateAttendanceForm(colId, course, batch, subject, date, &attendanceForm)
			if err != nil {
				return AUTH_500, err
			}
			return respondData(claims, "Updated at - "+strconv.FormatInt(attendanceForm.Ts, 10), err)
		}
		return AUTH_500, errors.New("no roles set")
	} else if strings.Contains(roles, "admin") {
		if request.Resource == "/metadata/update" {
			fmt.Printf("Inside the /metadata/update")
			_type := request.QueryStringParameters["type"]
			colId := request.QueryStringParameters["college_id"]
			course := strings.ToLower(request.QueryStringParameters["class"])
			batch := request.QueryStringParameters["batch"]
			if _type == "subjects" {
				err, ts := OnboardSubjects(colId, course, batch, userId, []byte(request.Body))
				if err != nil {
					return AUTH_500, err
				}
				return respondData(claims, "Updated at - "+strconv.FormatInt(ts, 10), err)
			} else if _type == "students" {
				err, ts := onboardStudents(colId, course, batch, userId, request.Body)
				if err != nil {
					return AUTH_500, err
				}
				return respondData(claims, "Updated at - "+strconv.FormatInt(ts, 10), err)
			}
		} else if request.Resource == "/metadata/getStudents" {
			colId := request.QueryStringParameters["college_id"]
			course := strings.ToLower(request.QueryStringParameters["class"])
			batch := request.QueryStringParameters["batch"]
			students, err := getStudents(colId, course, batch)
			if err != nil {
				return AUTH_500, err
			}
			if students == nil {
				return respondData(claims, "", nil)
			}
			fmt.Println(students)
			res, err := json.Marshal(students)
			if err != nil {
				return AUTH_500, err
			}
			return respondData(claims, string(res), err)
		} else if request.Resource == "/metadata/faculties" {
			//colId := request.QueryStringParameters["college_id"]
			//students, err := faculty.GetFaculties(colId)
			//if err != nil {
			//	return AUTH_500, err
			//}
			//if students == nil {
			//	return respondData(claims, "", nil)
			//}
			//fmt.Println(students)
			//res, err := json.Marshal(students)
			//if err != nil {
			//	return AUTH_500, err
			//}
			//return respondData(claims, string(res), err)
		} else if request.HTTPMethod == http.MethodPost && request.Resource == "/metadata/faculty/create" {
			colId := request.QueryStringParameters["college_id"]
			var facultyCreateForm faculty.Faculty
			err = json.Unmarshal([]byte(request.Body), &facultyCreateForm)
			isCreated, id := faculty.CreateFacultyMeta(colId, facultyCreateForm, userId)
			if !isCreated {
				return AUTH_504, err
			}
			return respondData(claims, string(id), err)
		}
	}
	return AUTH_404, errors.New("Resource not found")
}

type OnboardSubjectsMetadata struct {
	PK       string            `dynamodbav:"key"`
	SK       string            `dynamodbav:"skey"`
	Subjects map[string]string `dynamodbav:"subjects"`
	Ts       int64             `dynamodbav:"ts"`
	Updater  string            `dynamodbav:"uBy"`
}

type OnboardStudentsMetadata struct {
	PK       string `dynamodbav:"key"`
	SK       string `dynamodbav:"skey"`
	Students string `dynamodbav:"students"`
	Ts       int64  `dynamodbav:"ts"`
	Updater  string `dynamodbav:"uBy"`
}

type AttendanceFormUpdater struct {
	PK      string `dynamodbav:"sub"`
	SK      string `dynamodbav:"date"`
	Values  string `dynamodbav:"values"`
	Ts      int64  `dynamodbav:"ts"`
	Updater string `dynamodbav:"uBy"`
}

func updateAttendanceForm(colId string, course string, batch string, sub string, date string, req *attendanceForm) error {
	key := strings.ToLower(course) + "_" + strings.ToLower(batch) + "_" + strings.ToLower(sub)
	parsedReq, err := json.Marshal(req)
	if err != nil {
		return err
	}
	attendanceBook := AttendanceFormUpdater{
		PK:      key,
		SK:      date,
		Values:  string(parsedReq),
		Ts:      req.Ts,
		Updater: req.UBy,
	}
	data, err := attributevalue.MarshalMap(attendanceBook)
	_, err = DYNAMO_CFG.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("attendance" + "_" + colId),
		Item:      data,
	})

	if err != nil {
		return err
	}
	return nil
}

func onboardStudents(colId string, course string, batch string, userId string, reqBody string) (error, int64) {
	//reqBodyParsed := make(map[string]string)
	//err := json.Unmarshal(reqBody, &reqBodyParsed)
	fmt.Println(reqBody)
	//if err != nil {
	//	return err, 0
	//}
	ts := time.Now().UTC().Unix()
	subjects := OnboardStudentsMetadata{
		PK:       colId,
		SK:       strings.ToLower(course) + strings.ToLower(batch),
		Students: reqBody,
		Ts:       ts,
		Updater:  userId,
	}
	data2, _ := attributevalue.MarshalMap(subjects)
	_, err := DYNAMO_CFG.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("college_metadata"),
		Item:      data2,
	})
	return err, ts
}

func OnboardSubjects(colId string, course string, batch string, userId string, reqBody []byte) (error, int64) {
	reqBodyParsed := make(map[string]string)
	err := json.Unmarshal(reqBody, &reqBodyParsed)
	fmt.Println(reqBodyParsed)
	if err != nil {
		return err, 0
	}
	ts := time.Now().UTC().Unix()
	subjects := OnboardSubjectsMetadata{
		PK:       colId + "_" + course,
		SK:       batch,
		Subjects: reqBodyParsed,
		Ts:       ts,
		Updater:  userId,
	}
	data2, _ := attributevalue.MarshalMap(subjects)
	_, err = DYNAMO_CFG.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("college_metadata"),
		Item:      data2,
	})
	return err, ts
}

func getSubjects(collegeId string, class string, batch string) string {
	key, err := attributevalue.Marshal(collegeId + "_" + strings.ToLower(class))
	sKey, err := attributevalue.Marshal(batch)
	if err != nil {
		return ""
	}
	ck := map[string]types.AttributeValue{
		"key":  key,
		"skey": sKey,
	}
	fmt.Printf("key %s, skey %s\n", key, sKey)
	out, err := DYNAMO_CFG.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String("college_metadata"),
		Key:       ck,
	})
	item := out.Item["subjects"]
	var res interface{}
	_ = attributevalue.Unmarshal(item, &res)
	res, _ = json.Marshal(res)
	fmt.Printf("Marshaled result is - %s ", res)
	return fmt.Sprintf("%s", res)
}

func getAttendanceStudents(college string, course string, batch string, sub string, date string) (string, error) {
	students, err := getStudents(college, course, batch)
	studentsMap := make(map[string]int)
	for i := 0; i < len(students); i++ {
		studentsMap[students[i].Id] = i
	}
	aForm, err := getAttendanceForm(college, course, batch, sub, date)
	if err != nil {
		return "", err
	}
	for _, sId := range aForm.Students {
		if val, ok := studentsMap[sId]; ok {
			students[val].IsPresent = true
		}
	}
	res, err := json.Marshal(students)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func getAttendanceForm(college string, course string, batch string, sub string, date string) (attendanceForm, error) {
	key, err := attributevalue.Marshal(strings.ToLower(course) + "_" + strings.ToLower(batch) + "_" + strings.ToLower(sub))
	sKey, err := attributevalue.Marshal(date)
	if err != nil {
		return attendanceForm{}, err
	}
	ck := map[string]types.AttributeValue{
		"sub":  key,
		"date": sKey,
	}
	out, err := DYNAMO_CFG.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String("attendance" + "_" + strings.ToLower(college)),
		Key:       ck,
	})
	item := out.Item["values"]
	if item == nil {
		return attendanceForm{}, nil
	}
	var res1 string
	fmt.Println(item)
	err = attributevalue.Unmarshal(item, &res1)
	var res attendanceForm
	err = json.Unmarshal([]byte(res1), &res)
	return res, err
}

type student struct {
	Name      string `json:"name"`
	Id        string `json:"id"`
	IsPresent bool   `json:"isPresent"`
}

type attendanceForm struct {
	Students []string `json:"students"`
	Ts       int64    `json:"ts"`
	UBy      string   `json:"u_by"`
}

func getStudents(college string, course string, batch string) ([]student, error) {
	key, err := attributevalue.Marshal(college)
	sKey, err := attributevalue.Marshal(strings.ToLower(course) + strings.ToLower(batch))
	if err != nil {
		return nil, err
	}
	ck := map[string]types.AttributeValue{
		"key":  key,
		"skey": sKey,
	}
	out, err := DYNAMO_CFG.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String("college_metadata"),
		Key:       ck,
	})
	fmt.Println(out)
	item := out.Item["students"]
	if item == nil {
		return nil, nil
	}
	fmt.Println(item)
	var res string
	err = attributevalue.Unmarshal(item, &res)
	var parsedRes []student
	err = json.Unmarshal([]byte(res), &parsedRes)
	return parsedRes, err
}

func parseHmacToken(token string) string {
	decodeByte, err := base64.URLEncoding.DecodeString(token)
	decodeString := string(decodeByte[:])
	if err != nil {
		return ""
	}
	tknParts := strings.Split(decodeString, ".")
	currentTs := time.Now().UTC().Unix()
	if tknTs, _ := strconv.ParseInt(tknParts[0], 10, 64); tknTs <= currentTs {
		return ""
	}
	err = verifySignature(tknParts)
	if err != nil {
		return ""
	}
	fmt.Println(tknParts)
	return tknParts[1]
}

func verifySignature(tknParts []string) error {
	tsMessage := tknParts[0] + "." + tknParts[1]
	uEnc := base64.URLEncoding.EncodeToString([]byte(tsMessage))
	calculatedSignature := getSignature(uEnc)
	if tknParts[2] != calculatedSignature {
		return errors.New("Signature not matched")
	}
	return nil
}

func getHmacData(exp int64, resp string, userId string) string {
	tsMessage := strconv.FormatInt(exp, 10) + "." + resp + "." + userId
	uEnc := base64.URLEncoding.EncodeToString([]byte(tsMessage))
	return base64.URLEncoding.EncodeToString([]byte(tsMessage + "." + getSignature(uEnc)))
}

func respondData(claims jwt.MapClaims, response string, err error) (events.APIGatewayProxyResponse, error) {
	if err != nil {
		return AUTH_500, nil
	}
	//respHeader := ResponseHeaders
	//respHeader["cerp-api-token"] = tkn
	//respHeader["Set-Cookie"] = "_cerp_api_=" + tkn + "; Domain=localhost; Path=/;Secure; HttpOnly;SameSite=None;Priority=Medium"
	return events.APIGatewayProxyResponse{Body: response, StatusCode: 200, IsBase64Encoded: true, Headers: ResponseHeaders}, nil
}

func getSignature(encMessage string) string {
	mac := hmac.New(sha256.New, SECRET)
	_, err := mac.Write([]byte(encMessage))
	if err != nil {
		return ""
	}
	return hex.EncodeToString(mac.Sum(nil))
}

func fetchRoles(userId string) string {
	key, err := attributevalue.Marshal(userId)
	sKey, err := attributevalue.Marshal("roles")
	if err != nil {
		return ""
	}
	ck := map[string]types.AttributeValue{
		"user_id": key,
		"field":   sKey,
	}
	out, err := DYNAMO_CFG.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String("users_info"),
		Key:       ck,
	})
	if out == nil {
		return ""
	}
	item := out.Item["roles"]
	var res string
	_ = attributevalue.Unmarshal(item, &res)
	return res
}

func verifyToken(tokenStr string) (jwt.MapClaims, error) {
	tokenSlice := strings.Split(tokenStr, " ")
	var bearerToken string
	if len(tokenSlice) > 1 {
		bearerToken = tokenSlice[len(tokenSlice)-1]
	}

	// if no bearer token set return unauthorized.
	if bearerToken == "" {
		return nil, errors.New("unauthorized")
	}

	jwks, err := fetchJWKS()
	if err != nil {
		return nil, err
	}

	// Parse takes the token string using function to looking up the key.
	token, err := jwt.Parse(bearerToken, jwks.Keyfunc)
	if err != nil {
		if verr, ok := err.(*jwt.ValidationError); ok {
			if verr.Errors == jwt.ValidationErrorMalformed {
				return nil, errors.New("unauthorized")
			}
			if verr.Errors == jwt.ValidationErrorExpired {
				return nil, errors.New("token is expired")
			}
		}
		return nil, err
	}

	// handle nil token scenario, unlikely to happen.
	if token == nil {
		return nil, errors.New("no token after JWT parsing")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	// check if claims are present and token is valid.
	if ok && token.Valid {
		// return Allow authResponse with userEntity in authorizer context for next lambda in chain.
		err = validateClaims(claims)
		return claims, err
	}
	return nil, nil
}

func validateClaims(claims jwt.MapClaims) error {
	if !claims.VerifyIssuer(cfg_details.CLAIM_ISS, true) || claims["azp"] != cfg_details.CLAIM_CLIENT_ID {
		return errors.New("Issuer/CLIENT_ID is wrong")
	}
	return nil
}

func fetchJWKS() (*keyfunc.JWKS, error) {
	options := keyfunc.Options{
		RefreshErrorHandler: func(err error) {
			log.Printf("There was an error with the jwt.KeyFunc\nError:%s\n", err.Error())
		},
		RefreshUnknownKID: true,
	}
	return keyfunc.Get("https://dev-q0ywv1av1mdo8z4n.us.auth0.com/.well-known/jwks.json", options)
}
