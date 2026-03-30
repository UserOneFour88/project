package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

type Service struct {
	issuer        string
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewService(issuer, accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{
		issuer:        issuer,
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

type AccessClaims struct {
	jwt.RegisteredClaims
	UserID   int64  `json:"uid"`
	Username string `json:"uname"`
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	UserID int64 `json:"uid"`
}

func (s *Service) IssueAccess(userID int64, username string) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(s.accessTTL)
	claims := AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
		UserID:   userID,
		Username: username,
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := t.SignedString(s.accessSecret)
	return token, exp, err
}

func (s *Service) IssueRefresh(userID int64) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(s.refreshTTL)

	// Include random token id so refresh tokens are unique even for same user.
	jti, err := randomHex(16)
	if err != nil {
		return "", time.Time{}, err
	}

	claims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   fmt.Sprintf("%d", userID),
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
		UserID: userID,
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := t.SignedString(s.refreshSecret)
	return token, exp, err
}

func (s *Service) ParseAccess(token string) (AccessClaims, error) {
	var claims AccessClaims
	_, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		return s.accessSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	if err != nil {
		return AccessClaims{}, err
	}
	return claims, nil
}

func (s *Service) ParseRefresh(token string) (RefreshClaims, error) {
	var claims RefreshClaims
	_, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		return s.refreshSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	if err != nil {
		return RefreshClaims{}, err
	}
	return claims, nil
}

func randomHex(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

