package students

import (
	"cerpApi/cfg_details"
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestGetStudentsData(t *testing.T) {
	student := GetStudentsData("ni", "2025", "PUC-R")
	fmt.Println(student)
}

func TestUpdateSudentV1(t *testing.T) {
	student := GetStudentsData("ni", "2024-2027", "BCA")[0]
	student.Nationality = "INDIAN"
	UpdateStudentV2("ni", &student, "test-user")
}

func TestGetMetadata(t *testing.T) {
	_, _, res3 := getStructFieldNames(Student{})
	log.Printf("%v \n", res3)
}

func TestUpdateStudentRegNumsInBulk(t *testing.T) {
	usnRollMap := make(map[string]string)
	usnRollMap["2023-2026-BCOM-26"] = "U18GO22C0507"
	res, err := UpdateStudentRegNumsInBulk("ni", usnRollMap, "test-user")
	if err != nil {
		log.Printf("Error is %v\n", err)
		t.Fail()
	}
	log.Printf("Res %s\n", res)
}

func TestGetStructFieldNames(t *testing.T) {
	stud := Student{
		Name: "Mithun",
	}
	a, b, c := getStructFieldNames(stud)
	log.Printf("%v, %v, %v \n", a, b, c)
}

func TestGenerateId(t *testing.T) {
	res := cfg_details.GenerateUserIdV2("S-ni|"+"dtu@test.com"+"_"+"7777777777", "ni")
	fmt.Printf("Res %s\n", res)
}

var CFG, _ = config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-south-1"))

func TestTakeStudentBkupToS3(t *testing.T) {
	s3Client := s3.NewFromConfig(CFG)
	err := takeStudentBkupToS3(s3Client, "ni", "2025-2028-BCA", 4)
	if err != nil {
		log.Printf("Tes failed for %v\n", err)
		t.Fail()
	}
	log.Printf("Test case for TestTakeStudentBkupToS3 passed")
}
