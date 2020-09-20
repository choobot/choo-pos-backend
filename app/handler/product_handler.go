package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/choobot/choo-pos-backend/app/model"

	_ "github.com/go-sql-driver/mysql"
)

type ProductHandler interface {
	Create(product model.Product) error
	GetAll() ([]model.Product, error)
}

type ProductMySqlHandler struct {
	db *sql.DB
}

func NewProductMySqlHandler() ProductMySqlHandler {
	db, _ := sql.Open("mysql", os.Getenv("DATA_SOURCE_NAME"))
	return ProductMySqlHandler{
		db: db,
	}
}

func (this *ProductMySqlHandler) SetTimeZone() error {
	sql := `SET time_zone = 'Asia/Bangkok'`
	_, err := this.db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

func (this *ProductMySqlHandler) CreateTablesIfNotExist() error {
	sql := "SELECT 1 FROM product LIMIT 1"
	_, err := this.db.Query(sql)
	if err != nil {
		sql = `
		CREATE TABLE product (
			id VARCHAR(255) NOT NULL PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			price FLOAT NOT NULL,
			cover VARCHAR(255) NOT NULL,
			status INT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) CHARACTER SET utf8 COLLATE utf8_general_ci`

		if _, err = this.db.Exec(sql); err != nil {
			return err
		}

		//Load init data
		file, err := ioutil.ReadFile("data/books.json")
		if err != nil {
			return err
		}
		// var books model.Books
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

func (this *ProductMySqlHandler) Create(product model.Product) error {
	this.SetTimeZone()
	if err := this.CreateTablesIfNotExist(); err != nil {
		return err
	}
	sql := `INSERT INTO product (id, title, price, cover, status) VALUES(?, ?, ?, ?, 1)`
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

func (this *ProductMySqlHandler) GetAll() ([]model.Product, error) {
	this.SetTimeZone()
	if err := this.CreateTablesIfNotExist(); err != nil {
		return nil, err
	}
	var products []model.Product
	rows, err := this.db.Query("SELECT id, title, price, cover, status, CONVERT_TZ(created_at,'GMT','Asia/Bangkok'), CONVERT_TZ(updated_at,'GMT','Asia/Bangkok') FROM product ORDER BY title")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var title string
		var price float32
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
