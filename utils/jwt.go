package utils

import (
	"os"
	"errors"
	"time"
	"log"
	"strconv"

	"github.com/golang-jwt/jwt/v4"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

type JWTClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateJWT(userID, role string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateJWT(tokenString string) (int, string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("Unexpected signing method: %v\n", token.Header["alg"])
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		log.Printf("Error parsing token: %v\n", err)
		return 0, "", err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		log.Println("Invalid token claims or token is not valid")
		return 0, "", errors.New("invalid token")
	}

	log.Printf("Parsed token claims: UserID: %s, Role: %s\n", claims.UserID, claims.Role)

	userID, err := strconv.Atoi(claims.UserID)
	if err != nil || userID == 0 {
		log.Printf("Invalid UserID in token claims: %v\n", claims.UserID)
		return 0, "", errors.New("invalid UserID in token claims")
	}

	return userID, claims.Role, nil
}