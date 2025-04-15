package cfg_details

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	s3v2 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/oklog/ulid"
)

const (
	//Auth Consfigs
	CLAIM_ISS       = "https://dev-q0ywv1av1mdo8z4n.us.auth0.com/"
	CLAIM_CLIENT_ID = "t4pVw4sPvFWgvq4n3DgvFGnavMatqYwv"
	API_URL         = "https://dev-q0ywv1av1mdo8z4n.us.auth0.com/api/v2"
	TOKEN           = ""
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
)

var CFG, _ = config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-south-1"))

// var CFG, _ = config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("mumbai"), config.WithRegion("ap-south-1"))
var DynamoCfg = dynamodb.NewFromConfig(CFG)
var SqsClient = sqs.NewFromConfig(CFG)
var Presigner = s3v2.NewPresignClient(s3v2.NewFromConfig(CFG))

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

func GetCurrentDate() string {
	currentDate := time.Now().Format("2006-01-02")
	return currentDate
}
