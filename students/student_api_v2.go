package students

import (
	"cerpApi/cfg_details"
	"log"
	"strconv"
	"strings"
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
	yoj := strings.Split(student.Batch, "-")[0]
	log.Printf("Creating the student for the yoj number - %s\n", yoj)
	PKKey := student.Batch + "-" + student.Branch
	sIdParts := strings.Split(student.Sid, "-")
	SKKey, err := strconv.Atoi(sIdParts[len(sIdParts)-1])
	if err != nil {
		log.Printf("Error converting SKKey to int via updating student record %v\n", err)
		return "", err
	}
	err = updateStudentRecord(college, student, PKKey, SKKey, uBy)
	return student.Sid, err
}
