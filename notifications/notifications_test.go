package notifications

import (
	"cerpApi/cfg_details"
	"testing"
)

func TestSendOtpTypeVault(t *testing.T) {
	ts := cfg_details.GenerateTtl(10)
	err := SendOtp("2025-2028-BCA-1", "vault", ts, "test", "test@test.com", "321434")
	if err != nil {
		t.Fail()
	}
}
