package onboard_data

import (
	"log"
	"testing"
)

func TestOnboardS2S(t *testing.T) {
	reqBody := `
{
		"HIN-24BCA": "U183432847, U183432848, U183432849",
		"KAN-24BCA": "U182432847, U182432848, U182432849"
}		
	`
	res, err := OnboardS2S("ni", "BCA", "2022-2025_SEM-6", "A", "TestUser", []byte(reqBody))
	if err != nil {
		log.Printf("OnboardS2S error %v\n", err)
		t.Fail()
	}
	log.Println(res)
}

func TestGetS2S(t *testing.T) {
	res, err := GetS2S("ni", "BCA", "2022-2025_SEM-6", "A")
	if err != nil {
		log.Printf("OnboardS2S error %v\n", err)
		t.Fail()
	}
	log.Println(res)
}

func TestGetS2SNoData(t *testing.T) {
	res, err := GetS2S("ni", "BCA", "2023-2026_SEM-4", "A")
	if err != nil {
		log.Printf("OnboardS2S error %v\n", err)
		t.Fail()
	}
	if res != "{}" {
		t.Fail()
	}
	log.Println(res)
}

func TestGetS2SPerSub1(t *testing.T) {
	res, err := GetS2SPerSub("ni", "PUC-R", "2024-2026_2ND-YR", "A", "CS-2025-PUC-R")
	if err != nil {
		log.Printf("No data found %v\n", err)
		t.Fail()
	}
	if res == nil {
		log.Printf("No data found")
	}
	log.Println(res)
}

func TestGetS2SPerSub2NoData(t *testing.T) {
	res, err := GetS2SPerSub("ni", "PUC-R", "2024-2026_1ST-YR", "A", "CS-2025-PUC-R")
	if err != nil {
		log.Printf("No data found %v\n", err)
		t.Fail()
	}
	if res == nil {
		log.Printf("No data found")
	}
	log.Println(res)
}
