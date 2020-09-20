package handler

import (
	"database/sql"
	"errors"
	"os"

	"github.com/choobot/choo-pos-backend/app/model"

	_ "github.com/go-sql-driver/mysql"
)

type UserHandler interface {
	CreateLog(id string) error
	GetAllLog() ([]model.UserLog, error)
}

type MySqlUserHandler struct {
	db *sql.DB
}

func NewMySqlUserHandler() MySqlUserHandler {
	db, _ := sql.Open("mysql", os.Getenv("DATA_SOURCE_NAME"))
	return MySqlUserHandler{
		db: db,
	}
}

func (this *MySqlUserHandler) SetTimeZone() error {
	sql := `SET time_zone = 'Asia/Bangkok'`
	_, err := this.db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

func (this *MySqlUserHandler) CreateTablesIfNotExist() error {
	sql := "SELECT 1 FROM user_log LIMIT 1"
	_, err := this.db.Query(sql)
	if err != nil {
		sql = `
		CREATE TABLE user_log (
			id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		) CHARACTER SET utf8 COLLATE utf8_general_ci`

		if _, err = this.db.Exec(sql); err != nil {
			return err
		}
	}

	return nil
}

func (this *MySqlUserHandler) CreateLog(id string) error {
	this.SetTimeZone()
	if err := this.CreateTablesIfNotExist(); err != nil {
		return err
	}
	sql := `INSERT INTO user_log (user_id) VALUES(?)`
	result, err := this.db.Exec(sql, id)
	if err != nil {
		return err
	}
	num, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if num != 1 {
		return errors.New("No record")
	}

	return nil
}

func (this *MySqlUserHandler) GetAllLog() ([]model.UserLog, error) {
	this.SetTimeZone()
	if err := this.CreateTablesIfNotExist(); err != nil {
		return nil, err
	}
	var userLogs []model.UserLog
	rows, err := this.db.Query("SELECT CONVERT_TZ(created_at,'GMT','Asia/Bangkok'), user_id FROM user_log ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userId string
		var createdAt string
		if err := rows.Scan(&createdAt, &userId); err != nil {
			return nil, err
		}
		userLog := model.UserLog{
			Time: createdAt,
			User: model.User{
				Id: userId,
			},
		}
		userLogs = append(userLogs, userLog)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return userLogs, nil
}
