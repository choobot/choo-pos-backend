package main

import (
	"os"

	"github.com/choobot/choo-pos-backend/app/controller"
	"github.com/choobot/choo-pos-backend/app/handler"
	"github.com/choobot/choo-pos-backend/app/service"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
)

func main() {
	oAuthSerivce := service.NewLineOAuthService()
	jwtService := service.NewLineJwtService()
	productHandler := handler.NewProductMySqlHandler()
	controller := controller.ApiController{
		OAuthService:   &oAuthSerivce,
		JwtService:     &jwtService,
		SessionService: &service.CookieSessionService{},
		ProductHandler: &productHandler,
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("choo-pos"))))

	e.GET("/user", controller.User)
	e.GET("/login", controller.Login)
	e.GET("/auth", controller.Auth)
	e.GET("/logout", controller.Logout)

	e.GET("/product", controller.GetAllProduct)
	e.POST("/product", controller.CreateProduct)
	// e.GET("/product/:id", controller.GetCustomer)
	// e.PUT("/product/:id", controller.UpdateProduct)
	// e.DELETE("/product/:id", controller.DeleteProduct)

	// e.GET("/cart", controller.Logout)
	// e.GET("/checkout", controller.Logout)

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
