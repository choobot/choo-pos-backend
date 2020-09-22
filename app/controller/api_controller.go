package controller

import (
	"net/http"
	"strings"

	"github.com/choobot/choo-pos-backend/app/handler"
	"github.com/choobot/choo-pos-backend/app/model"

	"github.com/google/uuid"
	"github.com/labstack/echo"
)

var visaTokenMap = map[string]string{}
var accessTokenIdTokenMap = map[string]string{}

type ApiController struct {
	OAuthHandler     handler.OAuthHandler
	SessionHandler   handler.SessionHandler
	ProductHandler   handler.ProductHandler
	UserHandler      handler.UserHandler
	PromotionHandler handler.PromotionHandler
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

func (this *ApiController) GetAccessTokenFromHeader(c echo.Context) string {
	accessToken := c.Request().Header.Get("Authorization")
	accessToken = strings.ReplaceAll(accessToken, "Bearer ", "")
	return accessToken
}

func (this *ApiController) VerifyAccessToken(c echo.Context) (*model.User, *ApiError) {
	accessToken := this.GetAccessTokenFromHeader(c)
	if idToken, ok := accessTokenIdTokenMap[accessToken]; ok {
		if user, err := this.OAuthHandler.Verify(idToken); err == nil {
			return user, nil
		}
	}
	err := &ApiError{
		Error: Error{
			Code:  http.StatusUnauthorized,
			Error: "Invalid user token, please use Login API first",
		},
	}
	return nil, err
}

func (this *ApiController) GetAccessToken(c echo.Context) error {
	this.SetNoCache(c)
	visa := c.QueryParam("visa")
	if accessToken, ok := visaTokenMap[visa]; ok {
		delete(visaTokenMap, visa)
		return c.JSONPretty(http.StatusOK, map[string]string{
			"token": accessToken,
		}, "  ")
	}
	err := &ApiError{
		Error: Error{
			Code:  http.StatusUnauthorized,
			Error: "Invalid visa, please use Login API first",
		},
	}
	return c.JSON(err.Error.Code, err)
}

func (this *ApiController) GetUserInfo(c echo.Context) error {
	this.SetNoCache(c)
	user, err := this.VerifyAccessToken(c)
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
	url := this.OAuthHandler.GenerateAuthCodeURL(oauthState)

	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func (this *ApiController) Auth(c echo.Context) error {
	this.SetNoCache(c)
	oauthState := this.SessionHandler.Get(c, "oauthState")
	callback := this.SessionHandler.Get(c, "callback")
	state := c.QueryParam("state")
	if oauthState == "" || state != oauthState {
		err := ApiError{
			Error: Error{
				Code:  http.StatusUnauthorized,
				Error: "Invalid oauth state",
			},
		}
		return c.JSON(err.Error.Code, err)
	}
	code := c.QueryParam("code")

	accessToken, idToken, user, e := this.OAuthHandler.ExchangeToken(code)
	if e != nil {
		err := ApiError{
			Error: Error{
				Code:  http.StatusInternalServerError,
				Error: e.Error(),
			},
		}
		return c.JSON(err.Error.Code, err)
	}
	visa := uuid.New().String()
	visaTokenMap[visa] = accessToken
	accessTokenIdTokenMap[accessToken] = idToken
	this.UserHandler.CreateLog(*user)

	callback += "?visa=" + visa

	return c.Redirect(http.StatusTemporaryRedirect, callback)
}

func (this *ApiController) Logout(c echo.Context) error {
	this.SetNoCache(c)
	if _, err := this.VerifyAccessToken(c); err != nil {
		return c.JSON(err.Error.Code, err)
	}

	accessToken := this.GetAccessTokenFromHeader(c)

	if e := this.OAuthHandler.Signout(accessToken); e != nil {
		err := ApiError{
			Error: Error{
				Code:  http.StatusInternalServerError,
				Error: e.Error(),
			},
		}
		return c.JSON(err.Error.Code, err)
	}
	delete(accessTokenIdTokenMap, accessToken)
	return c.NoContent(http.StatusOK)
}

func (this *ApiController) CreateProduct(c echo.Context) error {
	this.SetNoCache(c)
	if _, err := this.VerifyAccessToken(c); err != nil {
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
	if _, err := this.VerifyAccessToken(c); err != nil {
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

func (this *ApiController) GetAllUserLog(c echo.Context) error {
	this.SetNoCache(c)
	if _, err := this.VerifyAccessToken(c); err != nil {
		return c.JSON(err.Error.Code, err)
	}

	userLogs, e := this.UserHandler.GetAllLog()
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
		"user_logs": userLogs,
	})
}

func (this *ApiController) UpdateCart(c echo.Context) error {
	this.SetNoCache(c)
	if _, err := this.VerifyAccessToken(c); err != nil {
		return c.JSON(err.Error.Code, err)
	}

	order := &model.Order{}
	if e := c.Bind(order); e != nil {
		err := ApiError{
			Error: Error{
				Code:  http.StatusBadRequest,
				Error: e.Error(),
			},
		}
		return c.JSON(err.Error.Code, err)
	}

	if len(order.Items) > 0 {
		// Load Product detail
		var ids []interface{}
		for _, item := range order.Items {
			ids = append(ids, item.Product.Id)
		}

		productsMap, e := this.ProductHandler.GetByIds(ids)
		if e != nil {
			err := ApiError{
				Error: Error{
					Code:  http.StatusInternalServerError,
					Error: e.Error(),
				},
			}
			return c.JSON(err.Error.Code, err)
		}

		for i, item := range order.Items {
			order.Items[i].Product = productsMap[item.Product.Id]
			order.Items[i].Price = order.Items[i].Product.Price
		}

		// Calculate Discount
		order, e = this.PromotionHandler.CalculateDiscount(order, productsMap)
		if e != nil {
			err := ApiError{
				Error: Error{
					Code:  http.StatusInternalServerError,
					Error: e.Error(),
				},
			}
			return c.JSON(err.Error.Code, err)
		}
	}

	// Sum Total
	total := 0.0
	for _, item := range order.Items {
		total += item.Price
	}
	order.Total = total

	return c.JSONPretty(http.StatusOK, order, "  ")
}
