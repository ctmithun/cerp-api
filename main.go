package main

import (
	"bytes"
	"cerpApi/attendance"
	"cerpApi/cfg_details"
	"cerpApi/enquiry"
	"cerpApi/faculty"
	jwtVerifier "cerpApi/jwt"
	"cerpApi/onboard_data"
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
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/golang-jwt/jwt/v4"
	"github.com/grokify/go-awslambda"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

//const (
//	CLAIM_ISS       = "https://dev-q0ywv1av1mdo8z4n.us.auth0.com/"
//	CLAIM_CLIENT_ID = "t4pVw4sPvFWgvq4n3DgvFGnavMatqYwv"
//)

var SECRET = []byte("This is the way -- Mando")

// var CFG, _ = config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("mumbai"), config.WithRegion("ap-south-1"))
var CFG, _ = config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-south-1"))
var DYNAMO_CFG = dynamodb.NewFromConfig(CFG)
var ResponseHeaders = map[string]string{"Access-Control-Allow-Origin": "*", "X-Frame-Options": "SAMEORIGIN", "Strict-Transport-Security": "max-age=31557600; includeSubDomains"}
var AUTH_500 = getProxyResponse("Invalid Input!", 500)
var AUTH_504 = getProxyResponse("Service Unavailable Input!", 504)
var AUTH_404 = getProxyResponse("Service Not Found!", 404)
var AUTH_403 = getProxyResponse("Access Denied!", 403)
var AUTH_400 = getProxyResponse("Invalid fId!", 400)

const (
	ROLE_ADMIN      = "admin"
	ROLE_COUNSELLOR = "counsellor"
)

type postgres struct {
	db *pgxpool.Pool
}

var (
	pgInstance *postgres
	pgOnce     sync.Once
	s3Client   *s3.Client
	uploader   *manager.Uploader
)

func NewPG(ctx context.Context, connString string) (*postgres, error) {
	pgOnce.Do(func() {
		db, err := pgxpool.New(ctx, connString)
		if err != nil {
			err = fmt.Errorf("unable to create connection pool: %w", err)
			return
		}

		pgInstance = &postgres{db}
	})

	return pgInstance, nil
}

func (pg *postgres) Ping(ctx context.Context) error {
	return pg.db.Ping(ctx)
}

func (pg *postgres) Close() {
	pg.db.Close()
}

func getProxyResponse(body string, statusCode int) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{Body: body, StatusCode: statusCode, IsBase64Encoded: true, Headers: ResponseHeaders}
}

