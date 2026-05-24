package utils

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func GenerateUniqueString() string {
	return uuid.NewString()
}

func GeneratePrefixedUUID(prefix string) string {
	return fmt.Sprintf("%s-%s", strings.ToUpper(prefix), uuid.NewString())
}

func GenerateAdminID() string {
	return GeneratePrefixedUUID("AD")
}

func GenerateShopID() string {
	return GeneratePrefixedUUID("SH")
}
