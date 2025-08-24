package cfg_details

import (
	"crypto/md5"
	"encoding/hex"
	"log"
	"strings"
	"testing"
	"time"
)

func TestGenerateUlid(t *testing.T) {
	ulid := GenerateUlid(time.Now())
	if ulid == "" {
		t.Fail()
	}
}

func TestGenerateUserId(t *testing.T) {
	key := "ABC|526748|hyfh"
	hash := md5.Sum([]byte(key))
	uId := hex.EncodeToString(hash[:])
	if uId != GenerateUserId(key) {
		t.Fail()
	}
}

func TestGetSecretCfg(t *testing.T) {
	res, err := GetSecretCfg("token", "ni", "ap-south-1")
	if err != nil {
		log.Printf("Error test case failed %v\n", err)
		t.Fail()
	}
	log.Printf("Creds are %s\n", res)
}

func TestConvertSingleQuoteString(t *testing.T) {
	res := ConvertSingleQuoteString("2025-2028-BCA-5,2025-2028-BCA-10,2025-2028-BCA-1")
	if !strings.Contains(res, "'") {
		t.Fail()
	}
	log.Println(res)
}

func TestGenerateUUID(t *testing.T) {
	log.Println(GenerateUUID())
}
