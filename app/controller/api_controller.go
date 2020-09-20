package controller

import (
	"net/http"
	"strings"

	"github.com/choobot/choo-pos-backend/app/handler"
	"github.com/choobot/choo-pos-backend/app/model"

	"github.com/labstack/echo"
	"golang.org/x/oauth2"
)

var tokens = map[string]model.User{}

type ApiController struct {
	OAuthHandler   handler.OAuthHandler
	JwtHandler     handler.JwtHandler
	SessionHandler handler.SessionHandler
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
		token = this.SessionHandler.Get(c, "token").(string)
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
	oauthState := this.OAuthHandler.GenerateOAuthState(backendCallback)
	this.SessionHandler.Set(c, "callback", frontendCallback)
	this.SessionHandler.Set(c, "oauthState", oauthState)
	url := this.OAuthHandler.OAuthConfig().AuthCodeURL(oauthState, oauth2.AccessTypeOnline)

	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func (this *ApiController) Auth(c echo.Context) error {
	this.SetNoCache(c)
	oauthState := this.SessionHandler.Get(c, "oauthState").(string)
	callback := this.SessionHandler.Get(c, "callback").(string)
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
	oauthToken, e := this.OAuthHandler.OAuthConfig().Exchange(oauth2.NoContext, code)
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
	user, e := this.JwtHandler.ExtractUserProfile(oauthToken.Extra("id_token").(string))
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
	this.SessionHandler.Set(c, "token", token)

	return c.Redirect(http.StatusTemporaryRedirect, callback)
}

func (this *ApiController) Logout(c echo.Context) error {
	this.SetNoCache(c)
	user, err := this.verifyToken(c)
	if err != nil {
		return c.JSON(err.Error.Code, err)
	}

	oauthToken := user.Token

	if e := this.OAuthHandler.Signout(oauthToken); e != nil {
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
