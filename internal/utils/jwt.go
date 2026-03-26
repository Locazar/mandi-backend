package utils

import (
    "errors"
    "time"

    "github.com/golang-jwt/jwt/v4"
)

type Claims struct {
    UserID uint `json:"user_id"`
    jwt.RegisteredClaims
}

func GenerateTokenPair(userID uint, accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) (accessToken string, refreshToken string, err error) {
    now := time.Now()
    accessClaims := Claims{
        UserID: userID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(now.Add(accessTTL)),
            IssuedAt:  jwt.NewNumericDate(now),
        },
    }
    at := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
    accessToken, err = at.SignedString([]byte(accessSecret))
    if err != nil {
        return
    }

    refreshClaims := Claims{
        UserID: userID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(now.Add(refreshTTL)),
            IssuedAt:  jwt.NewNumericDate(now),
        },
    }
    rt := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
    refreshToken, err = rt.SignedString([]byte(refreshSecret))
    return
}

func ParseToken(tokenStr, secret string) (*Claims, error) {
    tok, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(secret), nil
    })
    if err != nil {
        return nil, err
    }
    if claims, ok := tok.Claims.(*Claims); ok && tok.Valid {
        return claims, nil
    }
    return nil, errors.New("invalid token")
}
