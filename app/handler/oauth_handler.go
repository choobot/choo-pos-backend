package handler

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/choobot/choo-pos-backend/app/model"
	"golang.org/x/oauth2"
)

type OAuthHandler interface {
	GenerateOAuthState(callback string) string
	GenerateAuthCodeURL(callback string) string
	ExchangeToken(code string) (string, string, *model.User, error)
	Verify(idToken string) (*model.User, error)
	Signout(oauthToken string) error
}

type LineOAuthHandler struct {
	OAuthConfig *oauth2.Config
}

func NewLineOAuthHandler() LineOAuthHandler {
	OAuthConfig := oauth2.Config{
		ClientID:     os.Getenv("LINE_LOGIN_ID"),
		ClientSecret: os.Getenv("LINE_LOGIN_SECRET"),
		Scopes:       []string{"openid", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://access.line.me/oauth2/v2.1/authorize",
			TokenURL: "https://api.line.me/oauth2/v2.1/token",
		},
	}
	return LineOAuthHandler{
		OAuthConfig: &OAuthConfig,
	}
}

func (this *LineOAuthHandler) ExchangeToken(code string) (string, string, *model.User, error) {
	accessToken := ""
	idToken := ""
	user := &model.User{}
	oauthToken, e := this.OAuthConfig.Exchange(oauth2.NoContext, code)
	if e != nil {
		return accessToken, idToken, nil, e
	}
	accessToken = string(oauthToken.AccessToken)
	idToken = oauthToken.Extra("id_token").(string)
	user, e = this.Verify(idToken)
	if e != nil {
		return accessToken, idToken, nil, e
	}
	return accessToken, idToken, user, nil
}

func (this *LineOAuthHandler) GenerateOAuthState(callback string) string {
	this.OAuthConfig.RedirectURL = callback
	salt := "choo-pos"
	data := []byte(strconv.Itoa(int(time.Now().Unix())) + salt)
	return fmt.Sprintf("%x", sha1.Sum(data))
}

func (this *LineOAuthHandler) GenerateAuthCodeURL(state string) string {
	return this.OAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (this *LineOAuthHandler) Verify(idToken string) (*model.User, error) {
	form := url.Values{}
	form.Add("id_token", idToken)
	form.Add("client_id", this.OAuthConfig.ClientID)
	req, err := http.NewRequest("POST", "https://api.line.me/oauth2/v2.1/verify", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		return nil, errors.New((string(body)))
	}
	var user model.User
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (this *LineOAuthHandler) Signout(oauthToken string) error {
	form := url.Values{}
	form.Add("access_token", oauthToken)
	form.Add("client_id", this.OAuthConfig.ClientID)
	form.Add("client_secret", this.OAuthConfig.ClientSecret)
	req, err := http.NewRequest("POST", "https://api.line.me/oauth2/v2.1/revoke", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		return errors.New((string(body)))
	}
	return nil
}
