package controller

import (
	"net/http"
	"strings"

	"github.com/choobot/choo-pos-backend/app/handler"
	"github.com/choobot/choo-pos-backend/app/model"
	"github.com/choobot/choo-pos-backend/app/service"

	"github.com/labstack/echo"
	"golang.org/x/oauth2"
)

var tokens = map[string]model.User{}

type ApiController struct {
	OAuthService   service.OAuthService
	JwtService     service.JwtService
	SessionService service.SessionService
	ProductHandler handler.ProductHandler
}

type ApiError struct {
	Error Error `json:"error"`
}

type Error struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

func (this *ApiController) SetNoCache(c echo.Context) {
	c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Response().Header().Set("Pragma", "no-cache")
	c.Response().Header().Set("Expires", "0")
}

func (this *ApiController) verifyToken(c echo.Context) (*model.User, *ApiError) {
	token := c.Request().Header.Get("Authorization")
	if token == "" {
		token = this.SessionService.Get(c, "token").(string)
	} else {
		token = strings.ReplaceAll(token, "Bearer ", "")
	}
	if user, ok := tokens[token]; ok {
		return &user, nil
	}
	err := &ApiError{
		Error: Error{
			Code:  http.StatusUnauthorized,
			Error: "No user token, please use Login API first",
		},
	}
	return nil, err
}

func (this *ApiController) User(c echo.Context) error {
	this.SetNoCache(c)
	user, err := this.verifyToken(c)
	if err != nil {
		return c.JSON(err.Error.Code, err)
	}
	return c.JSONPretty(http.StatusOK, user, "  ")
}

func (this *ApiController) Login(c echo.Context) error {
	this.SetNoCache(c)
	frontendCallback := c.QueryParam("callback")
	backendCallback := "https://" + c.Request().Host + "/auth"
	oauthState := this.OAuthService.GenerateOAuthState(backendCallback)
	this.SessionService.Set(c, "callback", frontendCallback)
	this.SessionService.Set(c, "oauthState", oauthState)
	url := this.OAuthService.OAuthConfig().AuthCodeURL(oauthState, oauth2.AccessTypeOnline)

	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func (this *ApiController) Auth(c echo.Context) error {
	this.SetNoCache(c)
	oauthState := this.SessionService.Get(c, "oauthState").(string)
	callback := this.SessionService.Get(c, "callback").(string)
	state := c.QueryParam("state")
	if oauthState != "" && state != oauthState {
		err := ApiError{
			Error: Error{
				Code:  http.StatusUnauthorized,
				Error: "Invalid oauth state",
			},
		}
		return c.JSON(err.Error.Code, err)
	}
	code := c.QueryParam("code")
	oauthToken, e := this.OAuthService.OAuthConfig().Exchange(oauth2.NoContext, code)
	if e != nil {
		err := ApiError{
			Error: Error{
				Code:  http.StatusInternalServerError,
				Error: e.Error(),
			},
		}
		return c.JSON(err.Error.Code, err)
	}
	token := string(oauthToken.AccessToken)
	user, e := this.JwtService.ExtractUserProfile(oauthToken.Extra("id_token").(string))
	user.Token = token
	if e != nil {
		err := ApiError{
			Error: Error{
				Code:  http.StatusInternalServerError,
				Error: e.Error(),
			},
		}
		return c.JSON(err.Error.Code, err)
	}
	tokens[token] = user
	this.SessionService.Set(c, "token", token)

	return c.Redirect(http.StatusTemporaryRedirect, callback)
}

func (this *ApiController) Logout(c echo.Context) error {
	this.SetNoCache(c)
	user, err := this.verifyToken(c)
	if err != nil {
		return c.JSON(err.Error.Code, err)
	}

	oauthToken := user.Token

	if e := this.OAuthService.Signout(oauthToken); e != nil {
		err := ApiError{
			Error: Error{
				Code:  http.StatusInternalServerError,
				Error: e.Error(),
			},
		}
		return c.JSON(err.Error.Code, err)
	}
	delete(tokens, user.Token)
	return c.NoContent(http.StatusOK)
}

func (this *ApiController) CreateProduct(c echo.Context) error {
	this.SetNoCache(c)
	_, err := this.verifyToken(c)
	if err != nil {
		return c.JSON(err.Error.Code, err)
	}

	product := &model.Product{}
	if e := c.Bind(product); e != nil {
		err := ApiError{
			Error: Error{
				Code:  http.StatusBadRequest,
				Error: e.Error(),
			},
		}
		return c.JSON(err.Error.Code, err)
	}
	if e := this.ProductHandler.Create(*product); e != nil {
		err := ApiError{
			Error: Error{
				Code:  http.StatusInternalServerError,
				Error: e.Error(),
			},
		}
		return c.JSON(err.Error.Code, err)
	}

	return c.NoContent(http.StatusOK)
}

func (this *ApiController) GetAllProduct(c echo.Context) error {
	this.SetNoCache(c)
	_, err := this.verifyToken(c)
	if err != nil {
		return c.JSON(err.Error.Code, err)
	}

	products, e := this.ProductHandler.GetAll()
	if e != nil {
		err := ApiError{
			Error: Error{
				Code:  http.StatusInternalServerError,
				Error: e.Error(),
			},
		}
		return c.JSON(err.Error.Code, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"books": products,
	})
}
