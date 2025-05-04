package enquiry

import (
	"cerpApi/cfg_details"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const UT_USER = "UTUser"
const CONN_STR = "postgresql://cerp:7pFJwJHKWvWwIRycQ9yXew@weary-flapper-8111.j77.aws-ap-south-1.cockroachlabs.cloud:26257/cerp?sslmode=verify-full"

func TestCreateEnquiry(t *testing.T) {
	enq := EnquiryForm{
		Name:           "Harry Potter",
		Course:         "BCOM",
		Mobile:         "5555555555",
		Reference:      "Friend",
		Location:       "BSK 2nd Stage",
		Comments:       []string{"Party seems interested after 1st counselling"},
		Status:         "In Process",
		CouncillorId:   cfg_details.GenerateUserId("ni|asd@asd.com|213324324324|"),
		CouncillorName: "Varun J",
	}
	err := CreateEnquiry("ni", &enq, UT_USER)
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
	}
}

func TestGetEnquiries(t *testing.T) {
	enquiries, err := GetActiveEnquiries("ni", "")
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
	}
	fmt.Println(enquiries)
}

func TestGetActiveEnquiriesWithPagination(t *testing.T) {
	enquiries, err := GetActiveEnquiries("ni", "01JJ7AVXRF5BYXPTZ9CGCFB7D9")
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
	}
	fmt.Println(enquiries)
}

func TestUpdateActiveEnquiry(t *testing.T) {
	err := UpdateActiveEnquiry("ni", "01JJ8R5ASPDQ25W0AVH0K23JAB", "test comment-2 - "+time.Now().String(), UT_USER)
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
	}
	fmt.Println("Update Successful - ")
}

func TestDeactivateEnquiryWithEmptyKey(t *testing.T) {
	err := deactivateActiveEnquiry("ni", "")
	if err != nil {
		fmt.Println(err.Error())
		if cfg_details.INPUT_ERROR != err.Error() {
			t.Fail()
		}
	} else {
		t.Fail()
	}
}

func TestDeactivateEnquiryBestCase(t *testing.T) {
	err := deactivateActiveEnquiry("ni", "01JJ8R5ASPDQ25W0AVH0K23JAB")
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
	}
	fmt.Println("Deactivate Successful - ")
}

func TestGetActiveEnquiry(t *testing.T) {
	_, err := GetActiveEnquiry("ni", "01JJ9NTDC0BPBX22AMZVQT7750")
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
	}
}

func TestGetEnqV2(t *testing.T) {
	var pgOnce sync.Once
	pgOnce.Do(func() {
		ctx := context.Background()
		db, err := pgxpool.New(ctx, CONN_STR)
		if err != nil {
			err = fmt.Errorf("unable to create connection pool: %w", err)
			return
		}
		res, err := GetEnqV2(db, ctx, 1053340366668365825, "ni")
		db.Close()
		if err != nil {
			fmt.Println(err.Error())
			t.Fail()
		}
		fmt.Println(res)
	})
}

func TestUpdateEnqStatusOnly(t *testing.T) {
	var pgOnce sync.Once
	pgOnce.Do(func() {
		ctx := context.Background()
		db, err := pgxpool.New(ctx, CONN_STR)
		if err != nil {
			err = fmt.Errorf("unable to create connection pool: %w", err)
			return
		}
		err = UpdateEnqStatusOnly(db, ctx, "ni", "1066927174255902721", "Admitted", "Test-User")
		db.Close()
		if err != nil {
			fmt.Println(err.Error())
			t.Fail()
		}
	})
}

func TestAddEnqV2(t *testing.T) {

	current_time := time.Now()
	var pgOnce sync.Once
	pgOnce.Do(func() {
		ctx := context.Background()
		db, err := pgxpool.New(ctx, CONN_STR)
		if err != nil {
			err = fmt.Errorf("unable to create connection pool: %w", err)
			return
		}
		frmData := FormData{
			Name:           "Riaz Pasha",
			Course:         "PUC-R",
			Mobile:         "9800780908",
			Reference:      "NA",
			Location:       "BDA Banashakari",
			Status:         "Initiated",
			CouncillorId:   "323kjahe2398",
			CouncillorName: "Nikhil",
			UpdatedBy:      "323kjahe2398",
			Batch:          "25-26",
			Ts:             current_time.Format("2006-01-02 15:04:05"),
			Doe:            "2025-04-04",
		}
		_, err = AddEnqV2(db, ctx, frmData, "ni", "")
		db.Close()
		if err != nil {
			fmt.Println("Insertion failed - ")
			t.Fail()
		}
		return
	})
	fmt.Println("Successfully Inserted - ")
}

func TestAddCommentV2(t *testing.T) {
	var pgOnce sync.Once
	pgOnce.Do(func() {
		ctx := context.Background()
		db, err := pgxpool.New(ctx, CONN_STR)
		if err != nil {
			err = fmt.Errorf("unable to create connection pool: %w", err)
			return
		}
		comments := Comment{
			Comment: "Test Comment UT - 2",
			EqId:    "1053567950420541441",
		}
		err = AddCommentV2(db, ctx, comments, "ni")
		db.Close()
		if err != nil {
			fmt.Println("Insertion failed - ")
			t.Fail()
		}
		return
	})
}

