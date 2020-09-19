package controller

import (
	"log"
	"net/http"

	"github.com/choobot/choo-pos-backend/app/service"

	"github.com/labstack/echo"
	"golang.org/x/oauth2"
)

type WebController struct {
	OAuthService   service.OAuthService
	JwtService     service.JwtService
	SessionService service.SessionService
}

type ApiError struct {
	Code  int
	Error string
}

func (this *WebController) SetNoCache(c echo.Context) {
	c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Response().Header().Set("Pragma", "no-cache")
	c.Response().Header().Set("Expires", "0")
}

func (this *WebController) Index(c echo.Context) error {
	this.SetNoCache(c)
	oauthToken := this.SessionService.Get(c, "oauthToken")
	if oauthToken == nil {
		error := ApiError{
			Code:  http.StatusUnauthorized,
			Error: "No user session, please use Login API first",
		}
		return c.JSON(http.StatusInternalServerError, error)
	}
	oauthName := this.SessionService.Get(c, "oauthName")
	oauthPicture := this.SessionService.Get(c, "oauthPicture")
	if oauthName == nil || oauthPicture == nil {
		error := ApiError{
			Code:  http.StatusInternalServerError,
			Error: "Cannot get user info from session",
		}
		return c.JSON(http.StatusInternalServerError, error)
	}
	data := map[string]string{
		"oauthName":    oauthName.(string),
		"oauthPicture": oauthPicture.(string),
	}

	return c.JSON(http.StatusOK, data)
}

func (this *WebController) Login(c echo.Context) error {
	this.SetNoCache(c)
	oauthState := this.OAuthService.GenerateOAuthState()
	this.SessionService.Set(c, "oauthState", oauthState)
	url := this.OAuthService.OAuthConfig().AuthCodeURL(oauthState, oauth2.AccessTypeOnline)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func (this *WebController) Auth(c echo.Context) error {
	this.SetNoCache(c)
	oauthState := this.SessionService.Get(c, "oauthState")

	state := c.QueryParam("state")
	if oauthState != "" && state != oauthState {
		log.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthState, state)
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}
	code := c.QueryParam("code")
	oauthToken, err := this.OAuthService.OAuthConfig().Exchange(oauth2.NoContext, code)
	this.SessionService.Set(c, "oauthToken", oauthToken.AccessToken)
	if err != nil {
		error := ApiError{
			Code:  http.StatusInternalServerError,
			Error: err.Error(),
		}
		return c.JSON(http.StatusInternalServerError, error)
	}

	idToken, err := this.JwtService.ExtractIdToken(oauthToken.Extra("id_token").(string))
	if err != nil {
		error := ApiError{
			Code:  http.StatusInternalServerError,
			Error: err.Error(),
		}
		return c.JSON(http.StatusInternalServerError, error)
	}
	this.SessionService.Set(c, "oauthId", idToken.Id)
	this.SessionService.Set(c, "oauthName", idToken.Name)
	this.SessionService.Set(c, "oauthPicture", idToken.Picture)
	return c.Redirect(http.StatusTemporaryRedirect, "/")
}

func (this *WebController) Logout(c echo.Context) error {
	this.SetNoCache(c)
	oauthToken := this.SessionService.Get(c, "oauthToken")
	if oauthToken != nil {
		err := this.OAuthService.Signout(oauthToken.(string))
		if err != nil {
			error := ApiError{
				Code:  http.StatusInternalServerError,
				Error: err.Error(),
			}
			return c.JSON(http.StatusInternalServerError, error)
		}
	}
	this.SessionService.Destroy(c)
	return c.Redirect(http.StatusTemporaryRedirect, "/")
}
