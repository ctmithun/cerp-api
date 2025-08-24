package uv

import (
	"cerpApi/cfg_details"
	"encoding/json"
	"log"
	"testing"
)

func TestGetUvDocs(t *testing.T) {
	colId := "ni"
	docs := GetUvDocs(colId)
	if len(docs) == 0 {
		log.Printf("Docs not available for the college %s\n", colId)
		t.Fail()
	}
	log.Println(docs)
}

func TestOnboardUvDocs(t *testing.T) {
	students := make([]string, 0)
	s1 := map[string]string{
		"id":   "2025-2028-BCA-5",
		"name": "Sample Kumar",
	}
	s1Str, err := json.Marshal(s1)
	if err != nil {
		t.Fail()
	}
	students = append(students, string(s1Str))
	err = OnboardUvDocs("ni", "2025-2028_bca_sem1", DOC_TYPE[0], students, "test-user")
	if err != nil {
		t.Fail()
	}
}

func TestGenerateHash(t *testing.T) {
	obj := getCollectDocumentWrapper("2025-2028_bca_sem1", "MC", "2025-2028-BCA-1", "Collected document", "test-user")
	hash1, err := cfg_details.GenerateHash(obj)
	if err != nil {
		t.Fail()
		log.Printf("Error generating hash %v\n", err)
	}
	log.Printf("Hash1 %s\n", hash1)
	obj2 := getCollectDocumentWrapper("2025-2028_bca_sem1", "MC", "2025-2028-BCA-1", "Collected document", "test-user")
	hash2, err := cfg_details.GenerateHash(obj2)
	if err != nil {
		t.Fail()
		log.Printf("Error generating hash %v\n", err)
	}
	log.Printf("Hash2 %s\n", hash2)
	if condition := hash1 != hash2; condition {
		t.Fail()
		log.Printf("Hash mismatch: %s != %s\n", hash1, hash2)

	}
}

func TestCollectDocment(t *testing.T) {
	otp, err := SendOtpVerification("ni", "2025-2028_bca_sem1", "MC", "2025-2028-BCA-2", "Collected document", "test-user")
	if err != nil {
		t.Fail()
	}
	log.Printf("OTP is %s\n", otp)
	err = CollectDocument("ni", "2025-2028_bca_sem1", "MC", "2025-2028-BCA-2", "Collected document", otp, "test-user")
	if err != nil {
		t.Fail()
		log.Printf("Error in collecting document %v\n", err)
	}
}

func TestSendOtpVerification(t *testing.T) {
	otp, err := SendOtpVerification("ni", "2025-2028_bca_sem1", "MC", "2025-2028-BCA-1", "", "test-user")
	if err != nil {
		t.Fail()
	}
	log.Println("OTP is ", otp)
}