func main() {
	_, err := NewPG(context.Background(), "postgresql://cerp:7pFJwJHKWvWwIRycQ9yXew@weary-flapper-8111.j77.aws-ap-south-1.cockroachlabs.cloud:26257/cerp?sslmode=verify-full")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer log.Println("In main method after handler and after DB close stack call")
	defer pgInstance.Close()

	s3Client = s3.NewFromConfig(CFG)
	uploader = manager.NewUploader(s3Client)

	//fileName := "signature.jpg" // Change this to the file you want to upload
	//file, err := os.Open(fileName)
	//if err != nil {
	//	log.Fatalf("Failed to open file: %v", err)
	//}
	//defer file.Close()

	// Upload the fil
	lambda.Start(handler)

	//if err != nil {
	//	log.Fatalf("Failed to load AWS config: %v", err)
	//}

}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	claims, err := jwtVerifier.VerifyToken(request.Headers["Authorization"])
	if err != nil {
		return AUTH_500, err
	}
	userId := claims["sub"].(string)
	fmt.Printf("Resource is %s\n", request.Resource)
	//fmt.Printf("Resource path requested is %s\n", claims)
	//fmt.Printf("Resource Headers requested are %s\n", request.Headers)
	if request.Resource == "/roles" {
		res := extractRoles(claims)
		if res == "" {
			return AUTH_403, errors.New("no roles set")
		}
		tkn := getHmacData(int64(int(claims["exp"].(float64))), res, userId)
		log.Printf("Encoded Token is %s\n", tkn)
		return respondData(tkn, err)
	} else if request.Resource == "/userRoles" {
		res := fetchRoles(strings.Split(userId, "|")[1])
		if res == "" {
			return AUTH_500, errors.New("no roles set")
		}
		return respondData(res, err)
	}
	roles := parseHmacToken(request.Headers["cerp-api-token"])
	if roles == "" {
		return AUTH_403, errors.New("Invalid Token!!")
	}
	fmt.Printf("Found roles %s\n", roles)
	if request.Resource == "/admission/admit" && request.HTTPMethod == http.MethodPost {
		if strings.Contains(roles, ROLE_COUNSELLOR) || strings.Contains(roles, ROLE_ADMIN) {
			sId, err := handleAdmission(request, userId)
			if err != nil {
				return respondError(err.Error())
			}
			return respondData(sId, err)
		}
		return AUTH_403, errors.New("No roles set")

	} else if request.Resource == "/attendance/students" {
		if strings.Contains(roles, "faculty") || strings.Contains(roles, "admin") {
			colId := request.QueryStringParameters["college_id"]
			course := request.QueryStringParameters["class"]
			batch := request.QueryStringParameters["batch"]
			subject := request.QueryStringParameters["subject"]
			date := request.QueryStringParameters["date"]
			cs := request.QueryStringParameters["class_section"]
			res, err := getAttendanceStudents(colId, course, batch, subject, date, cs)
			if err != nil {
				return AUTH_500, err
			}
			return respondData(res, err)
		}
		return AUTH_500, errors.New("no roles set")
	} else if request.Resource == "/attendance/export" && request.HTTPMethod == http.MethodGet {
		if strings.Contains(roles, "faculty") || strings.Contains(roles, "admin") {
			colId := request.QueryStringParameters["college_id"]
			course := request.QueryStringParameters["class"]
			batch := request.QueryStringParameters["batch"]
			subject := request.QueryStringParameters["subject"]
			cs := request.QueryStringParameters["class_section"]
			res := attendance.GetAttendanceReport(colId, course, batch, subject, cs)
			if res == "" {
				return respondData("Attendance Export Data Unavailable!", nil)
			}
			return respondData(res, nil)
		}
		return AUTH_500, errors.New("no roles set")
	} else if request.Resource == "/subject" {
		fmt.Println("Inisde the subject API...")
		if strings.Contains(roles, "faculty") || strings.Contains(roles, "admin") {
			fmt.Println("Getting the subjects ...")
			res := getSubjects(request.QueryStringParameters["college_id"], request.QueryStringParameters["class"], request.QueryStringParameters["batch"])
			return respondData(res, err)
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
			cs := request.QueryStringParameters["class_section"]
			val, exist := request.QueryStringParameters["send_absent_notification"]
			sn := false
			if exist {
				sn, err = strconv.ParseBool(val)
				if err != nil {
					return respondError("Invalid input to send_notification query parameter")
				}
			}
			if err != nil {
				log.Printf("Error parsing send_absent_notification %v\n", err)
				return AUTH_500, err
			}
			var attendanceForm attendance.AttendanceForm
			err = json.Unmarshal([]byte(request.Body), &attendanceForm)
			fmt.Println(attendanceForm)
			attendanceForm.UBy = userId
			attendanceForm.Ts = time.Now().UTC().Unix()
			if err != nil {
				return respondError(err.Error())
			}
			studentsSet := make([]attendance.Student, 0)
			if sn {
				studentsSet, err = getStudents(colId, course, batch, cs)
				if err != nil {
					log.Printf("Error getting students while updating the attendance %v\n", err)
					return respondError(err.Error())
				}
			}
			err = attendance.UpdateAttendance(colId, course, batch, subject, date, cs, sn, &attendanceForm, studentsSet)
			if err != nil {
				return AUTH_500, err
			}
			return respondData("Updated at - "+strconv.FormatInt(attendanceForm.Ts, 10), err)
		}
		return AUTH_500, errors.New("no roles set")
	} else if request.Resource == "/metadata/faculty/subjects" {
		if strings.Contains(roles, "faculty") || strings.Contains(roles, "admin") {
			colId := request.QueryStringParameters["college_id"]
			data := faculty.GetFacultyAssignedSubjects(colId, userId)
			return respondData(data, err)
		}
	} else if strings.Contains(request.Resource, "/enq/") && (strings.Contains(roles, ROLE_COUNSELLOR) || strings.Contains(roles, ROLE_ADMIN)) {
		userName := claims[cfg_details.ALLOWED_URL+"name"].(string)
		return handleEnquiry(roles, request, userId, ctx, userName)
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
				return respondData("Updated at - "+strconv.FormatInt(ts, 10), err)
			} else if _type == "students" {
				cs := request.QueryStringParameters["class_section"]
				err, ts := onboardStudentsAttendance(colId, course, batch, userId, cs, request.Body)
				if err != nil {
					return AUTH_500, err
				}
				return respondData("Updated at - "+strconv.FormatInt(ts, 10), err)
			} else if _type == "s2s" {
				cs := request.QueryStringParameters["class_section"]
				res, err := onboard_data.OnboardS2S(colId, course, batch, cs, userId, []byte(request.Body))
				if err != nil {
					return respondError("Update Failed - " + err.Error())
				}
				return respondData("Updated at - "+res, err)
			}
		} else if request.Resource == "/metadata/getStudents" {
			colId := request.QueryStringParameters["college_id"]
			course := strings.ToLower(request.QueryStringParameters["class"])
			batch := request.QueryStringParameters["batch"]
			cs := request.QueryStringParameters["class_section"]
			students, err := getStudents(colId, course, batch, cs)
			if err != nil {
				return AUTH_500, err
			}
			if students == nil {
				return respondData("", nil)
			}
			fmt.Println(students)
			res, err := json.Marshal(students)
			if err != nil {
				return AUTH_500, err
			}
			return respondData(string(res), err)
		} else if request.Resource == "/metadata/s2s" && request.HTTPMethod == http.MethodGet {
			colId := request.QueryStringParameters["college_id"]
			course := strings.ToLower(request.QueryStringParameters["class"])
			batch := request.QueryStringParameters["batch"]
			cs := request.QueryStringParameters["class_section"]
			if err != nil {
				return respondError(err.Error())
			}
			res, err := onboard_data.GetS2S(colId, course, batch, cs)
			if err != nil {
				return respondError(err.Error())
			}
			return respondData(res, err)
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
			return respondData(string(id), err)
		} else if request.HTTPMethod == http.MethodPost && request.Resource == "/metadata/faculty/update" {
			colId := request.QueryStringParameters["college_id"]
			isRoleUpdate, _ := strconv.ParseBool(request.QueryStringParameters["lore"])
			var facultyCreateForm faculty.Faculty
			err = json.Unmarshal([]byte(request.Body), &facultyCreateForm)
			if err != nil {
				return AUTH_504, err
			}
			isUpdated, res := faculty.ModifyFacultyData(colId, facultyCreateForm, userId, isRoleUpdate)
			if !isUpdated {
				return respondError(res)
			}
			return respondData(res, err)
		} else if request.HTTPMethod == http.MethodGet && request.Resource == "/metadata/faculty/manage" {
			colId := request.QueryStringParameters["college_id"]
			fId, err := url.PathUnescape(request.QueryStringParameters["fId"])
			if err != nil {
				return AUTH_400, err
			}
			return respondData(faculty.GetFacultiesData(colId, fId), err)
		} else if request.HTTPMethod == http.MethodDelete && request.Resource == "/metadata/faculty/delete" {
			colId := request.QueryStringParameters["college_id"]
			var facultyCreateForm faculty.Faculty
			err = json.Unmarshal([]byte(request.Body), &facultyCreateForm)
			if err != nil {
				return AUTH_504, err
			}
			isDeactivated, errMessage := faculty.DeleteFaculty(colId, facultyCreateForm, userId)
			if !isDeactivated {
				return respondError(errMessage)
			}
			return respondData(errMessage, err)
		} else if request.HTTPMethod == http.MethodDelete && request.Resource == "/metadata/faculty/deactivate" {
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
			return respondData(errMessage, err)
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
			return respondData(sId, err)
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
			return respondData(res, err)
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
			return respondData(res, err)
		} else if request.HTTPMethod == http.MethodGet && request.Resource == "/metadata/student/manage" {
			colId := request.QueryStringParameters["college_id"]
			batch := request.QueryStringParameters["batch"]
			stream := request.QueryStringParameters["class"]
			res, err := json.Marshal(students.GetStudentsData(colId, batch, stream))
			if err != nil {
				return AUTH_504, err
			}
			return respondData(string(res), err)
		}
	}
	return AUTH_404, errors.New("Resource not found")
}

