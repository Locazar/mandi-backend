package token

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/utils"
)

type jwtAuth struct {
	adminSecretKey string
	userSecretKey  string
}

// New TokenAuth
func NewTokenService(cfg config.Config) TokenService {

	return &jwtAuth{
		adminSecretKey: cfg.AdminAuthKey,
		userSecretKey:  cfg.UserAuthKey,
	}
}

var (
	ErrInvalidUserType    = errors.New("invalid user type")
	ErrInvalidToken       = errors.New("invalid token")
	ErrFailedToParseToken = errors.New("failed to parse token to claims")
	ErrExpiredToken       = errors.New("token expired")
)

type jwtClaims struct {
	TokenID   string
	UserID    string
	ExpiresAt time.Time
	UsedFor   UserType
	// jwt.RegisteredClaims
}

// Generate a new JWT token string from token request
func (c *jwtAuth) GenerateToken(req GenerateTokenRequest) (GenerateTokenResponse, error) {
	fmt.Printf("Generating token for userID: %d, userType: %s\n", req.UserID, req.UsedFor)
	if req.UsedFor != Admin && req.UsedFor != User {

		return GenerateTokenResponse{}, ErrInvalidUserType
	}

	tokenID := utils.GenerateUniqueString()
	claims := &jwtClaims{
		TokenID: tokenID,
		UserID:  fmt.Sprintf("%d", req.UserID),
		// RegisteredClaims: jwt.RegisteredClaims{
		// 	ExpiresAt: jwt.NewNumericDate(req.ExpirationDate),
		// },
		UsedFor:   req.UsedFor,
		ExpiresAt: req.ExpireAt,
	}
	fmt.Printf("Claims for token generation: %+v\n", claims)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	var (
		tokenString string
		err         error
	)
	// sign the token by user type
	if req.UsedFor == Admin {
		fmt.Printf("Signing token with admin secret key\n")
		tokenString, err = token.SignedString([]byte(c.adminSecretKey))
	} else {
		fmt.Printf("Signing token with user secret key\n")
		tokenString, err = token.SignedString([]byte(c.userSecretKey))
	}

	fmt.Printf("Generated token string: %s\n", tokenString)
	if err != nil {
		fmt.Printf("Error signing token: %v\n", err)
	}
	if err != nil {
		return GenerateTokenResponse{}, fmt.Errorf("failed to sign the token \nerror:%w", err)
	}

	response := GenerateTokenResponse{
		TokenID:     tokenID,
		TokenString: tokenString,
	}
	fmt.Printf("Generated token response: %+v\n", response)
	return response, nil
}

// Verify JWT token string and return TokenResponse
func (c *jwtAuth) VerifyToken(req VerifyTokenRequest) (VerifyTokenResponse, error) {

	if req.UsedFor != Admin && req.UsedFor != User {
		return VerifyTokenResponse{}, ErrInvalidUserType
	}

	token, err := jwt.ParseWithClaims(req.TokenString, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {

		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		if req.UsedFor == Admin {
			return []byte(c.adminSecretKey), nil
		}
		return []byte(c.userSecretKey), nil
	})

	if err != nil {
		if errors.Is(err, ErrExpiredToken) {
			return VerifyTokenResponse{}, ErrExpiredToken
		}
		return VerifyTokenResponse{}, ErrInvalidToken
	}
	claims, ok := token.Claims.(*jwtClaims)
	if !ok {
		return VerifyTokenResponse{}, ErrFailedToParseToken
	}

	userID, err := strconv.ParseUint(claims.UserID, 10, 64)
	if err != nil {
		return VerifyTokenResponse{}, fmt.Errorf("failed to parse user ID: %w", err)
	}

	response := VerifyTokenResponse{
		TokenID: claims.TokenID,
		UserID:  uint(userID),
	}
	return response, nil
}

func (c *jwtAuth) DecodeTokenData(tokenString string) string {
	// Remove Bearer prefix if present
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	// Try with admin secret key first
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(c.adminSecretKey), nil
	})

	if err != nil {
		// If admin key fails, try with user secret key
		token, err = jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return []byte(c.userSecretKey), nil
		})
	}

	if err != nil {
		return ""
	}
	claims, ok := token.Claims.(*jwtClaims)
	if !ok {
		return ""
	}
	return claims.UserID
}

func (c *jwtAuth) DecodeTokenDataToGetData(tokenString string) (string, UserType, error) {
	// Remove Bearer prefix if present
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	// Try with admin secret key first
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(c.adminSecretKey), nil
	})

	if err != nil {
		// If admin key fails, try with user secret key
		token, err = jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return []byte(c.userSecretKey), nil
		})
	}

	if err != nil {
		return "", "", err
	}
	claims, ok := token.Claims.(*jwtClaims)
	if !ok {
		return "", "", ErrFailedToParseToken
	}
	return claims.UserID, claims.UsedFor, nil
}

// Validate claims
func (c *jwtClaims) Valid() error {
	if time.Since(c.ExpiresAt) > 0 {
		return ErrExpiredToken
	}
	return nil
}
