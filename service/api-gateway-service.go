package service

import (
	"fmt"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/nillga/jwt-server/entity"
)

type Claims struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Mail     string `json:"email"`
	IsAdmin  bool   `json:"admin"`
	jwt.StandardClaims
}

type ApiGatewayService interface {
	Authenticate(authorizationHeader string) (*entity.User, error)
}

type service struct{}

func NewApiGatewayService() ApiGatewayService {
	return &service{}
}

func (s *service) Authenticate(authorizationHeader string) (*entity.User, error) {
	if authorizationHeader == "" {
		return nil, fmt.Errorf("unauthenticated")
	}
	credentials := strings.Split(authorizationHeader, "Bearer")
	if len(credentials) != 2 {
		return nil, fmt.Errorf("invalid credential format")
	}
	token := strings.TrimSpace(credentials[1])
	if len(token) < 1 {
		return nil, fmt.Errorf("invalid credentials")
	}

	user, err := s.readToken(token)
	if err != nil {
		return nil, err
	}

	return &entity.User{
		Id:       user.Id,
		Username: user.Username,
		Email: user.Email,
		Admin:    user.Admin,
	}, nil
}

func (c *Claims) decodeJwt(token string) error {
	if _, err := jwt.ParseWithClaims(token, c, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	}); err != nil {
		return err
	}
	return nil
}

var secretKey = os.Getenv("SECRET_KEY")

func (s *service) readToken(token string) (*entity.User, error) {
	claims := &Claims{}

	if err := claims.decodeJwt(token); err != nil {
		return nil, err
	}

	return &entity.User{
		Id:       claims.Id,
		Username: claims.Username,
		Email:    claims.Mail,
		Admin:    claims.IsAdmin,
	}, nil
}
