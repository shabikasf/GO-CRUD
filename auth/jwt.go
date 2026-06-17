package auth

import (
	"time"
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

var SECRET = []byte("23456uiolesdfghnmertyuksedfghbnmwdfgvhbnm")
var ErrInvalidToken = errors.New("invalid token")


type Claims struct{
	UserID int `json:"user_id"`
	Username string `json:"username"`
	Role string `json:"role"`
	jwt.RegisteredClaims 
	
}

func GenerateToken(userID int, username string, role string) (string, error){
	claims := Claims{
		UserID: userID,
		Username: username,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	tokens := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return tokens.SignedString(SECRET)

}

func ValidateToken(tokenString string) (*Claims, error){
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token)(interface{}, error){
			return SECRET, nil
		},
	)

	if err != nil{
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)

	if !ok{
		return nil, ErrInvalidToken
	}

	return claims, nil

}