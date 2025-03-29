package attendance

import (
	"cerpApi/cfg_details"
	"cerpApi/notifications"
	"cerpApi/students"
	"cerpApi/subject"
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log"
	"math"
	"sort"
	"strings"
)

type Student struct {
	Name         string `json:"name"`
	Id           string `json:"id"`
	IsPresent    bool   `json:"isPresent"`
	NotifyMobile string `json:"notify_mobile"`
}

type AttendanceFormUpdater struct {
	PK       string `dynamodbav:"sub"`
	SK       string `dynamodbav:"date"`
	Values   string `dynamodbav:"values"`
	TimeSlot string `dynamodbav:"time_slot"`
	Ts       int64  `dynamodbav:"ts"`
	Updater  string `dynamodbav:"uBy"`
}

type AttendanceForm struct {
	Students []string `json:"students"`
	TimeSlot string   `json:"time_slot"`
	Ts       int64    `json:"ts"`
	UBy      string   `json:"u_by"`
	WorkLog  string   `json:"work_log"`
}

var isTest = false

func GetAttendanceForm(college string, course string, batch string, sub string, date string, cs string) (AttendanceForm, error) {
	key, err := getAttendanceKey(course, batch, sub, cs)
	sKey, err := attributevalue.Marshal(date)
	if err != nil {
		return AttendanceForm{}, err
	}
	ck := map[string]types.AttributeValue{
		"sub":  key,
		"date": sKey,
	}
	out, err := cfg_details.DynamoCfg.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String("attendance" + "_" + strings.ToLower(college)),
		Key:       ck,
	})
	item := out.Item["values"]
	if item == nil {
		return AttendanceForm{}, nil
	}
	var res1 string
	err = attributevalue.Unmarshal(item, &res1)
	var res AttendanceForm
	err = json.Unmarshal([]byte(res1), &res)
	log.Printf("The response is %s %s\n", res.TimeSlot, res.WorkLog)
	return res, err
}

func getAttendanceKey(course string, batch string, sub string, cs string) (types.AttributeValue, error) {
	return attributevalue.Marshal(getAttendanceKeyStr(course, batch, sub, cs))
}

func getAttendanceKeyStr(course string, batch string, sub string, cs string) string {
	return strings.ToLower(course) + "_" + strings.ToLower(batch) + "_" + strings.ToLower(sub) + "_" + strings.ToLower(cs)
}

func UpdateAttendance(colId string, course string, batch string, sub string, date string, cs string, sn bool, req *AttendanceForm, studentsMasterSet []Student) error {
	err := updateAttendanceForm(colId, course, batch, sub, date, cs, req)
	if err != nil {
		return err
	}
	if sn {
		go processAbsenteesSmsNotifications(colId, course, batch, sub, date, req, studentsMasterSet)
	}
	return err
}

func processAbsenteesSmsNotifications(colId string, course string, batch string, sub string, date string, attendanceDetails *AttendanceForm, fullStudentsSet []Student) {
	log.Printf("In Absent notification System...")
	subjects := subject.GetSubjects(colId, course, batch)
	subName, exists := subjects[sub]
	if !exists {
		log.Printf("No matcing subject found, check the data onboarding %s %s %s %v\n", colId, course, batch, subjects)
	}
	presentMap := make(map[string]bool)
	presentStudents := attendanceDetails.Students
	for _, stud := range presentStudents {
		presentMap[stud] = true
	}
	absentees := make([]interface{}, 0)
	for _, stud := range fullStudentsSet {
		if !presentMap[stud.Id] && (stud.NotifyMobile != "" || isTest) {
			if isTest {
				stud.NotifyMobile = "9743213012"
			}
			notifyData := map[string]string{
				"mobile": stud.NotifyMobile,
				"id":     stud.Id,
				"name":   stud.Name,
			}
			absentees = append(absentees, notifyData)
		}
	}
	notifyWrapper := notifications.AbsentNotificationWrapper{
		NotifyChannel: "whatsapp",
		Data:          absentees,
		Timeslot:      attendanceDetails.TimeSlot,
		Date:          date,
		Subject:       subName,
	}
	notifications.NotifyUsers(notifyWrapper, "Q")
}

func updateAttendanceForm(colId string, course string, batch string, sub string, date string, cs string, req *AttendanceForm) error {
	key := strings.ToLower(course) + "_" + strings.ToLower(batch) + "_" + strings.ToLower(sub) + "_" + strings.ToLower(cs)
	parsedReq, err := json.Marshal(req)
	if err != nil {
		return err
	}
	attendanceBook := AttendanceFormUpdater{
		PK:       key,
		SK:       date,
		Values:   string(parsedReq),
		TimeSlot: req.TimeSlot,
		Ts:       req.Ts,
		Updater:  req.UBy,
	}
	data, err := attributevalue.MarshalMap(attendanceBook)
	_, err = cfg_details.DynamoCfg.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("attendance" + "_" + colId),
		Item:      data,
	})

	if err != nil {
		return err
	}
	return nil
}

type Item struct {
	Sub    string `dynamodbav:"sub"`
	Date   string `dynamodbav:"date"`
	Values string `dynamodbav:"values"`
	Ts     int64  `dynamodbav:"ts"`
}

