package students

import (
	"cerpApi/cfg_details"
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
)

const UT_USER = "UTUser"
const CONN_STR = "postgresql://cerp:7pFJwJHKWvWwIRycQ9yXew@weary-flapper-8111.j77.aws-ap-south-1.cockroachlabs.cloud:26257/cerp?sslmode=verify-full"

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

// func TestGetStudentVault(t *testing.T) {
// 	GetStudentsVault("ni", "BCA", "2025-2028")
// 	t.Fail()
// }

func TestGetStudentsVault(t *testing.T) {
	var pgOnce sync.Once
	pgOnce.Do(func() {
		ctx := context.Background()
		db, err := pgxpool.New(ctx, CONN_STR)
		if err != nil {
			log.Printf("unable to create connection pool: %v", err)
			return
		}
		studs, err := GetStudentsVault(db, "ni", "'2025-2028-BCA-5','2025-2028-BCA-10','2025-2028-BCA-1','2025-2028-BCA-6','2025-2028-BCA-2','2025-2028-BCA-8','2025-2028-BCA-3','2025-2028-BCA-4','2025-2028-BCA-9','2025-2028-BCA-7'")
		db.Close()
		if err != nil {
			fmt.Println(err.Error())
			t.Fail()
		}
		if strings.Contains(studs, "T02") {
			t.Fail()
		}
		log.Println(studs)
	})
}

func TestSaveStudentVault(t *testing.T) {

	docs := StudentDocs{
		Sid: "T03",
		Docs: []VaultDoc{
			{DocId: 1, Entry: 1, DocType: "PDF", DocName: "Doc One", CollectedDate: "2024-01-01 12:15:12", Comment: "Test doc"},
			{Id: 11, DocId: 3, Entry: 1, DocType: "Image", DocName: "Doc Two", CollectedDate: "2024-01-15 13:12:21", Comment: "Another doc"},
		},
	}

	var pgOnce sync.Once
	pgOnce.Do(func() {
		ctx := context.Background()
		db, err := pgxpool.New(ctx, CONN_STR)
		if err != nil {
			log.Printf("unable to create connection pool: %v", err)
			return
		}
		err = SaveVault(db, "ni", UT_USER, docs, "")
		db.Close()
		if err != nil {
			fmt.Println(err.Error())
			t.Fail()
		}
		log.Printf("ID is %s\n", docs.Sid)
	})
	// log.Printf("Test finished")
}

func TestGetStudentsDataV2S1(t *testing.T) {
	res := GetStudentsDataV2("ni", "2025-2028", "BCA")
	if len(res) == 0 {
		t.Fail()
	}
	log.Printf("Results are res - %v\n", res[0].UpdatedBy)
}

func getStudentDocs() StudentDocs {
	docs := StudentDocs{
		Sid: "2025-2028-BA-11",
		Docs: []VaultDoc{
			{DocId: 1, Entry: 1, DocType: "PDF", DocName: "Doc One", CollectedDate: "2024-01-01 13:15:12", Comment: "Test doc"},
			{Id: 11, DocId: 3, Entry: 1, DocType: "Image", DocName: "Doc Two", CollectedDate: "2024-01-15 13:12:21", Comment: "Another doc"},
		},
	}
	return docs
}

func TestGenerateOtp(t *testing.T) {
	docs := getStudentDocs()
	err := GenerateOtp("ni", docs, "UT-USER")
	if err != nil {
		log.Printf("Error in test case for %v\n", err)
		t.Fail()
	}
}

func TestFetchSavedOtp(t *testing.T) {
	mapData, err := fetchSavedOtp("ni", "2025-2028-BCA-1")
	if err != nil {
		t.Fail()
	}
	log.Println(mapData)
}

func TestVerifyOtp(t *testing.T) {
	isOk, err := verifyOtp("ni", "2025-2028-BCA-1", "986705", getStudentDocs())
	if !isOk || err != nil {
		t.Fail()
	}
}

func TestGetStudentEmailById(t *testing.T) {
	email, err := GetStudentEmailById("ni", "2025-2026-PUC-R-HEBA-1")
	if err != nil {
		t.Fail()
	}
	log.Println(email)
}

func TestGetStudentVault(t *testing.T) {
	var pgOnce sync.Once
	pgOnce.Do(func() {
		ctx := context.Background()
		db, err := pgxpool.New(ctx, CONN_STR)
		if err != nil {
			log.Printf("unable to create connection pool: %v", err)
			return
		}
		docs, err := GetStudentVault(db, "ni", "2025-2028-BCA-10")
		defer db.Close()
		if err != nil {
			fmt.Println(err.Error())
			t.Fail()
		}
		log.Printf("ID is %v\n", docs)
	})
}