func handleAdmission(request events.APIGatewayProxyRequest, uBy string) (string, error) {

	colId := request.QueryStringParameters["college_id"]
	contentType := request.Headers["content-type"]

	if !strings.HasPrefix(contentType, "multipart/form-data") {
		return "", errors.New("Invalid Content-Type")
	}
	boundary, err := extractBoundary(contentType)
	if err != nil || boundary == "" {
		log.Printf("Invalid boundary %v\n", err)
		return "", errors.New("Invalid Boundary")
	}
	mr, err := awslambda.NewReaderMultipart(request)
	fmt.Println("Inside handleAdmission - ")
	formMap := make(map[string]string)
	formFilesMap := make(map[string][]byte)
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading multipart data %v\n", err)
			return err.Error(), err
		}
		content, err := io.ReadAll(part)

		// Skip non-file fields
		if part.FileName() == "" {
			formMap[part.FormName()] = string(content)
			continue
		}
		if len(content) <= 0 {
			return "", errors.New("Empty Files are not allowed to upload!!")
		}
		fType := strings.Split(part.FileName(), ".")[1]
		fileKey := part.FormName() + "." + fType
		formFilesMap[fileKey] = content
		fileProps := make(map[string]string)
		fileProps["type"] = fType
		fileProps["key"] = fileKey
		fileProps["uploaded"] = "true"
		strFileProps, err := json.Marshal(fileProps)
		if err != nil {
			log.Printf("Error marshalling file properties %v\n", err)
			return err.Error(), err
		}
		formMap[part.FormName()] = string(strFileProps)
	}

	jsonData, err := json.Marshal(formMap)
	if err != nil {
		log.Printf("Error marshalling form json %v\n", err)
		return err.Error(), err
	}
	var student students.Student
	err = json.Unmarshal(jsonData, &student)
	if err != nil {
		log.Printf("Error unmarshalling from json to struct %v\n", err)
		return err.Error(), err
	}
	sId, err := students.OnboardStudentV2(colId, &student, uBy)
	if err != nil {
		return err.Error(), err
	}
	for k, v := range formFilesMap {
		err = uploadToS3(k, v, sId, colId)
		if err != nil {
			log.Printf("Error uploading to s3 %s %s %s %v\n", sId, colId, k, err)
			return err.Error(), err
		}
	}
	return sId, nil
}

