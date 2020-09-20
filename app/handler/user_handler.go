package handler

import (
	"database/sql"
	"errors"
	"os"

	"github.com/choobot/choo-pos-backend/app/model"

	_ "github.com/go-sql-driver/mysql"
)

type UserHandler interface {
	CreateLog(user model.User) error
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
	rows, err := this.db.Query(sql)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		sql = `
		CREATE TABLE user_log (
			id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			picture VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		) CHARACTER SET utf8 COLLATE utf8_general_ci`

		if _, err = this.db.Exec(sql); err != nil {
			return err
		}
	}

	return nil
}

func (this *MySqlUserHandler) CreateLog(user model.User) error {
	this.SetTimeZone()
	if err := this.CreateTablesIfNotExist(); err != nil {
		return err
	}
	sql := `INSERT INTO user_log (user_id, name, picture) VALUES(?, ?, ?)`
	result, err := this.db.Exec(sql, user.Id, user.Name, user.Picture)
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
	rows, err := this.db.Query("SELECT CONVERT_TZ(created_at,'GMT','Asia/Bangkok'), user_id, name, picture FROM user_log ORDER BY created_at DESC")
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var userId string
		var name string
		var picture string
		var createdAt string
		if err := rows.Scan(&createdAt, &userId, &name, &picture); err != nil {
			return nil, err
		}
		userLog := model.UserLog{
			Time: createdAt,
			User: model.User{
				Id:      userId,
				Name:    name,
				Picture: picture,
			},
		}
		userLogs = append(userLogs, userLog)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return userLogs, nil
}
