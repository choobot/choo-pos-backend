package handler

import (
	"database/sql"
	"errors"
	"os"

	"github.com/choobot/choo-pos-backend/app/model"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type OrderHandler interface {
	Create(order model.Order) (*model.Order, error)
	// GetAll() ([]model.Order, error)
	GetById(id string) (*model.Order, error)
}

type MySqlOrderHandler struct {
	db *sql.DB
}

func NewMySqlOrderHandler() MySqlOrderHandler {
	db, _ := sql.Open("mysql", os.Getenv("DATA_SOURCE_NAME"))
	return MySqlOrderHandler{
		db: db,
	}
}

func (this *MySqlOrderHandler) SetTimeZone() error {
	sql := `SET time_zone = 'Asia/Bangkok'`
	_, err := this.db.Exec(sql)
	if err != nil {
		return err
	}

	return nil
}

func (this *MySqlOrderHandler) CreateTablesIfNotExist() error {
	sql := "SELECT 1 FROM sale_order LIMIT 1"
	rows, err := this.db.Query(sql)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		sql = `
		CREATE TABLE sale_order (
			id VARCHAR(255) NOT NULL PRIMARY KEY,
			total FLOAT NOT NULL,
			subtotal FLOAT NOT NULL,
			cash FLOAT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		) CHARACTER SET utf8 COLLATE utf8_general_ci`

		if _, err = this.db.Exec(sql); err != nil {
			return err
		}
	}

	sql = "SELECT 1 FROM sale_order_item LIMIT 1"
	rows, err = this.db.Query(sql)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		sql = `
		CREATE TABLE sale_order_item (
			id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			order_id VARCHAR(255) NOT NULL,
			price FLOAT NOT NULL,
			product_id VARCHAR(255) NOT NULL,
			product_title VARCHAR(255) NOT NULL,
			product_price FLOAT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		) CHARACTER SET utf8 COLLATE utf8_general_ci`

		if _, err = this.db.Exec(sql); err != nil {
			return err
		}
	}

	return nil
}

func (this *MySqlOrderHandler) Create(order model.Order) (*model.Order, error) {
	this.SetTimeZone()
	if err := this.CreateTablesIfNotExist(); err != nil {
		return nil, err
	}
	orderId := uuid.New().String()
	sql := `INSERT INTO sale_order (id, total, subtotal, cash) VALUES(?, ?, ?, ?)`
	result, err := this.db.Exec(sql, orderId, order.Total, order.Subtotal, order.Cash)
	if err != nil {
		return nil, err
	}
	num, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if num != 1 {
		return nil, errors.New("No record")
	}

	for _, item := range order.Items {
		sql := `INSERT INTO sale_order_item (order_id, price, product_id, product_title, product_price) VALUES(?, ?, ?, ?, ?)`
		result, err := this.db.Exec(sql, orderId, item.Price, item.Product.Id, item.Product.Title, item.Product.Price)
		if err != nil {
			return nil, err
		}
		num, err := result.RowsAffected()
		if err != nil {
			return nil, err
		}
		if num != 1 {
			return nil, errors.New("No record")
		}
	}

	return this.GetById(orderId)
}

func (this *MySqlOrderHandler) GetById(id string) (*model.Order, error) {
	this.SetTimeZone()
	if err := this.CreateTablesIfNotExist(); err != nil {
		return nil, err
	}
	order := model.Order{}
	sql := "SELECT sale_order.id AS order_id, sale_order.total, sale_order.subtotal, sale_order.cash, sale_order.created_at, sale_order_item.id AS order_item_id, sale_order_item.price, sale_order_item.product_id, sale_order_item.product_title, sale_order_item.product_price FROM sale_order JOIN sale_order_item ON sale_order.id=sale_order_item.order_id WHERE sale_order.id=?"

	rows, err := this.db.Query(sql, id)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	var first = true
	for rows.Next() {
		var orderId string
		var total float64
		var subtotal float64
		var cash float64
		var createdAt string
		var orderItemId string
		var price float64
		var productId string
		var productTitle string
		var productPrice float64
		if err := rows.Scan(&orderId, &total, &subtotal, &cash, &createdAt, &orderItemId, &price, &productId, &productTitle, &productPrice); err != nil {
			return nil, err
		}
		if first {
			order = model.Order{
				Id:        orderId,
				Total:     total,
				Subtotal:  subtotal,
				Cash:      cash,
				CreatedAt: createdAt,
				Items:     []model.OrderItem{},
			}
			first = false
		}
		item := model.OrderItem{
			Id:    orderItemId,
			Price: price,
			Product: model.Product{
				Id:    productId,
				Title: productTitle,
				Price: productPrice,
			},
		}
		order.Items = append(order.Items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &order, nil
}
