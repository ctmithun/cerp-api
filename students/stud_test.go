package students

import (
	"fmt"
	"log"
	"testing"
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
