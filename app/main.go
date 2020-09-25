package main

import (
	"os"

	"github.com/choobot/choo-pos-backend/app/controller"
	"github.com/choobot/choo-pos-backend/app/handler"
	"github.com/choobot/choo-pos-backend/app/validate"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
	validator "gopkg.in/go-playground/validator.v10"
)

func main() {
	oAuthSerivce := handler.NewLineOAuthHandler()
	productHandler := handler.NewMySqlProductHandler()
	userHandler := handler.NewMySqlUserHandler()
	orderHandler := handler.NewMySqlOrderHandler()
	controller := controller.ApiController{
		OAuthHandler:     &oAuthSerivce,
		SessionHandler:   &handler.CookieSessionHandler{},
		ProductHandler:   &productHandler,
		UserHandler:      &userHandler,
		PromotionHandler: &handler.FixPromotionHandler{},
		OrderHandler:     &orderHandler,
	}

	e := echo.New()
	e.Validator = &validate.CustomValidator{Validator: validator.New()}

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("choo-pos"))))
	e.Use(middleware.CORS())

	e.GET("/user", controller.GetUserInfo)
	e.GET("/user/login", controller.Login)
	e.GET("/user/logout", controller.Logout)
	e.GET("/user/token", controller.GetAccessToken)
	e.GET("/auth", controller.Auth)

	e.GET("/product", controller.GetAllProduct)
	e.POST("/product", controller.CreateProduct)

	e.PUT("/cart", controller.UpdateCart)

	e.POST("/order", controller.CreateOrder)
	e.GET("/order/:id", controller.GetOrder)
	e.GET("/order", controller.GetAllOrder)

	e.GET("/user/log", controller.GetAllUserLog)

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
