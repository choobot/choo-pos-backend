package main

import (
	"os"

	"github.com/choobot/choo-pos-backend/app/controller"
	"github.com/choobot/choo-pos-backend/app/service"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
)

func main() {
	oAuthSerivce := service.NewLineOAuthService()
	jwtService := service.NewLineJwtService()
	webController := controller.WebController{
		OAuthService:   &oAuthSerivce,
		JwtService:     &jwtService,
		SessionService: &service.CookieSessionService{},
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("choo-pos"))))

	e.GET("/", webController.Index)
	e.GET("/login", webController.Login)
	e.GET("/auth", webController.Auth)
	e.GET("/logout", webController.Logout)

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
