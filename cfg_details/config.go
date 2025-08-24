package cfg_details

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	s3v2 "github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
	"github.com/oklog/ulid"
)

const (
	//Auth Consfigs
	CLAIM_ISS       = "https://dev-q0ywv1av1mdo8z4n.us.auth0.com/"
	CLAIM_CLIENT_ID = "t4pVw4sPvFWgvq4n3DgvFGnavMatqYwv"
	API_URL         = "https://dev-q0ywv1av1mdo8z4n.us.auth0.com/api/v2"
	TOKEN           = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IlNwb3ZQcXFBUFVKbWljZjdMRFkxViJ9.eyJpc3MiOiJodHRwczovL2Rldi1xMHl3djFhdjFtZG84ejRuLnVzLmF1dGgwLmNvbS8iLCJzdWIiOiJVNXZHa0lmUHJiU3pGR3RtVkRVb2hwYTEyS3pvR0ZiOUBjbGllbnRzIiwiYXVkIjoiaHR0cHM6Ly9kZXYtcTB5d3YxYXYxbWRvOHo0bi51cy5hdXRoMC5jb20vYXBpL3YyLyIsImlhdCI6MTc0NjIwMTAwOSwiZXhwIjoxNzQ2ODA1ODA5LCJzY29wZSI6InJlYWQ6Y2xpZW50X2dyYW50cyB1cGRhdGU6dXNlcnMgZGVsZXRlOnVzZXJzIGNyZWF0ZTp1c2VycyByZWFkOnJvbGVzIGNyZWF0ZTpyb2xlX21lbWJlcnMgcmVhZDpyb2xlX21lbWJlcnMgY3JlYXRlOmNsaWVudF9jcmVkZW50aWFscyIsImd0eSI6ImNsaWVudC1jcmVkZW50aWFscyIsImF6cCI6IlU1dkdrSWZQcmJTekZHdG1WRFVvaHBhMTJLem9HRmI5In0.cHah2rHXTQ5rebLNNgssF6D6e8eJW7zm8gSWdK8DA7k_vSvNUA_i_m5J0LNLs6X4QgJDknJFFMK9gyAenwS7YAzrK4uTgxw49kn8WuZ_uktiMFzROeHt24oBwBH8EVXlWNJ2I_nhLOD0ejGiSQIFC3WQYFBFI0rvld4WpRlCHgpkUTHzt_sWft7ggDYXIrILC3IrBzjIWkc_QoDFFTwnD8vPpUNWYHYxoSiMMgMtSa6BWmM6dd597jUz2SLPxja7DRFKR6a_ClQw3RD7jNQbluhM4huZvOa8XWdY3dH95VsNJCTVwgYH9GQepiEZZL8dE4NEX_QkNLagM6uQEsHVOw"
	FACULTY_ROLE    = "rol_lBEDS47aRUeh4jUs"
	STUDENT_ROLE    = "rol_lPXVvRU2Qwcu1Ocd"
	COUNSELOR_ROLE  = "rol_6IUAefHzgyUuPLIN"
	ADMIN_ROLE      = "rol_xNJlq9ki29STAJwF"
	ALLOWED_URL     = "https://d1lcyxhbs0jrme.cloudfront.net/"
	JWKS            = "https://dev-q0ywv1av1mdo8z4n.us.auth0.com/.well-known/jwks.json"

	// User related Errors
	INVALID_DATA      = "Invalid User data"
	AUTH0_UNAVAILABLE = "Service Unavailable - 101"

	//Error Codes
	DATA_ERROR           = "Cloud Service Unavailable - "
	CODE_ERROR           = "Code error - "
	INPUT_ERROR          = "invalid input"
	DATA_SERVICE_ERROR   = "Data Service Unavailable"
	AUTH_PROVIDER_PREFIX = "auth0|"

	// Tables
	COLLEGE_METADATA_TABLE = "college_metadata"

	//S3 Buckets
	BUCKET_STUDENTS_FACULTIES = "cerp-students"
	EXPORT_PATH               = "/export"
)

var CFG, _ = config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-south-1"))

// var CFG, _ = config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("mumbai"), config.WithRegion("ap-south-1"))
var DynamoCfg = dynamodb.NewFromConfig(CFG)
var SqsClient = sqs.NewFromConfig(CFG)
var Presigner = s3v2.NewPresignClient(s3v2.NewFromConfig(CFG))
var NA = "NA"
var DATE_LAYOUTS = map[string]string{
	"d-m-yyyy":   "2-1-2006",   // non-padded day-month-year
	"yyyy-mm-dd": "2006-01-02", // ISO 8601
}

