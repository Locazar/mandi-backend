package utils

import (
    "golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
    b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(b), err
}

func CompareHashAndPassword(hash, password string) error {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
