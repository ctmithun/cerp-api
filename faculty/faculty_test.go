package faculty

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
)

const CONN_STR = "postgresql://cerp:7pFJwJHKWvWwIRycQ9yXew@weary-flapper-8111.j77.aws-ap-south-1.cockroachlabs.cloud:26257/cerp?sslmode=verify-full"

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
	formMap["offer_letter-2.pdf"] = `
	{
		"type" : "pdf",
		"key": "offer_letter-2.pdf",
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

func TestGetFacultyRoles(t *testing.T) {
	roles := "faculty,counselor"
	rolesLst := strings.Split(roles, ",")
	res := getRoles(rolesLst)
	log.Printf("%v\n", res)
}

func TestUpdateProfileTag(t *testing.T) {
	s3Client := s3.NewFromConfig(CFG)
	faculty := Faculty{
		Photo: "1169021b-b8d0-47ff-948d-834fbd57f351.png",
	}
	err := UpdateProfileTag(s3Client, "ni", faculty.Photo)
	if err != nil {
		t.Fail()
	}
}

func TestGetEmpNo(t *testing.T) {
	var pgOnce sync.Once
	pgOnce.Do(func() {
		ctx := context.Background()
		db, err := pgxpool.New(ctx, CONN_STR)
		if err != nil {
			log.Printf("unable to create connection pool: %v\n", err)
			return
		}
		empNo := getEmpNo(db, "ni", "auth0|ni|10cf752de18764243e9e4cd4c61bec5f")
		db.Close()
		if empNo == "" {
			fmt.Println("Employee doesn't exist")
			t.Fail()
			return
		}
		fmt.Println(empNo)
	})
}

func TestSetEmpNo(t *testing.T) {
	var pgOnce sync.Once
	pgOnce.Do(func() {
		ctx := context.Background()
		db, err := pgxpool.New(ctx, CONN_STR)
		if err != nil {
			log.Printf("unable to create connection pool: %v\n", err)
			return
		}
		records := [][]string{
			// {"auth0|ni|10cf752de18764243e9e4cd4c61bec5f", "2023-08-05"},
			// {"auth0|ni|f6d4f3ad943bd2bf0eb5d25bcc1312d8", "2016-08-01"},
			// {"auth0|ni|8732828446d275b74d06b8264b52213d", "2025-06-23"},
			// {"auth0|ni|ecb3b4a7d9a65846d992b0fa17dca6dd", "2025-04-23"},
			// {"auth0|ni|c94fdacdcd80d944abd5f0e4ca1820ed", "2024-09-01"},
			// {"auth0|ni|131b67e086db75f3d9ada4b71fde7faa", "2025-06-23"},
			// {"auth0|ni|afeefc67a499dd9480f36d0536f46ce5", "2016-07-01"},
			// {"auth0|ni|efdb9a30028804f0538d6ba0ec8c1291", "2024-08-02"},
			// {"auth0|ni|d65d1ca6c825b3b5e85035eb93a316c7", "2025-03-01"},
			// {"auth0|ni|5c4ff8dd9ed0a67c2772c33e65f8ce22", "2025-03-02"},
			// {"auth0|ni|a94b3297e8bd846f00604ce2bb62c705", "2023-07-10"},
			// {"auth0|ni|24c92783c5553d7690c5bec1d514d504", "2023-07-01"},
			// {"auth0|ni|f0964ec76f34fab25bd7231f2bee71ff", "2021-04-02"},
			// {"auth0|ni|96b5a3914c9f1a474364f2b2ffdca9a2", "2023-01-01"},
			// {"auth0|ni|cb4817ea96a937f2ec6ee2c9ac01cfe4", "2024-10-01"},
			// {"auth0|ni|89c0a0a6b19b8b1dfad71897230e0ee1", "2022-10-01"},
			// {"auth0|ni|7f23195d09320b49e2f5bdcc70bd2d53", "2015-01-01"},
			// {"auth0|ni|7518f1470d14fbc652e0edbdba2355ae", "2025-06-22"},
			// {"auth0|ni|7a62a40ac9a9f008c89468a1c26c2bf4", "2025-07-01"},
			// {"auth0|ni|e95cb941d463ec4080c5d315a0a32c36", "2025-03-01"},
			// {"auth0|ni|416824b2ac1c32f99d2cfc9d48e2cd8e", "2025-06-01"},
			// {"auth0|ni|e93b35b6090aab48cca54be1df85ac6d", "2025-05-03"},
			// {"auth0|ni|462af153b89273335689fc3dcc676615", "2024-07-02"},
			// {"auth0|ni|30066c61b71cc268696eccea55c8d072", "2024-03-01"},
			// {"auth0|ni|1b3fb36011d1cd63c6945e42da7e35bc", "2025-06-20"},
			// {"auth0|ni|161ccfd2666790ef960b3a64fcc2c168", "2023-01-01"},
			// {"auth0|ni|ce5a70bd1bb06d9c69c8f6523ddd781b", "2025-04-24"},
			// {"auth0|ni|6562d3f9c6619c112c4817f725379711", "2023-06-01"},
			// {"auth0|ni|f76c4653949365eb79f42fbba57a08fd", "2025-05-15"},
			// {"auth0|ni|cc8a37ed6f8bc5539dc9c29953e5da4c", "2021-01-01"},
			// {"auth0|ni|2f79ae5bb4fea2f889f2dd7707f1a8da", "2023-02-01"},
			// {"auth0|ni|8f7b05c6192a205a890baf5d1f94e2c2", "2025-05-19"},
			// {"auth0|ni|4df8c35d97d9b2a74ccb03a1acc99ada", "2025-05-19"},
		}
		sort.Slice(records, func(i, j int) bool {
			return records[i][1] < records[j][1]
		})
		for _, record := range records {
			log.Println(record[0], record[1])
			empNo := setEmpNo(db, "ni", record[0], record[1])
			if empNo == "" {
				fmt.Println("Employee doesn't exist")
				t.Fail()
				return
			}
			log.Printf("Generated employee number for %s - %s %s\n", record[0], empNo, record[1])
		}
		for i := 0; i < len(records); i++ {

		}
		db.Close()
	})
}

func TestExportForIdGeneration(t *testing.T) {
	s3Client := s3.NewFromConfig(CFG)
	fileLink, err := ExportForIdGeneration(s3Client, "ni", []string{
		"auth0|ni|10cf752de18764243e9e4cd4c61bec5f",
		"auth0|ni|f6d4f3ad943bd2bf0eb5d25bcc1312d8",
		"auth0|ni|8732828446d275b74d06b8264b52213d",
		"auth0|ni|ecb3b4a7d9a65846d992b0fa17dca6dd",
		"auth0|ni|c94fdacdcd80d944abd5f0e4ca1820ed",
		"auth0|ni|131b67e086db75f3d9ada4b71fde7faa",
		"auth0|ni|afeefc67a499dd9480f36d0536f46ce5",
		"auth0|ni|efdb9a30028804f0538d6ba0ec8c1291",
		"auth0|ni|d65d1ca6c825b3b5e85035eb93a316c7",
		"auth0|ni|5c4ff8dd9ed0a67c2772c33e65f8ce22",
		"auth0|ni|a94b3297e8bd846f00604ce2bb62c705",
		"auth0|ni|24c92783c5553d7690c5bec1d514d504",
		"auth0|ni|f0964ec76f34fab25bd7231f2bee71ff",
		"auth0|ni|96b5a3914c9f1a474364f2b2ffdca9a2",
		"auth0|ni|cb4817ea96a937f2ec6ee2c9ac01cfe4",
		"auth0|ni|89c0a0a6b19b8b1dfad71897230e0ee1",
		"auth0|ni|7f23195d09320b49e2f5bdcc70bd2d53",
		"auth0|ni|7518f1470d14fbc652e0edbdba2355ae",
		"auth0|ni|7a62a40ac9a9f008c89468a1c26c2bf4",
		"auth0|ni|e95cb941d463ec4080c5d315a0a32c36",
		"auth0|ni|416824b2ac1c32f99d2cfc9d48e2cd8e",
		"auth0|ni|e93b35b6090aab48cca54be1df85ac6d",
		"auth0|ni|462af153b89273335689fc3dcc676615",
		"auth0|ni|30066c61b71cc268696eccea55c8d072",
		"auth0|ni|1b3fb36011d1cd63c6945e42da7e35bc",
		"auth0|ni|161ccfd2666790ef960b3a64fcc2c168",
		"auth0|ni|ce5a70bd1bb06d9c69c8f6523ddd781b",
		"auth0|ni|6562d3f9c6619c112c4817f725379711",
		"auth0|ni|f76c4653949365eb79f42fbba57a08fd",
		"auth0|ni|cc8a37ed6f8bc5539dc9c29953e5da4c",
		"auth0|ni|2f79ae5bb4fea2f889f2dd7707f1a8da",
		"auth0|ni|8f7b05c6192a205a890baf5d1f94e2c2",
		"auth0|ni|4df8c35d97d9b2a74ccb03a1acc99ada",
	})
	if err != nil {
		log.Printf("Testcase failed for %v\n", err)
		t.Fail()
	}
	t.Fail()
	log.Println(fileLink)
}

func TestGetProfilePicsFromS3(t *testing.T) {
	s3Client := s3.NewFromConfig(CFG)
	getProfilePicsFromS3(s3Client, "ni", "c1028583-9f3f-4566-ae05-25aaccd068ea.jpeg")
}

func TestResize(t *testing.T) {
	resize("c1028583-9f3f-4566-ae05-25aaccd068ea.jpeg")
}

func TestGetFacultiesData(t *testing.T) {
	GetFacultiesData("ni", "")
}
