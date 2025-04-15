package enquiry

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Comment struct {
	Id      string `json:"id"`
	Comment string `json:"comment"`
	EqId    string `json:"eq_id"`
	Ts      string `json:"ts"`
}

type FormData struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	Course         string `json:"course"`
	Mobile         string `json:"mobile"`
	ParentMobile   string `json:"parent_mobile"`
	Reference      string `json:"reference"`
	Location       string `json:"location"`
	Status         string `json:"status"`
	CouncillorId   string `json:"councillor_id"`
	CouncillorName string `json:"councillor_name"`
	UpdatedBy      string `json:"u_by"`
	Batch          string `json:"batch"`
	PrevQual       string `json:"prev_qual"`
	Ts             string `json:"ts"`
	Doe            string `json:"doe"`
}

func ListEnqV2(con *pgxpool.Pool, ctx context.Context, colId string) ([]FormData, error) {
	tableName := getEnqTableName(colId)
	rows, err := con.Query(ctx, "select id, name, course, mobile, parent_mobile, status, councillor_name, location, doe from "+tableName+" order by id desc limit 500")
	if err != nil {
		fmt.Println("Error querying the Eq table for fetching - " + err.Error())
	}
	res := make([]FormData, 0)
	for rows.Next() {
		var pR FormData
		var ts time.Time
		var doe time.Time
		err := rows.Scan(&pR.Id, &pR.Name, &pR.Course, &pR.Mobile, &pR.ParentMobile, &pR.Status, &pR.CouncillorName, &pR.Location, &doe)
		pR.Ts = ts.Format("2006-01-02 15:04:05")
		pR.Doe = doe.Format("2006-01-02")
		if err != nil {
			fmt.Println("Error querying the Eq table for fetching - " + err.Error())
			return nil, err
		}
		res = append(res, pR)
	}
	return res, err
}

func getEnqTableName(colId string) string {
	return "students.enquiry_" + colId
}

func getCommentsTableName(colId string) string {
	return "students.comments_" + colId
}

func GetEnqV2(con *pgxpool.Pool, ctx context.Context, id int64, colId string) (FormData, error) {
	tableName := getEnqTableName(colId)
	rows, err := con.Query(ctx, "select * from "+tableName+" where id = $1", id)
	if err != nil {
		fmt.Println("Error querying the Eq table for fetching - " + err.Error())
	}
	var pR FormData
	for rows.Next() {
		var ts time.Time
		err := rows.Scan(&pR.Id, &pR.Name, &pR.Course, &pR.Location, &pR.CouncillorId, &pR.CouncillorName, &pR.Mobile, &pR.Reference,
			&pR.Status, &pR.UpdatedBy, &ts, &pR.Batch, &pR.PrevQual, &pR.ParentMobile, &pR.Doe)
		pR.Ts = ts.Format("2006-01-02 15:04:05")
		if err != nil {
			fmt.Println("Error querying the Eq table for fetching - " + err.Error())
			continue
		}
	}
	return pR, err
}

func DelEnqV2(con *pgxpool.Pool, ctx context.Context, id int64, colId string) error {
	row, err := con.Query(ctx, "Delete from "+getEnqTableName(colId)+" where id = "+strconv.FormatInt(id, 10))
	if err != nil {
		log.Println("Error deleting the Eq - " + err.Error())
		return err
	}
	defer row.Close()
	if row.Next() {
		err := row.Scan()
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			fmt.Println("Error deleting the Eq table - " + err.Error())
		}
	}
	return err
}

func AddCommentV2(con *pgxpool.Pool, ctx context.Context, comments Comment, colId string) error {
	commentsTable := getCommentsTableName(colId)
	query := `Insert INTO ` + commentsTable + ` ("comments", "eq_id") values (@comment, @eq_id) RETURNING id`
	args := pgx.NamedArgs{
		"comment": comments.Comment,
		"eq_id":   comments.EqId,
	}
	row, err := con.Query(ctx, query, args)
	if err != nil {
		log.Printf("Error while creating connection %v\n", err)
		return err
	}
	defer row.Close()
	if err != nil {
		fmt.Println("Error querying the Comments table for inserting - Query - " + err.Error())
	}
	var id int
	if row.Next() {
		err = row.Scan(&id)
	}
	if err != nil {
		fmt.Println("Error querying the Comments table for inserting - Scan - " + err.Error())
	}
	return err
}

func GetCommentV2(con *pgxpool.Pool, ctx context.Context, id int64, colId string) ([]Comment, error) {
	commentsTable := getCommentsTableName(colId)
	query := `Select id, comments, ts, eq_id from ` + commentsTable + ` where eq_id = ` + strconv.FormatInt(id, 10)
	row, err := con.Query(ctx, query)
	comments := make([]Comment, 0)
	defer row.Close()
	for row.Next() {
		var cmt Comment
		var ts time.Time
		err = row.Scan(&cmt.Id, &cmt.Comment, &ts, &cmt.EqId)
		if err != nil {
			fmt.Println("Error querying the Comments table for fetching - Scan - " + err.Error())
		}
		cmt.Ts = ts.Format("2006-01-02 15:04:05")
		comments = append(comments, cmt)
	}
	return comments, err

}

func UpdateEnqV2(con *pgxpool.Pool, ctx context.Context, formData FormData, colId string, uBy string) error {
	tableName := getEnqTableName(colId)
	query := `Update ` + tableName + ` set name=$1, mobile=$2, parent_mobile=$3, status=$4, prev_qual=$7, u_by=$5, location=$8, 
					reference=$9, doe=$10 WHERE id = $6`
	_, err := con.Exec(ctx, query, formData.Name, formData.Mobile, formData.ParentMobile, formData.Status, uBy, formData.Id,
		formData.PrevQual, formData.Location, formData.Reference, formData.Doe)
	if err != nil {
		fmt.Println("Error deleting the Eq table - " + err.Error())
	}
	return err
}

func AddEnqV2(con *pgxpool.Pool, ctx context.Context, eqForm FormData, colId string, uBy string) (string, error) {
	tableName := getEnqTableName(colId)
	query := `INSERT INTO ` + tableName + ` ("name", "course", "location", "councillor_id", "councillor_name", "mobile", "parent_mobile", 
				"reference", "status", "u_by", "batch", "prev_qual", "doe")
				values (@name, @course, @location, @councillor_id, @councillor_name, @mobile, @parent_mobile, @reference, @status, 
				@u_by, @batch, @prev_qual, @doe)
				RETURNING id`
	args := pgx.NamedArgs{
		"name":            eqForm.Name,
		"course":          eqForm.Course,
		"location":        eqForm.Location,
		"councillor_id":   uBy,
		"councillor_name": eqForm.CouncillorName,
		"mobile":          eqForm.Mobile,
		"parent_mobile":   eqForm.ParentMobile,
		"reference":       eqForm.Reference,
		"status":          eqForm.Status,
		"u_by":            uBy,
		"batch":           eqForm.Batch,
		"prev_qual":       eqForm.PrevQual,
		"doe":             eqForm.Doe,
	}
	rows, err := con.Query(context.TODO(), query, args)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	defer rows.Close()
	var id int
	if rows.Next() {
		if err := rows.Scan(&id); err != nil {
			log.Printf("Error in rows Scan - %v\n", err)
			return "", err
		}
	}
	if err != nil {
		log.Println(err.Error())
	}
	return strconv.Itoa(id), err
}
