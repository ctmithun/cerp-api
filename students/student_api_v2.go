package students

import (
	"cerpApi/cfg_details"
	"context"
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func OnboardStudentV2(college string, student *Student, uBy string) (string, error) {
	uId := cfg_details.GenerateUserIdV2(getStudentIdKey(college, *student), college)
	yoj := strings.Split(student.Batch, "-")[0]
	log.Printf("Creating the student for the yoj number - %s\n", yoj)
	PKKey := student.Batch + "-" + student.Branch
	log.Printf("Creating the student for the Row number - %s\n", PKKey)
	SKKey := getRowNumber(PKKey, college)
	student.Id = uId
	student.Sid = PKKey + "-" + strconv.Itoa(SKKey)
	student.Doj = cfg_details.GetCurrentDate()
	student.Yoa = yoj
	err := persistStudentRecord(college, student, PKKey, SKKey, uBy, false)
	return student.Sid, err
}

func UpdateStudentV2(college string, student *Student, uBy string) (string, error) {
	// yoj := strings.Split(student.Batch, "-")[0]
	// log.Printf("Updating the student for the yoj number - %s\n", yoj)
	PKKey := student.Batch + "-" + student.Branch
	sIdParts := strings.Split(student.Sid, "-")
	SKKey, err := strconv.Atoi(sIdParts[len(sIdParts)-1])
	genId := cfg_details.GenerateUserIdV2(getStudentIdKey(college, *student), college)
	student.Id = genId
	if err != nil {
		log.Printf("Error converting SKKey to int via updating student record %v\n", err)
		return "", err
	}
	err = updateStudentRecord(college, student, PKKey, SKKey, uBy)
	return student.Sid, err
}

func UpdateStudentRegNumsInBulk(college string, rollToUsnMap map[string]string, uBy string) (string, error) {
	var wg sync.WaitGroup
	failedMaps := make(map[string]bool)
	log.Printf("Starting Reg Num update by %s\n", uBy)
	for roll, usn := range rollToUsnMap {
		wg.Add(1)
		go func(roll_ string, usn_ string, failTracker map[string]bool, uBy string) {
			defer wg.Done()
			err := updateStudentRegNum(college, roll_, usn_, uBy)
			if err != nil {
				log.Printf("Failed to update item %v: %v", roll_, err)
				failTracker[roll_] = true
			}
		}(roll, usn, failedMaps, uBy)
	}
	wg.Wait()
	if len(failedMaps) > 0 {
		log.Printf("Reg Nums update has failed for %v attempted by - %s\n", failedMaps, uBy)
		if len(failedMaps) == len(rollToUsnMap) {
			return "Failed", errors.New("fully failed")
		}
		return "Partially Succeeded", errors.New("partially failed")
	} else {
		log.Printf("Reg Nums is successfully updated by - %s\n", uBy)
	}
	return "Succeeded", nil
}

func updateStudentRegNum(college string, roll string, usn string, uBy string) error {
	pk, sk := extractBatchAndRoll(roll)
	log.Printf("Updating the Reg Num for %s Roll-Num %s\n", pk, sk)
	_, err := cfg_details.DynamoCfg.UpdateItem(
		context.TODO(),
		&dynamodb.UpdateItemInput{
			TableName: aws.String(college + "_students"),
			Key: map[string]types.AttributeValue{
				"pk":      &types.AttributeValueMemberS{Value: pk},
				"row_num": &types.AttributeValueMemberN{Value: sk},
			},
			UpdateExpression: aws.String("SET #val.#usn = :student_id, ts = :ts, uBy = :uBy, #sid=:student_id"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":student_id": &types.AttributeValueMemberS{Value: usn},
				":ts":         &types.AttributeValueMemberN{Value: strconv.FormatInt(time.Now().Unix(), 10)},
				":uBy":        &types.AttributeValueMemberS{Value: uBy},
			},
			ExpressionAttributeNames: map[string]string{
				"#val": "value",
				"#usn": "UniversitySeatNumber",
				"#sid": "student_id",
			},
		},
	)
	return err
}

func extractBatchAndRoll(sId string) (string, string) {
	idx := strings.LastIndex(sId, "-")
	if idx == -1 {
		return "", "" // no dash found
	}
	return sId[:idx], sId[idx+1:]
}
