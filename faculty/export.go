package faculty

import (
	"bytes"
	"cerpApi/cfg_details"
	"context"
	"encoding/json"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/disintegration/imaging"
	"github.com/xuri/excelize/v2"
)

var HEADERS = []string{"Emp No", "Name", "Department", "Designation", "Blood Group", "Date Of Joining", "Bank Name", "Bank Account Number", "IFSC", "Photo"}

func ExportForIdGeneration(s3Client *s3.Client, colId string, fIds []string) (string, error) {

	var faculties []Faculty
	for _, fId := range fIds {
		facultyData := getFacultiesData(colId, fId)
		if len(facultyData) > 0 {
			faculties = append(faculties, facultyData[0])
		}
	}

	// Create Excel file
	f := excelize.NewFile()
	sheet := "Sheet1"
	f.SetSheetName("Sheet1", "Faculty")
	sheet = "Faculty"

	for i, h := range HEADERS {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetCellValue(sheet, fmt.Sprintf("%s1", col), h)
	}
	rowCounter := 2
	for _, fac := range faculties {
		if fac.EmpNo == "" {
			continue
		}
		row := rowCounter
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), fac.EmpNo)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), fac.Name)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), fac.Department)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), fac.Designation)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), fac.BloodGroup)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), cfg_details.ParseDateStr(fac.Doj))
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), fac.BankDetails.BankName)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", row), fac.BankDetails.AccountNumber)
		f.SetCellValue(sheet, fmt.Sprintf("I%d", row), fac.BankDetails.IFSC)

		// Download photo from S3.
		if fac.Photo != "" {

			log.Println("Processing the " + fac.Photo)
			rowHeight := getProfilePicsFromS3(s3Client, colId, fac.Photo)
			log.Println("RowHeight is ", rowHeight)
			cell := fmt.Sprintf("J%d", row)
			enable, disable := true, false
			f.SetColWidth(sheet, "A", "J", 20) // 100px / 7.5
			f.SetRowHeight(sheet, row, rowHeight/1.5)
			imagePath := "/tmp/" + fac.Photo
			if err := f.AddPicture(sheet, cell, imagePath, &excelize.GraphicOptions{
				PrintObject:     &enable,
				LockAspectRatio: true,
				OffsetX:         15,
				OffsetY:         10,
				Locked:          &disable,
				ScaleX:          0.5,
				ScaleY:          0.5,
			}); err != nil {
				log.Printf("Failed to add image: %v\n", err)
			}

			// Clean up
			os.Remove(imagePath)
		} else {
			f.SetCellValue(sheet, fmt.Sprintf("J%d", row), "Image Unavailable")
		}
		rowCounter = rowCounter + 1
	}
	curTs := cfg_details.GetCurrentTsStr()
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		log.Fatalf("failed to write Excel file to buffer: %v", err)
	}
	key := colId + "/faculty/export/faculty_export_" + curTs
	err := uploadExcelToS3(s3Client, key, &buf)
	if err != nil {
		log.Printf("Error uploading excel to s3 %s\n", key)
		return "", nil
	}
	presignLink, err := preSignFile(key)
	if err != nil {
		return "", nil
	}
	return presignLink, nil
}

func preSignFile(key string) (string, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(cfg_details.BUCKET_STUDENTS_FACULTIES),
		Key:    aws.String(key),
	}
	ctx, cancel := cfg_details.GetTimeoutContext()
	defer cancel()
	presignedURL, err := cfg_details.Presigner.PresignGetObject(ctx, input, s3.WithPresignExpires(1*time.Minute))
	if err != nil {
		log.Printf("Error in presogning the requested document - %s error-%v\n", key, err)
		return "", err
	}
	enc := url.QueryEscape(presignedURL.URL)
	body, _ := json.Marshal(cfg_details.FileResponse{URL: enc})
	return string(body), nil
}

func uploadExcelToS3(client *s3.Client, keyName string, fileBuffer *bytes.Buffer) error {
	_, err := client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:        aws.String(cfg_details.BUCKET_STUDENTS_FACULTIES),
		Key:           aws.String(keyName), // e.g. "exports/faculty_export_2025.xlsx"
		Body:          bytes.NewReader(fileBuffer.Bytes()),
		ContentLength: aws.Int64(int64(fileBuffer.Len())),
		Tagging:       aws.String(cfg_details.ExpireTags()),
		ContentType:   aws.String("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"),
	})
	return err
}

func getProfilePicsFromS3(s3Client *s3.Client, colId, key string) float64 {
	ctx, cancel := cfg_details.GetTimeoutContext()
	defer cancel()
	getObj, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(cfg_details.BUCKET_STUDENTS_FACULTIES),
		Key:    aws.String(colId + "/faculty/profilepics/" + key),
	})
	if err != nil {
		log.Printf("Error getting photo for %s: %v", key, err)
		return 0
	}
	defer getObj.Body.Close()
	localPath := filepath.Join("/tmp", filepath.Base(key))
	outFile, _ := os.Create(localPath)
	_, _ = outFile.ReadFrom(getObj.Body)
	outFile.Close()
	return resize(localPath)
}

func resize(key string) float64 {
	src, err := imaging.Open(key)
	if err != nil {
		log.Printf("failed to open image: %v", err)
	}

	// Resize to width 200px, height auto (0 keeps aspect ratio)
	dst := imaging.Resize(src, 200, 0, imaging.Lanczos)

	// Save the resulting image
	err = imaging.Save(dst, key)
	if err != nil {
		log.Printf("failed to save image: %v", err)
	}
	imageHeight := dst.Bounds().Dy()
	rowHeight := float64(imageHeight) / 1.33
	return rowHeight
}
