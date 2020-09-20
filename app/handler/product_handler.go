package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strings"

	"github.com/choobot/choo-pos-backend/app/model"

	_ "github.com/go-sql-driver/mysql"
)

type ProductHandler interface {
	Create(product model.Product) error
	GetAll() ([]model.Product, error)
	GetByIds(ids []interface{}) (map[string]model.Product, error)
}

type MySqlProductHandler struct {
	db *sql.DB
}

func NewMySqlProductHandler() MySqlProductHandler {
	db, _ := sql.Open("mysql", os.Getenv("DATA_SOURCE_NAME"))
	return MySqlProductHandler{
		db: db,
	}
}

func (this *MySqlProductHandler) SetTimeZone() error {
	sql := `SET time_zone = 'Asia/Bangkok'`
	_, err := this.db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

func (this *MySqlProductHandler) CreateTablesIfNotExist() error {
	sql := "SELECT 1 FROM product LIMIT 1"
	rows, err := this.db.Query(sql)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		sql = `
		CREATE TABLE product (
			id VARCHAR(255) NOT NULL PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			price FLOAT NOT NULL,
			cover VARCHAR(255) NOT NULL,
			status INT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL
		) CHARACTER SET utf8 COLLATE utf8_general_ci`

		if _, err = this.db.Exec(sql); err != nil {
			return err
		}

		//Load init data
		file, err := ioutil.ReadFile("data/books.json")
		if err != nil {
			return err
		}
		var books model.Books
		err = json.Unmarshal(file, &books)
		if err != nil {
			return err
		}
		for _, book := range books.Books {
			if err := this.Create(book); err != nil {
				return err
			}
		}

	}

	return nil
}

func (this *MySqlProductHandler) Create(product model.Product) error {
	this.SetTimeZone()

	if err := this.CreateTablesIfNotExist(); err != nil {
		return err
	}
	sql := `INSERT INTO product (id, title, price, cover, status, updated_at) VALUES(?, ?, ?, ?, 1, CURRENT_TIMESTAMP)`
	result, err := this.db.Exec(sql, product.Id, product.Title, product.Price, product.Cover)
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

func (this *MySqlProductHandler) GetAll() ([]model.Product, error) {
	this.SetTimeZone()
	if err := this.CreateTablesIfNotExist(); err != nil {
		return nil, err
	}
	var products []model.Product
	rows, err := this.db.Query("SELECT id, title, price, cover, status, CONVERT_TZ(created_at,'GMT','Asia/Bangkok'), CONVERT_TZ(updated_at,'GMT','Asia/Bangkok') FROM product ORDER BY title")
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id string
		var title string
		var price float64
		var cover string
		var status int
		var createdAt string
		var updatedAt string
		if err := rows.Scan(&id, &title, &price, &cover, &status, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		product := model.Product{
			Id:        id,
			Title:     title,
			Price:     price,
			Cover:     cover,
			Status:    status,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
		products = append(products, product)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func (this *MySqlProductHandler) GetByIds(ids []interface{}) (map[string]model.Product, error) {
	this.SetTimeZone()
	if err := this.CreateTablesIfNotExist(); err != nil {
		return nil, err
	}
	productsMap := map[string]model.Product{}
	sql := "SELECT id, title, price, cover, status, CONVERT_TZ(created_at,'GMT','Asia/Bangkok'), CONVERT_TZ(updated_at,'GMT','Asia/Bangkok') FROM product WHERE id IN (?" + strings.Repeat(",?", len(ids)-1) + ")"

	rows, err := this.db.Query(sql, ids...)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id string
		var title string
		var price float64
		var cover string
		var status int
		var createdAt string
		var updatedAt string
		if err := rows.Scan(&id, &title, &price, &cover, &status, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		product := model.Product{
			Id:        id,
			Title:     title,
			Price:     price,
			Cover:     cover,
			Status:    status,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		}
		productsMap[id] = product
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return productsMap, nil
}
