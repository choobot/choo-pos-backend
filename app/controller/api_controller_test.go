package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/choobot/choo-pos-backend/app/handler"
	"github.com/choobot/choo-pos-backend/app/model"
	"github.com/golang/mock/gomock"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func createGoodTokenRequest() (echo.Context, *httptest.ResponseRecorder) {
	accessTokenIdTokenMap["token_value"] = "id_token_value"
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer token_value")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func createBadTokenRequest() (echo.Context, *httptest.ResponseRecorder) {
	accessTokenIdTokenMap["token_value"] = "id_token_value"
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer bad_token_value")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func TestApiController_Login(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	q := req.URL.Query()
	q.Add("callback", "callback_value")
	req.URL.RawQuery = q.Encode()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockOAuthHandler := handler.NewMockOAuthHandler(controller)
	mockOAuthHandler.EXPECT().GenerateOAuthState("https://example.com/auth").Return("state_value")
	mockOAuthHandler.EXPECT().GenerateAuthCodeURL("state_value").Return("auth_url_value")

	mockSessionHandler := handler.NewMockSessionHandler(controller)
	mockSessionHandler.EXPECT().Set(c, "callback", "callback_value").Return()
	mockSessionHandler.EXPECT().Set(c, "oauthState", "state_value").Return()

	apiController := ApiController{
		OAuthHandler:   mockOAuthHandler,
		SessionHandler: mockSessionHandler,
	}
	apiController.Login(c)

	assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
	assert.Equal(t, "auth_url_value", rec.Header().Get("Location"))
}

func TestApiController_Auth(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	q := req.URL.Query()
	q.Add("state", "state_value")
	q.Add("code", "code_value")
	req.URL.RawQuery = q.Encode()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockOAuthHandler := handler.NewMockOAuthHandler(controller)
	mockUserHandler := handler.NewMockUserHandler(controller)

	mockSessionHandler := handler.NewMockSessionHandler(controller)
	mockSessionHandler.EXPECT().Get(c, "oauthState").Return("wrong_state_value")
	mockSessionHandler.EXPECT().Get(c, "callback").Return("callback_value")

	apiController := ApiController{
		OAuthHandler:   mockOAuthHandler,
		SessionHandler: mockSessionHandler,
		UserHandler:    mockUserHandler,
	}
	apiController.Auth(c)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	mockSessionHandler.EXPECT().Get(c, "oauthState").Return("state_value")
	mockSessionHandler.EXPECT().Get(c, "callback").Return("callback_value")
	mockOAuthHandler.EXPECT().ExchangeToken("code_value").Return("", "", nil, errors.New("error_value"))

	apiController.Auth(c)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	mockSessionHandler.EXPECT().Get(c, "oauthState").Return("state_value")
	mockSessionHandler.EXPECT().Get(c, "callback").Return("callback_value")
	mockOAuthHandler.EXPECT().ExchangeToken("code_value").Return("access_token_value", "id_token_value", &model.User{}, nil)
	mockUserHandler.EXPECT().CreateLog(gomock.Any()).Return(nil)

	apiController.Auth(c)

	assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
}