func GenerateUserId(key string) string {
	hash := md5.Sum([]byte(key))
	uId := hex.EncodeToString(hash[:])
	return uId
}

func GenerateUserIdV2(key string, colId string) string {
	hash := md5.Sum([]byte(key))
	uId := hex.EncodeToString(hash[:])
	return AUTH_PROVIDER_PREFIX + colId + "|" + uId
}

func GenerateUlid(ts time.Time) string {
	entropy := ulid.Monotonic(rand.New(rand.NewSource(ts.Unix())), 0)
	return ulid.MustNew(ulid.Timestamp(ts), entropy).String()
}

// _ is used as separator
func GetPk(params ...string) string {
	res := ""
	for _, param := range params {
		res = res + strings.ToLower(param) + "_"
	}
	res = strings.Trim(res, "_")
	return res
}

// - is used as separator
func GetPk2(params ...string) string {
	demarcator := "-"
	res := ""
	for _, param := range params {
		res = res + param + demarcator
	}
	res = strings.Trim(res, demarcator)
	return res
}

func GetCurrentDate() string {
	currentDate := time.Now().Format("2006-01-02")
	return currentDate
}

type FileResponse struct {
	URL string `json:"url"`
}

func UnmarshalItem(item map[string]types.AttributeValue) map[string]interface{} {
	var result map[string]interface{}
	_ = attributevalue.UnmarshalMap(item, &result)
	return result
}

func ArchiveTags() string {
	var tagPairs []string
	tagPairs = append(tagPairs, url.QueryEscape("archive")+"="+url.QueryEscape(strconv.FormatBool(true)))
	return strings.Join(tagPairs, "&")
}

func NoTags() string {
	var tagPairs []string
	return strings.Join(tagPairs, "&")
}

func ExpireTags() string {
	var tagPairs []string
	tagPairs = append(tagPairs, url.QueryEscape("expire")+"="+url.QueryEscape(strconv.FormatBool(true)))
	return strings.Join(tagPairs, "&")
}

func ExpireS3Tag() []s3types.Tag {
	tags := []s3types.Tag{
		{
			Key:   aws.String("expire"),
			Value: aws.String("true"),
		},
	}
	return tags
}

func GetSecretCfg(secretKey string, name string, region string) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return "", err
	}

	svc := secretsmanager.NewFromConfig(cfg)

	result, err := svc.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(name),
	})
	if err != nil {
		log.Printf("Error in fetching the secret values %s %v\n", secretKey, err)
		return "", err
	}

	var secretMap map[string]string
	err = json.Unmarshal([]byte(*result.SecretString), &secretMap)
	if err != nil {
		log.Printf("Error in fetching the secret values unmarshaling %s %v\n", secretKey, err)
		return "", err
	}
	res, ok := secretMap[secretKey]
	if !ok {
		log.Printf("No configurations found for %s %s\n", secretKey, name)
		return "", errors.New("No configurations found for " + secretKey + " " + name)
	}
	return res, nil
}

func GetTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func GenerateTtl(timeInMin int) string {
	// log.Println(time.Now().Add(time.Duration(timeInMin) * time.Minute).Unix())
	return strconv.FormatInt(time.Now().Add(time.Duration(timeInMin)*time.Minute).Unix(), 10)
}

func GetCurrentTs() int64 {
	return time.Now().Unix()
}

func GetCurrentTsStr() string {
	return strconv.FormatInt(GetCurrentTs(), 10)
}

func ParseStrToInt64(inputStr string) int64 {
	val, err := strconv.ParseInt(inputStr, 10, 64)
	if err != nil {
		return -1
	}
	return val
}

func ConvertSingleQuoteString(inputStr string) string {
	parts := strings.Split(inputStr, ",")
	for i, val := range parts {
		parts[i] = "'" + val + "'"
	}
	result := strings.Join(parts, ",")
	return result
}

func GenerateUUID() string {
	id := uuid.New()
	return id.String()
}

func GenerateHash(obj interface{}) (string, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	// Step 2: Hash using SHA-256
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func ParseDateStr(rawDate string) string {
	parts := strings.Split(rawDate, "-")
	if len(parts) == 3 && len(parts[2]) == 3 {
		parts[2] = parts[2][1:] // remove the leading zero
	}
	return strings.Join(parts, "-")
}

func IsDateInRange(date time.Time, from time.Time, to time.Time) bool {
	return !date.Before(from) && !date.After(to)
}

func DetectAndParseDate(dateStr string) (time.Time, error) {
	for _, layout := range DATE_LAYOUTS {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unknown date format: %s", dateStr)
}