func uploadToS3(fileName string, fileData []byte, sId string, colId string) error {
	_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String("cerp-students"),
		Key:    aws.String(colId + "/" + sId + "/" + fileName),
		Body:   bytes.NewReader(fileData),
	})
	if err != nil {
		return fmt.Errorf("failed to upload %s to S3: %w", fileName, err)
	}
	return nil
}

func extractBoundary(contentType string) (string, error) {
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", err
	}
	log.Println(params)
	boundary, exists := params["boundary"]
	if !exists {
		return "", fmt.Errorf("boundary not found in Content-Type")
	}
	log.Printf("boundary found: %s\n", boundary)
	return "1073741824", nil
}

func handleEnquiry(roles string, request events.APIGatewayProxyRequest, uBy string, ctx context.Context, name string) (events.APIGatewayProxyResponse, error) {
	if strings.Contains(roles, ROLE_ADMIN) || strings.Contains(roles, ROLE_COUNSELLOR) {
		colId := request.QueryStringParameters["college_id"]
		if request.HTTPMethod == http.MethodPost && strings.Contains(request.Resource, "/create") {
			var formData enquiry.FormData
			err := json.Unmarshal([]byte(request.Body), &formData)
			if err != nil {
				fmt.Println(err.Error())
				return respondError(cfg_details.INVALID_DATA)
			}
			formData.CouncillorName = name
			id, err := enquiry.AddEnqV2(pgInstance.db, ctx, formData, colId, uBy)
			if err != nil {
				fmt.Println(err.Error())
				return respondError(cfg_details.INVALID_DATA)
			}
			return respondData201(id)
		} else if request.HTTPMethod == http.MethodPut && strings.Contains(request.Resource, "/update") {
			var formData enquiry.FormData
			err := json.Unmarshal([]byte(request.Body), &formData)
			if err != nil {
				return respondError(cfg_details.INVALID_DATA)
			}
			err = enquiry.UpdateEnqV2(pgInstance.db, ctx, formData, colId, uBy)
			if err != nil {
				return respondError(cfg_details.INVALID_DATA)
			}
			return respondData("Updated", err)
		} else if request.HTTPMethod == http.MethodGet && strings.Contains(request.Resource, "/list") {
			formData, err := enquiry.ListEnqV2(pgInstance.db, ctx, colId)
			if err != nil {
				return respondError(cfg_details.INVALID_DATA)
			}
			res, err := json.Marshal(formData)
			return respondData(string(res), err)
		} else if request.HTTPMethod == http.MethodGet && strings.Contains(request.Resource, "/enq/get") {
			eqId, err := strconv.ParseInt(request.QueryStringParameters["eq_id"], 10, 64)
			if err != nil {
				return respondError(cfg_details.INVALID_DATA)
			}
			formData, err := enquiry.GetEnqV2(pgInstance.db, ctx, eqId, colId)
			if err != nil {
				return respondError(cfg_details.INVALID_DATA)
			}
			res, err := json.Marshal(formData)
			return respondData(string(res), err)
		} else if request.HTTPMethod == http.MethodDelete && strings.Contains(request.Resource, "/delete") {
			eqId, err := strconv.ParseInt(request.QueryStringParameters["eq_id"], 10, 64)
			if err != nil {
				return respondError(cfg_details.INPUT_ERROR)
			}
			err = enquiry.DelEnqV2(pgInstance.db, ctx, eqId, colId)
			if err != nil {
				return respondError(cfg_details.DATA_SERVICE_ERROR)
			}
			return respondData("Deleted", err)
		} else if request.HTTPMethod == http.MethodPost && strings.Contains(request.Resource, "/comments/add") {
			log.Println("Inside comments/add")
			var comment enquiry.Comment
			err := json.Unmarshal([]byte(request.Body), &comment)
			if err != nil {
				log.Println(err.Error())
				return respondError(cfg_details.INPUT_ERROR)
			}
			err = enquiry.AddCommentV2(pgInstance.db, ctx, comment, colId)
			if err != nil {
				return respondError(cfg_details.DATA_SERVICE_ERROR)
			}
			log.Println("Inside comments/add - after inserting")
			return respondData("Comment Added", err)
		} else if request.HTTPMethod == http.MethodGet && strings.Contains(request.Resource, "/comments/get") {
			log.Println("Inside comments/get")
			eqId, err := strconv.ParseInt(request.QueryStringParameters["eq_id"], 10, 64)
			if err != nil {
				return respondError(cfg_details.INPUT_ERROR)
			}
			cmts, err := enquiry.GetCommentV2(pgInstance.db, ctx, eqId, colId)
			log.Println("Inside comments/get - Queried")
			if err != nil {
				return respondError(cfg_details.DATA_SERVICE_ERROR)
			}
			res, err := json.Marshal(cmts)
			if err != nil {
				return respondError(cfg_details.CODE_ERROR)
			}
			log.Println("Inside comments/get - Marshaled")
			return respondData(string(res), err)
		}
	}
	return respondError("404")
}

