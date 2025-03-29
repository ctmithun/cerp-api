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
