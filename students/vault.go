package students

import (
	"cerpApi/cfg_details"
	"cerpApi/notifications"
	"cerpApi/otp"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//	type Vault struct {
//		Student Student `json:"student"`
//		Vault   string  `json:"vault"`
//	}
const OTP_TYPE = "vault"

type StudentDocs struct {
	Sid  string     `json:"sid"`
	Docs []VaultDoc `json:"docs"`
}

type VaultDoc struct {
	Id            int64  `json:"id"`
	DocId         int16  `json:"doc_id"`
	Entry         int16  `json:"entry"`
	DocType       string `json:"doc_type"`
	DocName       string `json:"doc_name"`
	CollectedDate string `json:"collected_date"`
	ReturnedDate  string `json:"returned_date"`
	Comment       string `json:"comment"`
}

func GetStudentVault(con *pgxpool.Pool, colId string, sid string) (string, error) {
	tableName := getVaultTable(colId)
	ctx, cancel := cfg_details.GetTimeoutContext()
	defer cancel()
	rows, err := con.Query(ctx, "select entry, id, doc_id, doc_type, comment, TO_CHAR(collected_date, 'YYYY-MM-DD'), COALESCE(TO_CHAR(returned_date, 'YYYY-MM-DD'), '') from "+tableName+" where sid in ('"+sid+"')", pgx.QueryExecModeSimpleProtocol)
	if err != nil {
		log.Println("Error querying the Vault table for fetching - " + err.Error())
		return "", err
	}
	var vDocs []VaultDoc
	for rows.Next() {
		var vDoc VaultDoc
		err := rows.Scan(&vDoc.Entry, &vDoc.Id, &vDoc.DocId, &vDoc.DocType, &vDoc.Comment, &vDoc.CollectedDate, &vDoc.ReturnedDate)
		if err != nil {
			log.Printf("Error querying the vault table for - %s %v\n ", sid, err.Error())
			continue
		}
		vDocs = append(vDocs, vDoc)
	}
	if err != nil {
		return "", err
	}
	studentDocs := &StudentDocs{
		Sid:  sid,
		Docs: vDocs,
	}
	studentDocsJsonStr, err := json.Marshal(studentDocs)
	if err != nil {
		log.Printf("Error marshaling the output in GetStudentVault for the vault of the student-%s - Error - %v\n ", sid, err)
		return "", err
	}
	return string(studentDocsJsonStr), nil
}

func getVaultTableName(colId string) string {
	return "students.vault_" + colId
}

func GetStudentsVault(con *pgxpool.Pool, colId string, students string) (string, error) {
	tableName := getVaultTable(colId)
	ctx, cancel := cfg_details.GetTimeoutContext()
	defer cancel()
	rows, err := con.Query(ctx, "select distinct sid from "+tableName+" where sid in ("+students+")")
	if err != nil {
		log.Println("Error querying the Eq table for fetching - " + err.Error())
	}
	resMap := make(map[string]string, 0)
	for rows.Next() {
		var studentId string
		err := rows.Scan(&studentId)
		if err != nil {
			log.Println("Error querying the vault table for fetching students %v\n" + err.Error())
			continue
		}
		resMap[studentId] = ""
	}
	resStr, err := json.Marshal(resMap)
	if err != nil {
		return "", err
	}
	return string(resStr), err
}

func getVaultTable(colId string) string {
	return "students.vault_" + colId
}

func SaveVault(con *pgxpool.Pool, colId string, uBy string, studentDocs StudentDocs, otp string) error {
	sid := studentDocs.Sid
	isOk, err := verifyOtp(colId, sid, otp, studentDocs)
	if !isOk || err != nil {
		return errors.New("otp is invalid, retry")
	}
	log.Printf("Saving the vault for the student %s by %s\n", sid, uBy)
	tableName := getVaultTableName(colId)
	ctx, cancel := cfg_details.GetTimeoutContext()
	defer cancel()
	tx, err := con.Begin(ctx)
	if err != nil {
		return err
	}
	for _, doc := range studentDocs.Docs {
		if doc.Id == 0 {
			var insertedID int64
			err = tx.QueryRow(ctx, `INSERT INTO `+tableName+`(sid, doc_id, entry, doc_type, doc_name, comment, u_by, collected_date) 
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT DO NOTHING RETURNING id
        `, studentDocs.Sid, doc.DocId, doc.Entry, doc.DocType, doc.DocName, doc.Comment, uBy, doc.CollectedDate).Scan(&insertedID)
			if err != nil {
				log.Printf("Error while inserting the new document %s %v\n", doc.DocType, err)
				break
			}
		} else if doc.ReturnedDate != "" {
			_, err = tx.Exec(ctx, `UPDATE `+tableName+` SET returned_date = $1 WHERE returned_date IS DISTINCT FROM $1 and id = $2`, doc.ReturnedDate, doc.Id)
		}
		if isErr := dbError(err, doc); isErr != nil {
			log.Printf("Error in db exec %v\n", err)
			return err
		}
	}
	if isErr := tx.Commit(ctx); isErr != nil {
		log.Printf("Error in db commit %v\n", isErr)
		return isErr
	}
	log.Printf("Saved the vault for the student %s updated by %s\n", sid, uBy)
	return nil
}

func verifyOtp(colId, sId, otp string, studentDocs StudentDocs) (bool, error) {
	b, err := json.Marshal(studentDocs)
	if err != nil {
		log.Printf("Error while marshaling the studentdocs in verifyOtp %v\n", err)
		return false, err
	}
	hashStr, err := generateHash(b)
	if err != nil {
		return false, err
	}
	otpDetails, err := fetchSavedOtp(colId, sId)
	if err != nil {
		return false, err
	}
	currentTtl := cfg_details.GetCurrentTs()
	otpTtl := cfg_details.ParseStrToInt64(otpDetails["ttl"])
	if currentTtl > otpTtl {
		mes := "OTP expired"
		log.Printf("%s\n", mes)
		return false, errors.New(mes)
	}
	if otp != otpDetails["otp"] {
		mes := "OTP didn't match"
		log.Printf("%s\n", mes)
		return false, errors.New(mes)
	}
	if hashStr != otpDetails["hash"] {
		mes := "Content changed otp didn't match"
		log.Printf("%s\n", mes)
		return false, errors.New(mes)
	}
	return true, nil
}

func fetchSavedOtp(colId string, sId string) (map[string]string, error) {
	key, err := attributevalue.Marshal(sId)
	if err != nil {
		log.Printf("Error in fetchSavedOtp while marshaling the key %s %v\n", sId, err)
		return nil, err
	}
	sKey, err := attributevalue.Marshal("vault")
	if err != nil {
		log.Printf("Error in fetchSavedOtp while marshaling vault skey - %v\n", err)
		return nil, err
	}
	ck := map[string]types.AttributeValue{
		"sid":      key,
		"otp_type": sKey,
	}
	ctx, cancel := cfg_details.GetTimeoutContext()
	defer cancel()
	out, err := cfg_details.DynamoCfg.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(colId + "_otp"),
		Key:       ck,
	})
	if err != nil {
		log.Printf("Error in fetchSavedOtp while querying the DDB %v\n", err)
		return nil, err
	}
	if out == nil || len(out.Item) == 0 {
		return nil, nil
	}
	mapOtpData := make(map[string]string)
	mapOtpData["hash"] = out.Item["hash"].(*types.AttributeValueMemberS).Value
	mapOtpData["otp"] = out.Item["otp"].(*types.AttributeValueMemberS).Value
	mapOtpData["ttl"] = out.Item["ttl"].(*types.AttributeValueMemberS).Value
	return mapOtpData, nil
}