func respondError(res string) (events.APIGatewayProxyResponse, error) {
	switch res {
	case cfg_details.INVALID_DATA:
		return AUTH_403, errors.New(res)
	case "404":
		return AUTH_404, errors.New(res)
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

func onboardStudentsAttendance(colId string, course string, batch string, userId string, cs string, reqBody string) (error, int64) {
	//reqBodyParsed := make(map[string]string)
	//err := json.Unmarshal(reqBody, &reqBodyParsed)
	fmt.Println(reqBody)
	//if err != nil {
	//	return err, 0
	//}
	ts := time.Now().Unix()
	subjects := OnboardStudentsMetadata{
		PK:       colId,
		SK:       strings.ToLower(course) + "_" + strings.ToLower(batch) + "_" + strings.ToLower(cs),
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
	ts := time.Now().Unix()
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

func getAttendanceStudents(college string, course string, batch string, sub string, date string, cs string) (string, error) {
	studentsLst, err := getStudents(college, course, batch, cs)
	if err != nil {
		return "", err
	}
	sPerSub, err := onboard_data.GetS2SPerSub(college, course, batch, cs, sub)
	if err != nil {
		return "", err
	}
	filteredStudents := make([]attendance.Student, 0)
	if sPerSub != nil {
		for _, val := range studentsLst {
			_, exist := sPerSub[val.Id]
			if exist {
				filteredStudents = append(filteredStudents, val)
			}
		}
	}
	if len(filteredStudents) > 0 {
		studentsLst = filteredStudents
	}
	studentsMap := make(map[string]int)
	for i := 0; i < len(studentsLst); i++ {
		studentsMap[studentsLst[i].Id] = i
	}
	aForm, err := attendance.GetAttendanceForm(college, course, batch, sub, date, cs)
	if err != nil {
		return "", err
	}
	for _, sId := range aForm.Students {
		if val, ok := studentsMap[sId]; ok {
			studentsLst[val].IsPresent = true
		}
	}
	finRes := make(map[string]interface{})
	finRes["students"] = studentsLst
	finRes["time_slot"] = aForm.TimeSlot
	finRes["work_log"] = aForm.WorkLog
	res, err := json.Marshal(finRes)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func getStudents(college string, course string, batch string, cs string) ([]attendance.Student, error) {
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
	fmt.Println(item)
	var res string
	err = attributevalue.Unmarshal(item, &res)
	var parsedRes []attendance.Student
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
	currentTs := time.Now().Unix()
	tknTs, _ := strconv.ParseInt(tknParts[0], 10, 64)
	if tknTs <= currentTs {
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

func respondData(response string, err error) (events.APIGatewayProxyResponse, error) {
	if err != nil {
		return respondError(err.Error())
	}
	return events.APIGatewayProxyResponse{Body: response, StatusCode: 200, IsBase64Encoded: true, Headers: ResponseHeaders}, nil
}

func respondData201(response string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{Body: response, StatusCode: 201, IsBase64Encoded: true, Headers: ResponseHeaders}, nil
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
	roleArray := claims[cfg_details.ALLOWED_URL+"roles"].([]interface{})
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
