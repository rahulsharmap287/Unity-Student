package main

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("unity_student_secret")

func GenerateJWT(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username, // 👈 FIX: 'userId' ki jagah 'username' rakhein
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}
