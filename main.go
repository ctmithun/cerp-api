package main

import (
	"cerpApi/cfg_details"
	"cerpApi/faculty"
	jwtVerifier "cerpApi/jwt"
	"cerpApi/students"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	_ "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"net/url"
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
var AUTH_400 = getProxyResponse("Invalid fId!", 400)

func getProxyResponse(body string, statusCode int) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{Body: body, StatusCode: statusCode, IsBase64Encoded: true, Headers: ResponseHeaders}
}

func main() {

	//ch := make(chan bool)
	//
	//go func() {
	//	// Do some work
	//	ch <- false
	//}()
	//
	//// Wait for the goroutine to finish
	//res := <-ch
	//print(res)

	//log.Println("Running lambda-authorizer")
	lambda.Start(handler)
	//fmt.Println(faculty.GetFacultyAssignedSubjects("ni", "auth0|ni|93aec99611a99b76eb82124157566651"))
	//dsa := students.GetRowNumber("2025-BCA", "ni")
	//fmt.Println(dsa)

	//student := students.Student{
	//	Email:       "kkumar@email.com",
	//	Id:          "auth0|ni|54addb248438a25ab5fec6b4c673f06b",
	//	Name:        "Kiran Ram Kumar M",
	//	PhoneNumber: "9875673214",
	//	Doj:         "2025-01-15",
	//	Sid:         "2025-BCOM-1",
	//	Batch:       "2025",
	//	Stream:      "BCOM",
	//	Fees:        120000,
	//}
	//students.DeactivateStudent("ni", student, "test-user")
	//students.OnboardStudent("ni", &student, "test-user")
	//students.UpdateStudentRecord("ni", &student, "test-user")
	//students.GetStudentsData("ni", "2025", "BCA")
	//fmt.Println("Created Student - ", sId)
	//claims, _ := jwtVerifier.VerifyToken("Bearer eyJraWQiOiJQTExXaEpyUVFuSEZTQWcwWWhkSE5FcXJTZzVWVUhqVEdCd2M0V3YzcElvPSIsImFsZyI6IlJTMjU2In0.eyJzdWIiOiI5MDBiZWY3Mi1iNmM4LTQxOWQtYjhhMS02YTc4YmVhMzJiYTMiLCJpc3MiOiJodHRwczpcL1wvY29nbml0by1pZHAuZXUtd2VzdC0xLmFtYXpvbmF3cy5jb21cL2V1LXdlc3QtMV9aaFdIOFk1WU8iLCJ2ZXJzaW9uIjoyLCJjbGllbnRfaWQiOiIybmUzcXZiaWRtMzlnaWhlYzBvZmExczY5NyIsImV2ZW50X2lkIjoiY2YxNjM4OTctZjMyYi00MTNiLWFmYzEtNTI1OGU3NTVmODY1IiwidG9rZW5fdXNlIjoiYWNjZXNzIiwic2NvcGUiOiJhd3MuY29nbml0by5zaWduaW4udXNlci5hZG1pbiBwaG9uZSBvcGVuaWQgcHJvZmlsZSBlbWFpbCIsImF1dGhfdGltZSI6MTczNjIyNjU0NywiZXhwIjoxNzM2MjMwMTQ3LCJpYXQiOjE3MzYyMjY1NDcsImp0aSI6IjQ3YjgwNDcwLWRjYzktNDNlYy1hYzJkLTk1Yzc0YmI5YTU5MSIsInVzZXJuYW1lIjoiOTAwYmVmNzItYjZjOC00MTlkLWI4YTEtNmE3OGJlYTMyYmEzIn0.bD8Ksxw7isJWRbgzbhhioxicxXoCaAUVCCcOWYLpgmtrkMEi8efqkGm4C6o-aDQXLfeeQYXOtiIkdzIHqAW6MyfxxPNva6X7mncDyVywp6-7Mq964VOf8pn4tcUs6c29Rqv72rS7eM4kd2TCqdlJa_qyLFDs3oRZoCQ_zr1fbQvLDohIht4I-79dGz5-xvSsXE85BpyVo6cuz4COLy8Noir_bhZiVpvAxpiCH1b1vwKOeYL-zDgnz2xRp6IzjlVP0r8HdTkonvw5yUd63B468nN3uh8L77jfvgKrC4_XMuLBIcGpgIT_HV39c3PP0JcgLR1IC8zUhjZ9_fxsiyCeUg")
	//res := extractRoles(claims)
	//fmt.Println(res)
	//data := []byte("ni" + "|" + "ctiitm@rediffmail.com")
	//hash := md5.Sum(data)
	//uId := hex.EncodeToString(hash[:])
	//fmt.Println(uId)

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

	//data := faculty.Faculty{
	//	Email:       "tcta@cta.com",
	//	Id:          "auth0|ni|6eff353c8aa24efffb5d6dafa139296a",
	//	Name:        "Thippeswamy M",
	//	PhoneNumber: "9877880216",
	//	Doj:         "2025-01-01",
	//	Subjects:    "MBA_SEM-1_E21CT,MBA_SEM-1_E22CT",
	//	Description: "New adjunct faculty",
	//	Type:        "Full-Time",
	//}
	//faculty.DeactivateFaculty("ni", data, "test-user")
	//faculty.ModifyFacultyData("ni", data, "test-user")
	//var wg sync.WaitGroup
	//wg.Add(1)
	//var err error
	//faculty.SetUserRoles("auth0|ni|01JGVFGJXEYWBFNCTM7V0DMV22", &wg, &err)
	//wg.Wait()
	//token := parseHmacToken("MTczNjY4MTg5Ny5hZG1pbixmYWN1bHR5LmF1dGgwfDY3MjVlM2QwMDkyNDNkYjVhYzE0NGNkYy5jMzhkMGJjZDJlODBmZDkwNjA2YmM3Zjk0OTVkNzdlZTlmYmUwNjVhNGQwNGZlNTQxODk1YjY5M2I0YmNiYjA1")
	//fmt.Println(token)
	//fac := faculty.Faculty{
	//	Email:       "ctiitm2@rediffmail.com",
	//	Id:          "auth0|ni|01JGVJVTN917CXPRA8TKMHB0M2",
	//	Name:        "Manjula C P",
	//	PhoneNumber: "9743213022",
	//	Doj:         "2025-01-02",
	//	Subjects:    "BCA_SEM-1_CA-C3T,BCA_SEM-1_CA-C2T",
	//	Description: "",
	//}
	//faculty.CreateFacultyMeta("ni", fac, "test-user")
	//fId, _ := url.PathUnescape("auth0%7Cni%7Cc99b893c21d24e4b48e8a7e7c22f7d76")
	//faculty.GetFacultiesData("ni", fId)
	//students.GetStudentsData("ni", "2025", "BCA")
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	claims, err := jwtVerifier.VerifyToken(request.Headers["Authorization"])
	if err != nil {
		return AUTH_500, err
	}
	userId := claims["sub"].(string)
	//resPath := request.Resource
	fmt.Printf("Resource path requested is %s", claims)
	if request.Resource == "/roles" {
		res := extractRoles(claims)
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
	if roles == "" {
		return AUTH_403, errors.New("Invalid Token!!")
	}
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
	} else if request.Resource == "/metadata/faculty/subjects" {
		if strings.Contains(roles, "faculty") || strings.Contains(roles, "admin") {
			colId := request.QueryStringParameters["college_id"]
			data := faculty.GetFacultyAssignedSubjects(colId, userId)
			return respondData(claims, data, err)
		}
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
				err, ts := onboardStudentsAttendance(colId, course, batch, userId, request.Body)
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
			if err != nil {
				return AUTH_504, err
			}
			isCreated, id := faculty.CreateFacultyMeta(colId, facultyCreateForm, userId)
			if !isCreated {
				return AUTH_504, err
			}
			return respondData(claims, string(id), err)
		} else if request.HTTPMethod == http.MethodPost && request.Resource == "/metadata/faculty/update" {
			colId := request.QueryStringParameters["college_id"]
			var facultyCreateForm faculty.Faculty
			err = json.Unmarshal([]byte(request.Body), &facultyCreateForm)
			if err != nil {
				return AUTH_504, err
			}
			isUpdated, res := faculty.ModifyFacultyData(colId, facultyCreateForm, userId)
			if !isUpdated {
				return respondError(res)
			}
			return respondData(claims, res, err)
		} else if request.HTTPMethod == http.MethodGet && request.Resource == "/metadata/faculty/manage" {
			colId := request.QueryStringParameters["college_id"]
			fId, err := url.PathUnescape(request.QueryStringParameters["fId"])
			if err != nil {
				return AUTH_400, err
			}
			return respondData(claims, faculty.GetFacultiesData(colId, fId), err)
		} else if request.HTTPMethod == http.MethodDelete && request.Resource == "/metadata/faculty/delete" {
			colId := request.QueryStringParameters["college_id"]
			var facultyCreateForm faculty.Faculty
			err = json.Unmarshal([]byte(request.Body), &facultyCreateForm)
			if err != nil {
				return AUTH_504, err
			}
			isDeactivated, errMessage := faculty.DeactivateFaculty(colId, facultyCreateForm, userId)
			if !isDeactivated {
				return respondError(errMessage)
			}
			return respondData(claims, errMessage, err)
		} else if request.HTTPMethod == http.MethodPost && request.Resource == "/metadata/student/create" {
			colId := request.QueryStringParameters["college_id"]
			var studentOnboardForm students.Student
			err = json.Unmarshal([]byte(request.Body), &studentOnboardForm)
			if err != nil {
				return AUTH_504, err
			}
			sId, err := students.OnboardStudent(colId, &studentOnboardForm, userId)
			if err != nil {
				return AUTH_504, err
			}
			return respondData(claims, sId, err)
		} else if request.HTTPMethod == http.MethodPost && request.Resource == "/metadata/student/update" {
			colId := request.QueryStringParameters["college_id"]
			var studentOnboardForm students.Student
			err = json.Unmarshal([]byte(request.Body), &studentOnboardForm)
			if err != nil || studentOnboardForm.Id == "" {
				return AUTH_504, err
			}
			sId, res := students.UpdateStudentRecord(colId, &studentOnboardForm, userId)
			if !sId && res != "" {
				return respondError(res)
			}
			return respondData(claims, res, err)
		} else if request.HTTPMethod == http.MethodDelete && request.Resource == "/metadata/student/delete" {
			colId := request.QueryStringParameters["college_id"]
			var studentOnboardForm students.Student
			err = json.Unmarshal([]byte(request.Body), &studentOnboardForm)
			if err != nil || studentOnboardForm.Id == "" {
				return AUTH_504, err
			}
			sId, res := students.DeactivateStudent(colId, studentOnboardForm, userId)
			if !sId && res != "" {
				return respondError(res)
			}
			return respondData(claims, res, err)
		} else if request.HTTPMethod == http.MethodGet && request.Resource == "/metadata/student/manage" {
			colId := request.QueryStringParameters["college_id"]
			batch := request.QueryStringParameters["batch"]
			stream := request.QueryStringParameters["class"]
			res, err := json.Marshal(students.GetStudentsData(colId, batch, stream))
			if err != nil {
				return AUTH_504, err
			}
			return respondData(claims, string(res), err)
		}
	}
	return AUTH_404, errors.New("Resource not found")
}

func respondError(res string) (events.APIGatewayProxyResponse, error) {
	switch res {
	case cfg_details.INVALID_DATA:
		return AUTH_403, errors.New(res)
	default:
		return AUTH_504, errors.New(res)
	}
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

func onboardStudentsAttendance(colId string, course string, batch string, userId string, reqBody string) (error, int64) {
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
	fmt.Printf(tknParts[1])
	return tknParts[1]
}

func verifySignature(tknParts []string) error {
	tsMessage := tknParts[0] + "." + tknParts[1] + "." + tknParts[2]
	uEnc := base64.URLEncoding.EncodeToString([]byte(tsMessage))
	calculatedSignature := getSignature(uEnc)
	if tknParts[3] != calculatedSignature {
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

func extractRoles(claims jwt.MapClaims) string {
	roleArray := claims[cfg_details.ALLOWED_URL].([]interface{})
	res := make([]string, len(roleArray))
	for i := 0; i < len(roleArray); i++ {
		res[i] = strings.ToLower(roleArray[i].(string))
	}
	return strings.Join(res, ",")
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
