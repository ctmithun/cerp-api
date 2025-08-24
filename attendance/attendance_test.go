package attendance

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"testing"
)

func TestAbsentNotification(t *testing.T) {
	attendanceDetails := AttendanceForm{
		Students: []string{"BCA1"},
		TimeSlot: "9.30 - 10.30 AM",
		Ts:       0,
		UBy:      "Test",
		WorkLog:  "Test",
	}
	fullStudentDataSet := []Student{
		{
			Name:         "BCA1",
			Id:           "BCA1",
			IsPresent:    true,
			NotifyMobile: "9743213012",
		},
		{
			Name:         "BCA2",
			Id:           "BCA2",
			IsPresent:    false,
			NotifyMobile: "9743186443",
		},
	}
	processAbsenteesSmsNotifications("ni", "bca", "2024-2027", "Data Structure", "2025-03-20", &attendanceDetails, fullStudentDataSet)
}

func TestAttendanceReport1(t *testing.T) {
	items, err := batchGetItems("ni", "bca_2024-2027_sem-2_ds-24bca_a")
	if err != nil {
		t.Fail()
		log.Printf("Test case failed for %v\n", err)
		return
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Date < items[j].Date
	})
	fmt.Printf("The output is %v\n", items)
}

func TestBatchGetItems(t *testing.T) {
	items, err := batchGetItems("ni", "bca_2024-2027_sem-2_ds-lab-24bca_a")
	if err != nil {
		t.Fail()
		log.Printf("Test case failed for %v\n", err)
		return
	}
	fmt.Printf("The output is %v\n", items)
}

func TestAttendanceReport2(t *testing.T) {
	res := GetAttendanceReport("ni", "BCA", "2024-2027_SEM-2", "DS-LAB-24BCA", "A", "", "")
	//num := float64(1) / float64(3)
	//num2 := float32(math.Round(num*10000) / 100)
	//res1 := num2
	//fmt.Println(res1)
	if res == "" {
		log.Printf("No records found")
	}
	jsonRes, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		t.Fail()
		log.Printf("Test case failed for %v\n", err)
	}
	log.Printf("The output is %v\n", string(jsonRes))
}

func TestGetPmn(t *testing.T) {
	item := getPmn("ni", "U18FE24S0001")
	fmt.Println(item)
}

func TestGetPmnNaN(t *testing.T) {
	item := getPmn("ni", "U18FR22C0101")
	fmt.Println(item)
}

func TestStringMob(t *testing.T) {
	mob := strings.Split("8901236541.0", ".")[0]
	fmt.Print(mob)
}

func TestGetAttendanceForm(t *testing.T) {
	res, _ := GetAttendanceForm("ni", "BBA", "2024-2027_SEM-2", "ET-25BBA", "8-4-2025", "A")
	log.Printf("Res - %v\n", res)
}