func batchGetItems(colId string, key string) ([]Item, error) {
	tableName := "attendance_" + colId

	input := &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String("#s = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: key},
		},
		ProjectionExpression: aws.String("#s, #val, #date, ts"), // Use alias instead of "sub"
		ExpressionAttributeNames: map[string]string{
			"#s":    "sub", // Alias "sub" since it's a reserved word
			"#val":  "values",
			"#date": "date",
		},
	}

	result, err := cfg_details.DynamoCfg.Query(context.TODO(), input)
	if err != nil {
		return nil, err
	}

	var items []Item
	err = attributevalue.UnmarshalListOfMaps(result.Items, &items)
	return items, err
}

func GetAttendanceReport(college string, course string, batch string, sub string, cs string) string {
	studData, err := students.GetStudents(college, course, batch, cs)
	if err != nil {
		log.Printf("Error in GetAttendance Report while fetching the students master data - %v\n", err)
		return ""
	}
	var res string
	err = attributevalue.Unmarshal(studData, &res)
	var parsedRes []Student
	err = json.Unmarshal([]byte(res), &parsedRes)
	studMap := make(map[string]Student)
	for _, val := range parsedRes {
		studMap[val.Id] = val
	}
	keyStr := getAttendanceKeyStr(course, batch, sub, cs)
	items, err := batchGetItems(college, keyStr)
	if err != nil {
		log.Printf("Error Fetching the batchGetItems %v\n", err)
	}
	dates := make([]string, 0)
	datesToAttendanceMap := make(map[string]Item)
	for _, item := range items {
		dates = append(dates, item.Date)
		datesToAttendanceMap[item.Date] = item
	}
	sort.Slice(dates, func(i, j int) bool {
		return dates[i] < dates[j]
	})
	attendanceRecords := generateReport(dates, studMap, datesToAttendanceMap)
	if attendanceRecords != nil {
		res, err := json.Marshal(AttendancePerClass{
			AttendanceReports: attendanceRecords,
			Key:               keyStr,
		})
		if err != nil {
			return ""
		}
		return string(res)
	}
	return ""
}

type StudentAttendanceRecord struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Date      string `json:"date"`
	IsPresent bool   `json:"is_present"`
}

type StudentAttendanceReport struct {
	AttendanceRecords StudentAttendanceRecord `json:"attendance_records"`
	TotalDays         int16                   `json:"total_days"`
	PresentDays       int16                   `json:"present_days"`
	Percentage        float32                 `json:"percentage"`
}

type AttendanceReport struct {
	Date     string                    `json:"date"`
	TimeSlot string                    `json:"time_slot"`
	WorkLog  string                    `json:"work_log"`
	Students []StudentAttendanceReport `json:"students"`
}

type AttendancePerClass struct {
	AttendanceReports []AttendanceReport `json:"attendance_reports"`
	Key               string             `json:"key"`
}

func generateReport(dates []string, studMap map[string]Student, dateVsItemsMap map[string]Item) []AttendanceReport {
	res := make([]AttendanceReport, 0)
	studentsAttendanceRecords := make(map[string][]StudentAttendanceReport)
	for ind, date := range dates {
		attendanceReport := AttendanceReport{
			Date:     date,
			TimeSlot: "",
			WorkLog:  "",
			Students: nil,
		}
		studentsAttendees := make(map[string]bool)
		val, exist := dateVsItemsMap[date]
		if exist {
			var studentRecord AttendanceForm
			err := json.Unmarshal([]byte(val.Values), &studentRecord)
			if err != nil {
				log.Printf("Error in unmarshalling student attendance report - %v\n", err)
				return nil
			}
			for _, studId := range studentRecord.Students {
				studentsAttendees[studId] = true
			}
			attendanceReport.TimeSlot = studentRecord.TimeSlot
			attendanceReport.WorkLog = studentRecord.WorkLog
		}
		attendanceReport.Students = make([]StudentAttendanceReport, 0)
		for id, stud := range studMap {
			_, isPresent := studentsAttendees[id]
			studData := StudentAttendanceRecord{
				Id:        id,
				Date:      date,
				IsPresent: isPresent,
				Name:      stud.Name,
			}
			studReport := StudentAttendanceReport{
				AttendanceRecords: studData,
				TotalDays:         int16(ind + 1),
				PresentDays:       0,
				Percentage:        0,
			}
			lst, is2ndDay := studentsAttendanceRecords[id]
			studReport.AttendanceRecords = studData
			if is2ndDay {
				if studData.IsPresent {
					studReport.PresentDays = lst[len(lst)-1].PresentDays + 1
				} else {
					studReport.PresentDays = lst[len(lst)-1].PresentDays
				}
			} else {
				if studData.IsPresent {
					studReport.PresentDays = 1
				}
				lst = []StudentAttendanceReport{}
			}
			studReport.Percentage = calculatePercentage(studReport.PresentDays, studReport.TotalDays)
			//float32(math.Round(float64(studReport.PresentDays)/float64(studReport.TotalDays))) * 100
			studentsAttendanceRecords[id] = append(lst, studReport)
			attendanceReport.Students = append(attendanceReport.Students, studReport)
		}
		res = append(res, attendanceReport)
	}
	return res
}

func calculatePercentage(presentDays int16, totalDays int16) float32 {
	num := float64(presentDays) / float64(totalDays)
	num2 := float32(math.Round(num*10000) / 100)
	return num2
}
