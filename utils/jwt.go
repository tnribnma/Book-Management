package utils

import (
	"time"
	"github.com/golang-jwt/jwt/v5"
)
func CreateToken(userID int64,secret string) (string,error){
	claims:=jwt.MapClaims{
		"user_id":userID,
		"exp":time.Now().Add(24*time.Hour).Unix(),
	}
	t:=jwt.NewWithClaims(jwt.SigningMethodHS256,claims)
	return t.SignedString([]byte(secret))
}
func ParseToken(tokenStr,secret string) (int64,error){
	t,err:=jwt.Parse(tokenStr,func(token *jwt.Token) (interface{},error){
		return []byte(secret) ,nil
	})
	if err!=nil ||  !t.Valid{
		return 0,err
	}
	claims:=t.Claims.(jwt.MapClaims)
	return int64(claims["user_id"].(float64)),nil
}
