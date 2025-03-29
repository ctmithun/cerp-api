package cfg_details

import (
	"crypto/md5"
	"encoding/hex"
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
