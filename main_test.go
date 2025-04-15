package main

import (
	"fmt"
	"testing"
)

func TestGetAttendanceStudents(t *testing.T) {
	students, err := getAttendanceStudents("ni", "PUC-R", "2024-2026_2ND-YR", "CS-2025-PUC-R", "2025-3-28", "A")
	if err != nil {
		return
	}
	fmt.Println(students)
}

func TestGetAttendanceStudents2(t *testing.T) {
	students, err := getAttendanceStudents("ni", "BBA", "2023-2026_SEM-4", "AI-24BBA", "2025-3-28", "A")
	if err != nil {
		return
	}
	fmt.Println(students)
}

func TestGetAttendanceStudents3(t *testing.T) {
	// college_id=ni&class=BBA&batch=2024-2027_SEM-2&subject=ET-25BBA&date=9-4-2025&class_section=A
	students, err := getAttendanceStudents("ni", "BBA", "2024-2027_SEM-2", "ET-25BBA", "9-4-2025", "A")
	if err != nil {
		return
	}
	fmt.Println(students)
}
