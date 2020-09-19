package service

import (
	"errors"
	"os"

	"github.com/dgrijalva/jwt-go"
)

type IdToken struct {
	Id      string
	Name    string
	Picture string
}

type JwtService interface {
	ExtractIdToken(tokenValue string) (IdToken, error)
}

func NewLineJwtService() LineJwtService {
	return LineJwtService{
		ClientSecret: os.Getenv("LINE_LOGIN_SECRET"),
	}
}

type LineJwtService struct {
	ClientSecret string
}

func (this *LineJwtService) ExtractIdToken(tokenValue string) (IdToken, error) {
	var idToken IdToken
	token, _ := jwt.Parse(tokenValue, func(token *jwt.Token) (interface{}, error) {
		return []byte(this.ClientSecret), nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		idToken.Id = claims["sub"].(string)
		idToken.Name = claims["name"].(string)
		idToken.Picture = claims["picture"].(string)
	} else {
		return idToken, errors.New("Cannot claim user info")
	}
	return idToken, nil
}
