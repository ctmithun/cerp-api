package cfg_details

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

const (
	CLAIM_ISS       = "https://dev-q0ywv1av1mdo8z4n.us.auth0.com/"
	CLAIM_CLIENT_ID = "t4pVw4sPvFWgvq4n3DgvFGnavMatqYwv"
	API_URL         = "https://dev-q0ywv1av1mdo8z4n.us.auth0.com/api/v2"
	TOKEN           = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IlNwb3ZQcXFBUFVKbWljZjdMRFkxViJ9.eyJpc3MiOiJodHRwczovL2Rldi1xMHl3djFhdjFtZG84ejRuLnVzLmF1dGgwLmNvbS8iLCJzdWIiOiJVNXZHa0lmUHJiU3pGR3RtVkRVb2hwYTEyS3pvR0ZiOUBjbGllbnRzIiwiYXVkIjoiaHR0cHM6Ly9kZXYtcTB5d3YxYXYxbWRvOHo0bi51cy5hdXRoMC5jb20vYXBpL3YyLyIsImlhdCI6MTczNjY5NjI2NiwiZXhwIjoxNzM3MzAxMDY2LCJzY29wZSI6InJlYWQ6Y2xpZW50X2dyYW50cyB1cGRhdGU6dXNlcnMgZGVsZXRlOnVzZXJzIGNyZWF0ZTp1c2VycyByZWFkOnJvbGVzIGNyZWF0ZTpyb2xlX21lbWJlcnMgcmVhZDpyb2xlX21lbWJlcnMgY3JlYXRlOmNsaWVudF9jcmVkZW50aWFscyIsImd0eSI6ImNsaWVudC1jcmVkZW50aWFscyIsImF6cCI6IlU1dkdrSWZQcmJTekZHdG1WRFVvaHBhMTJLem9HRmI5In0.K_7YuI3_xY7AMKOI0fY8XUasyiAXLnFkTyPAmjXRV_A0hn4P3sCF5jkRsm5zSB0Mz7Aaou2kWMAOGxVSP-Sv0v2l3go4OZmGcKeI_yKvB0ZKArwjoXK8UZ__j0jDv21WG4X135bFUq8iG482DsgrCWD_qc35SI21KOrlweRKCYGRpLq2v_FNIMo4c5yuiQZHwCkgyOlgv6b8oI2dFnTeFhyAKnaAJC9AmBY_wXxoOrdmD7_jDecxTzqiTLPVBY5OyAgQ0t86BTdoAcLEmPLXFCapDHLY-phP5yIQDg54efYouGfCRP7Q8K9usk3ea_kk4gO9nL9izg09IY-M_3wZsw"
	FACULTY_ROLE    = "rol_lBEDS47aRUeh4jUs"
	STUDENT_ROLE    = "rol_lPXVvRU2Qwcu1Ocd"
	ALLOWED_URL     = "https://d1lcyxhbs0jrme.cloudfront.net/roles"
	JWKS            = "https://dev-q0ywv1av1mdo8z4n.us.auth0.com/.well-known/jwks.json"

	INVALID_DATA      = "Invalid User data"
	AUTH0_UNAVAILABLE = "Service Unavailable"
)

var CFG, _ = config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-south-1"))

// var CFG, _ = config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile("mumbai"), config.WithRegion("ap-south-1"))
var DynamoCfg = dynamodb.NewFromConfig(CFG)

func GenerateUserId(key string) string {
	hash := md5.Sum([]byte(key))
	uId := hex.EncodeToString(hash[:])
	return uId
}
