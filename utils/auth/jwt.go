package auth

import (
	"os"
	"time"

	"github.com/YasserCherfaoui/MarketProGo/models"
	"github.com/golang-jwt/jwt"
)

type MyClaims struct {
	UserID    uint            `json:"user_id"`
	UserType  models.UserType `json:"user_type"`
	CompanyID *uint           `json:"company_id"`
	jwt.StandardClaims
}

func GenerateToken(userID uint, userType models.UserType, companyID *uint) (string, error) {
	claims := MyClaims{
		UserID:    userID,
		UserType:  userType,
		CompanyID: companyID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func ValidateToken(tokenString string) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	return token.Claims.(*MyClaims), nil
}
