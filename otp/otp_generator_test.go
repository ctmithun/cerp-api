package otp

import (
	"log"
	"testing"
)

func TestGenerateOtp(t *testing.T) {
	otp, err := GenerateOtp("2025-2028-BCA-1", "vault")
	if err != nil {
		t.Fail()
	}
	log.Println(otp)
}
