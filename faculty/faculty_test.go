package faculty

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestUpdateFileMeta1(t *testing.T) {
	formMap := make(map[string]string)
	formMap["offer_letter"] = `
	{
		"type" : "pdf",
		"key": "offer_letter.pdf",
		"uploaded": "true"
	}
	`
	UpdateFileMeta("ni", formMap, "testUser", "testUser")
}

func TestUpdateFileMeta2(t *testing.T) {
	formMap := make(map[string]string)
	formMap["offer_letter2.pdf"] = `
	{
		"type" : "pdf",
		"key": "offer_letter2.pdf",
		"uploaded": "true"
	}
	`
	res, err := UpdateFileMeta("ni", formMap, "auth0|ni|c94fdacdcd80d944abd5f0e4ca1820ed", "testUser")
	if err != nil {
		t.Fail()
		log.Printf("Error is %v\n", err)
	}
	fmt.Println(res)
}

func TestRemoveFacultyFileFromS3Case1(t *testing.T) {
	s3Client := s3.NewFromConfig(CFG)
	err := RemoveFacultyFileFromS3(s3Client, "ni", "auth0|ni|c94fdacdcd80d944abd5f0e4ca1820ed", "resume.07")
	if err != nil {
		t.Fail()
		log.Printf("Error is %v\n", err)
	}
	fmt.Println("Deleted!!!")
}

func TestDeleteFacultyFile(t *testing.T) {
	s3Client := s3.NewFromConfig(CFG)
	res, err := DeleteFacultyFile(s3Client, "ni", "auth0|ni|c94fdacdcd80d944abd5f0e4ca1820ed", "offer_letter2.pdf", "test-user")
	if err != nil {
		t.Fail()
		log.Printf("Error is %v\n", err)
	}
	fmt.Printf("Deleted!!! %s\n", res)
}

func TestDownloadFacultyFile(t *testing.T) {
	url, err := DownloadFacultyFile("ni", "auth0|ni|c94fdacdcd80d944abd5f0e4ca1820ed", "photo.jpeg", nil)
	if err != nil {
		t.Fail()
	}
	log.Println(url)
}