func dbError(err error, doc VaultDoc) error {
	if err != nil {
		if strings.Contains(err.Error(), "\"unique_sid_docid_entry\"") {
			log.Printf("Document already present %s\n", doc.DocType)
			return nil
		} else {
			log.Printf("Error creating the Vault for the doc %v\n", err)
			return err
		}
	}
	return nil
}

func generateHash(content []byte) (string, error) {
	h := sha256.Sum256(content)
	hashStr := hex.EncodeToString(h[:])
	return hashStr, nil
}

func GenerateOtp(colId string, studentDocs StudentDocs, uBy string) error {
	b, err := json.Marshal(studentDocs)
	if err != nil {
		log.Printf("Error in GenerateHash while marshaling the studentdocs %v\n", err)
		return err
	}
	hashStr, err := generateHash(b)
	if err != nil {
		return err
	}
	otpStr, err := otp.GenerateOtp(hashStr, "vault")
	log.Printf("OTP generated for the student - %s by %s\n", studentDocs.Sid, uBy)
	if err != nil {
		return err
	}
	return insertToDb(colId, studentDocs.Sid, string(b), hashStr, otpStr)
}

func insertToDb(colId string, sId string, content string, hashVal string, otpStr string) error {
	// otpData := make(map[string]string)
	// otpData["sid"] = sId
	// otpData["otp_type"] = OTP_TYPE
	// otpData["otp"] = otpStr
	// otpData["hash"] = hashVal
	// otpData["ttl"] = cfg_details.GenerateTtl(10)
	// data, err := attributevalue.MarshalMap(otpData)
	// log.Println(otpData)
	// if err != nil {
	// 	log.Printf("Error in insertToDb while marshaling the data %v\n", err)
	// 	return err
	// }
	// ctx, cancel := cfg_details.GetTimeoutContext()
	// defer cancel()
	// _, err = cfg_details.DynamoCfg.PutItem(ctx, &dynamodb.PutItemInput{
	// 	TableName: aws.String(colId + "_otp"),
	// 	Item:      data,
	// })
	// if err != nil {
	// 	log.Printf("Error while writing to DB for the otp %s %v\n", sId, err)
	// 	return err
	// }
	ttl := cfg_details.GenerateTtl(10)
	otp.PersistOtp(colId, sId, hashVal, content, otpStr, OTP_TYPE, ttl)
	email, err := GetStudentEmailById(colId, sId)
	if err != nil {
		return err
	}
	return notifications.SendOtp(sId, OTP_TYPE, ttl, content, email, otpStr)
}