func TestGetCommentV2(t *testing.T) {
	var pgOnce sync.Once
	pgOnce.Do(func() {
		ctx := context.Background()
		db, err := pgxpool.New(ctx, CONN_STR)
		if err != nil {
			err = fmt.Errorf("unable to create connection pool: %w", err)
			return
		}
		cmts, err := GetCommentV2(db, ctx, 1053567950420546441, "ni")
		db.Close()
		if err != nil {
			fmt.Println("Insertion failed - ")
			t.Fail()
			return
		}
		fmt.Println(cmts)
	})
}

func TestUpdateEnqV2(t *testing.T) {
	var pgOnce sync.Once
	pgOnce.Do(func() {
		ctx := context.Background()
		db, err := pgxpool.New(ctx, CONN_STR)
		if err != nil {
			err = fmt.Errorf("unable to create connection pool: %w", err)
			return
		}
		frmData := FormData{
			Id:             "1053258325340749825",
			Name:           "Md Riaz Pasha",
			Course:         "PUC-R",
			Mobile:         "9800781908",
			Reference:      "NA",
			Location:       "BDA Banashakari",
			Status:         "In Progress",
			CouncillorId:   "323kjahe2398",
			CouncillorName: "Nikhil",
			UpdatedBy:      "323kjahe2398",
			Batch:          "25-26",
		}
		err = UpdateEnqV2(db, ctx, frmData, "ni", "UT-user")
		db.Close()
		if err != nil {
			fmt.Println("Update failed - ")
			t.Fail()
		}
		return
	})

}

func TestListEnqV2(t *testing.T) {
	var pgOnce sync.Once
	pgOnce.Do(func() {
		ctx := context.Background()
		db, err := pgxpool.New(ctx, CONN_STR)
		if err != nil {
			err = fmt.Errorf("unable to create connection pool: %w", err)
			return
		}
		res, err := ListEnqV2(db, ctx, "ni")
		db.Close()
		if err != nil || len(res) == 0 {
			fmt.Println(err.Error())
			t.Fail()
		}

		fmt.Println(res)
	})
}

func TestDelEnqV2(t *testing.T) {
	var pgOnce sync.Once
	pgOnce.Do(func() {
		ctx := context.Background()
		db, err := pgxpool.New(ctx, CONN_STR)
		if err != nil {
			err = fmt.Errorf("unable to create connection pool: %w", err)
			return
		}
		err = DelEnqV2(db, ctx, 1053355118851391489, "ni")
		db.Close()
		if err != nil {
			fmt.Println(err.Error())
			t.Fail()
		}
		fmt.Println("Deleted")
	})
}

func printMess() {
	fmt.Println("Defer during the panic in method example")
}

func TestToken(t *testing.T) {
	token := "MTc0MjExNjc4NS5hZG1pbixjb3Vuc2VsbG9yLGZhY3VsdHkuYXV0aDB8NjcyNWUzZDAwOTI0M2RiNWFjMTQ0Y2RjLjUzZDg2MTZkMmM4Zjc2N2I2OGUzZDhmYzUyYmMxZWI3ZTJkYjYxZTc1ODY2OGQ4NDEzMzFjYmY5YmIzYjNhOGY="
	decodeByte, err := base64.URLEncoding.DecodeString(token)
	decodeString := string(decodeByte[:])
	if err != nil {
		t.Fail()
	}
	tknParts := strings.Split(decodeString, ".")
	currentTs := time.Now().Unix()
	tknTs, _ := strconv.ParseInt(tknParts[0], 10, 64)
	if tknTs <= currentTs {
		t.Fail()
	}
	err = verifySignature(tknParts)
	if err != nil {
		t.Fail()
	}
	if !strings.Contains(tknParts[1], "counsellor") {
		t.Fail()
	}
}

func verifySignature(tknParts []string) error {
	tsMessage := tknParts[0] + "." + tknParts[1] + "." + tknParts[2]
	uEnc := base64.URLEncoding.EncodeToString([]byte(tsMessage))
	calculatedSignature := getSignature(uEnc)
	if tknParts[3] != calculatedSignature {
		return errors.New("Signature not matched")
	}
	return nil
}

var SECRET = []byte("This is the way -- Mando")

func getSignature(encMessage string) string {
	mac := hmac.New(sha256.New, SECRET)
	_, err := mac.Write([]byte(encMessage))
	if err != nil {
		return ""
	}
	return hex.EncodeToString(mac.Sum(nil))
}

func TestJsonParse(t *testing.T) {
	str := `
		{"id":null,"councillor_name":"","course":"","email":"sach@gmail.com","mobile":"9876512345","name":"Sachin","location":"BSK 2nd Stage","status":"Initiated"}
	`
	var frmData FormData
	err := json.Unmarshal([]byte(str), &frmData)
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
	fmt.Println(frmData)
}
