package authhelper

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type AuthHelper interface {
	IssueToken(indetifier string, claims map[string]string) (string, string, error)
	ValidateToken(identifier string, token string) (bool, error)
	RefreshToken(identifier string, accessToken string, refreshToken string) (string, string, error)
}

type JwtAuthHelper struct {
	db     StorageInterface
	expiry time.Duration
	secret string
}

func GetJwtAuthHelper(secret1 string, db1 StorageInterface, expiry1 time.Duration) *JwtAuthHelper {
	j := &JwtAuthHelper{
		db:     db1,
		expiry: expiry1,
		secret: secret1,
	}
	return j
}

func (jw *JwtAuthHelper) IssueToken(indentifier string, claims map[string]string) (string, string, error) {
	var accessToken, refreshToken string
	var err error
	token := jwt.New(jwt.SigningMethodHS256)
	tokenClaims := token.Claims.(jwt.MapClaims)
	for k, v := range claims {
		log.Printf("k=%s v=%s", k, v)
		tokenClaims[k] = v
	}
	tokenClaims["exp"] = time.Now().Add(jw.expiry).Unix()
	tokenClaims["identifier"] = indentifier
	accessToken, err = token.SignedString([]byte(jw.secret))
	if err != nil {
		return "", "", err
	}
	refreshToken, err = jw.db.GetRefreshToken(indentifier)
	if err != nil || refreshToken == "" {
		b := make([]byte, 32)
		_, err = rand.Read(b)
		if err != nil {
			return "", "", err
		}
		refreshToken = base64.StdEncoding.EncodeToString(b)
		err = jw.db.StoreRefreshToken(indentifier, refreshToken)
	}
	return accessToken, refreshToken, err
}

func (jwh *JwtAuthHelper) ValidateToken(identifier string, tokenString string) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwh.secret), nil
	})
	log.Printf("token %+v", token)
	if err != nil || !token.Valid {
		return false, err
	}
	claims := token.Claims.(jwt.MapClaims)
	if v, ok := claims["identifier"].(string); ok {
		if strings.Compare(identifier, v) == 0 {
			return true, err
		}
	}
	return false, err
}

func (jwh *JwtAuthHelper) RefreshToken(identifier string, accessToken string, refreshToken string) (string, string, error) {
	rt, err := jwh.db.GetRefreshToken(identifier)
	if err != nil {
		return "", "", err
	}
	if strings.Compare(rt, refreshToken) != 0 {
		return "", "", errors.New("refresh token not valid")
	}

	if err != nil {
		return "", "", err
	}
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwh.secret), nil
	})
	log.Printf("token %+v", token)
	if err != nil {
		if e, ok := err.(*jwt.ValidationError); ok {
			log.Printf("type correct err=%+v e=%+v e.Errors=%+v vee=%+v", err, e, e.Errors, jwt.ValidationErrorExpired)
			if e.Errors&jwt.ValidationErrorExpired == 0 {
				return "", "", err
			}
		} else {
			return "", "", err
		}

	}
	claims := token.Claims.(jwt.MapClaims)
	claimMap := make(map[string]string)
	for k, v := range claims {
		if s, ok := v.(string); ok {
			claimMap[k] = s
		}
	}
	err = jwh.db.DeleteRefreshToken(identifier)
	return jwh.IssueToken(identifier, claimMap)
}
