package handler

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
)

type SessionHandler interface {
	Get(c echo.Context, name string) string
	Set(c echo.Context, name string, value string)
	Destroy(c echo.Context)
}

type CookieSessionHandler struct {
}

func (this *CookieSessionHandler) Get(c echo.Context, name string) string {
	sess, _ := session.Get("session", c)
	return sess.Values[name].(string)
}

func (this *CookieSessionHandler) Destroy(c echo.Context) {
	sess, _ := session.Get("session", c)
	sess.Options.MaxAge = -1
	sess.Save(c.Request(), c.Response())
}

func (this *CookieSessionHandler) Set(c echo.Context, name string, value string) {
	sess, _ := session.Get("session", c)
	sess.Values[name] = value
	sess.Save(c.Request(), c.Response())
}
