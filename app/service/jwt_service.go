package service

import (
	"errors"
	"os"

	"github.com/choobot/choo-pos-backend/app/model"

	"github.com/dgrijalva/jwt-go"
)

type JwtService interface {
	ExtractUserProfile(tokenValue string) (model.User, error)
}

func NewLineJwtService() LineJwtService {
	return LineJwtService{
		ClientSecret: os.Getenv("LINE_LOGIN_SECRET"),
	}
}

type LineJwtService struct {
	ClientSecret string
}

func (this *LineJwtService) ExtractUserProfile(tokenValue string) (model.User, error) {
	var user model.User
	token, _ := jwt.Parse(tokenValue, func(token *jwt.Token) (interface{}, error) {
		return []byte(this.ClientSecret), nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		user.Id = claims["sub"].(string)
		user.Name = claims["name"].(string)
		user.Picture = claims["picture"].(string)
	} else {
		return user, errors.New("Cannot claim user info")
	}
	return user, nil
}
