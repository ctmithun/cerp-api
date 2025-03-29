package students

import (
	"fmt"
	"testing"
)

func TestGetStudentsData(t *testing.T) {
	student := GetStudentsData("ni", "2025", "PUC-R")
	fmt.Println(student)
}
