package main

import (
	"os"

	"github.com/choobot/choo-pos-backend/app/controller"
	"github.com/choobot/choo-pos-backend/app/handler"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
)

func main() {
	oAuthSerivce := handler.NewLineOAuthHandler()
	productHandler := handler.NewMySqlProductHandler()
	userHandler := handler.NewMySqlUserHandler()
	controller := controller.ApiController{
		OAuthHandler:     &oAuthSerivce,
		SessionHandler:   &handler.CookieSessionHandler{},
		ProductHandler:   &productHandler,
		UserHandler:      &userHandler,
		PromotionHandler: &handler.FixPromotionHandler{},
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("choo-pos"))))

	e.GET("/user", controller.User)
	e.GET("/user/login", controller.Login)
	e.GET("/user/logout", controller.Logout)
	e.GET("/user/token", controller.GetAccessToken)
	e.GET("/auth", controller.Auth)

	e.GET("/product", controller.GetAllProduct)
	e.POST("/product", controller.CreateProduct)
	// e.GET("/product/:id", controller.GetCustomer)
	// e.PUT("/product/:id", controller.UpdateProduct)
	// e.DELETE("/product/:id", controller.DeleteProduct)

	e.PUT("/cart", controller.UpdateCart)
	// e.GET("/checkout", controller.Logout)

	e.GET("/user/log", controller.GetAllUserLog)

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
