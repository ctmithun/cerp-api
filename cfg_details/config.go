package cfg_details

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	s3v2 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
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

func ExpireTags() string {
	var tagPairs []string
	tagPairs = append(tagPairs, url.QueryEscape("expire")+"="+url.QueryEscape(strconv.FormatBool(true)))
	return strings.Join(tagPairs, "&")
}