func TestApiController_Logout(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	c, rec := createBadTokenRequest()

	mockOAuthHandler := handler.NewMockOAuthHandler(controller)

	apiController := ApiController{
		OAuthHandler: mockOAuthHandler,
	}
	apiController.Logout(c)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	c, rec = createGoodTokenRequest()
	mockOAuthHandler.EXPECT().Verify("id_token_value").Return(&model.User{}, nil)
	mockOAuthHandler.EXPECT().Signout("token_value").Return(errors.New("error_value"))

	apiController.Logout(c)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	c, rec = createGoodTokenRequest()
	mockOAuthHandler.EXPECT().Verify("id_token_value").Return(&model.User{}, nil)
	mockOAuthHandler.EXPECT().Signout("token_value").Return(nil)

	apiController.Logout(c)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestApiController_GetAccessToken(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	q := req.URL.Query()
	q.Add("visa", "visa_value")
	req.URL.RawQuery = q.Encode()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	apiController := ApiController{}

	apiController.GetAccessToken(c)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	visaTokenMap["visa_value"] = "token_value"

	apiController.GetAccessToken(c)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestApiController_GetUserInfo(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	c, rec := createGoodTokenRequest()
	delete(accessTokenIdTokenMap, "token_value")
	mockOAuthHandler := handler.NewMockOAuthHandler(controller)

	apiController := ApiController{
		OAuthHandler: mockOAuthHandler,
	}
	apiController.GetUserInfo(c)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	c, rec = createGoodTokenRequest()
	mockOAuthHandler.EXPECT().Verify("id_token_value").Return(&model.User{}, nil)

	apiController.GetUserInfo(c)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestApiController_CreateProduct(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	c, rec := createGoodTokenRequest()
	mockOAuthHandler := handler.NewMockOAuthHandler(controller)
	mockProductHandler := handler.NewMockProductHandler(controller)
	mockOAuthHandler.EXPECT().Verify("id_token_value").Return(&model.User{}, nil)
	mockProductHandler.EXPECT().Create(gomock.Any()).Return(nil)

	apiController := ApiController{
		OAuthHandler:   mockOAuthHandler,
		ProductHandler: mockProductHandler,
	}

	apiController.CreateProduct(c)

	assert.Equal(t, http.StatusOK, rec.Code)

	c, rec = createGoodTokenRequest()
	mockOAuthHandler.EXPECT().Verify("id_token_value").Return(&model.User{}, nil)
	mockProductHandler.EXPECT().Create(gomock.Any()).Return(errors.New("error_value"))

	apiController.CreateProduct(c)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	c, rec = createBadTokenRequest()

	apiController.CreateProduct(c)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	jsonMsg := `{"bad_json"}`
	accessTokenIdTokenMap["token_value"] = "id_token_value"
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(jsonMsg))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer token_value")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	mockOAuthHandler.EXPECT().Verify("id_token_value").Return(&model.User{}, nil)

	apiController.CreateProduct(c)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestApiController_GetAllProduct(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	c, rec := createGoodTokenRequest()
	mockOAuthHandler := handler.NewMockOAuthHandler(controller)
	mockProductHandler := handler.NewMockProductHandler(controller)
	mockPromotionHandler := handler.NewMockPromotionHandler(controller)
	mockOAuthHandler.EXPECT().Verify("id_token_value").Return(&model.User{}, nil)
	mockProductHandler.EXPECT().GetAll().Return([]model.Product{}, nil)
	mockPromotionHandler.EXPECT().AddPromotionDetailToProduct(gomock.Any()).Return([]model.Product{}, nil)

	apiController := ApiController{
		OAuthHandler:     mockOAuthHandler,
		ProductHandler:   mockProductHandler,
		PromotionHandler: mockPromotionHandler,
	}

	apiController.GetAllProduct(c)

	assert.Equal(t, http.StatusOK, rec.Code)

	c, rec = createGoodTokenRequest()
	mockOAuthHandler.EXPECT().Verify("id_token_value").Return(&model.User{}, nil)
	mockProductHandler.EXPECT().GetAll().Return(nil, errors.New("error_value"))

	apiController.GetAllProduct(c)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	c, rec = createGoodTokenRequest()
	mockOAuthHandler.EXPECT().Verify("id_token_value").Return(&model.User{}, nil)
	mockProductHandler.EXPECT().GetAll().Return([]model.Product{}, nil)
	mockPromotionHandler.EXPECT().AddPromotionDetailToProduct(gomock.Any()).Return(nil, errors.New("error_value"))

	apiController.GetAllProduct(c)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	c, rec = createBadTokenRequest()

	apiController.GetAllProduct(c)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestApiController_GetAllUserLog(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	c, rec := createGoodTokenRequest()
	mockOAuthHandler := handler.NewMockOAuthHandler(controller)
	mockUserHandler := handler.NewMockUserHandler(controller)
	mockOAuthHandler.EXPECT().Verify("id_token_value").Return(&model.User{}, nil)
	mockUserHandler.EXPECT().GetAllLog().Return([]model.UserLog{}, nil)

	apiController := ApiController{
		OAuthHandler: mockOAuthHandler,
		UserHandler:  mockUserHandler,
	}

	apiController.GetAllUserLog(c)

	assert.Equal(t, http.StatusOK, rec.Code)

	c, rec = createGoodTokenRequest()
	mockOAuthHandler.EXPECT().Verify("id_token_value").Return(&model.User{}, nil)
	mockUserHandler.EXPECT().GetAllLog().Return(nil, errors.New("error_value"))

	apiController.GetAllUserLog(c)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	c, rec = createBadTokenRequest()

	apiController.GetAllUserLog(c)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestApiController_UpdateCart(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	jsonMsg := `{
		"items": [
			{
				"product": {
					"id": "9781408855676"
				}
			}
		]
	}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(jsonMsg))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer token_value")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	mockOAuthHandler := handler.NewMockOAuthHandler(controller)
	mockPromotionHandler := handler.NewMockPromotionHandler(controller)
	mockProductHandler := handler.NewMockProductHandler(controller)
	mockOAuthHandler.EXPECT().Verify("id_token_value").Return(&model.User{}, nil)
	order := model.Order{}
	json.Unmarshal([]byte(jsonMsg), &order)
	mockPromotionHandler.EXPECT().CalculateDiscount(gomock.Any(), gomock.Any()).Return(&order, nil)
	mockProductHandler.EXPECT().GetByIds(gomock.Any()).Return(map[string]model.Product{}, nil)

	apiController := ApiController{
		OAuthHandler:     mockOAuthHandler,
		ProductHandler:   mockProductHandler,
		PromotionHandler: mockPromotionHandler,
	}

	apiController.UpdateCart(c)

	assert.Equal(t, http.StatusOK, rec.Code)

	e = echo.New()
	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader(jsonMsg))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer token_value")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	mockOAuthHandler.EXPECT().Verify("id_token_value").Return(&model.User{}, nil)
	mockProductHandler.EXPECT().GetByIds(gomock.Any()).Return(map[string]model.Product{}, nil)
	mockPromotionHandler.EXPECT().CalculateDiscount(gomock.Any(), gomock.Any()).Return(nil, errors.New("error_value"))

	apiController.UpdateCart(c)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	e = echo.New()
	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader(jsonMsg))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer token_value")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	mockOAuthHandler.EXPECT().Verify("id_token_value").Return(&model.User{}, nil)
	mockProductHandler.EXPECT().GetByIds(gomock.Any()).Return(nil, errors.New("error_value"))

	apiController.UpdateCart(c)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	c, rec = createBadTokenRequest()

	apiController.UpdateCart(c)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	jsonMsg = `{"bad_json"}`
	accessTokenIdTokenMap["token_value"] = "id_token_value"
	e = echo.New()
	req = httptest.NewRequest(http.MethodGet, "/", strings.NewReader(jsonMsg))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer token_value")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	mockOAuthHandler.EXPECT().Verify("id_token_value").Return(&model.User{}, nil)

	apiController.UpdateCart(c)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
