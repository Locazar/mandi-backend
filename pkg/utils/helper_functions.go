package utils

import (
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// take userId from context
func GetUserIdFromContext(ctx *gin.Context) uint {
	userID := ctx.GetUint("userId")
	return userID
}

func StringToUint(str string) (uint, error) {
	val, err := strconv.Atoi(str)
	return uint(val), err
}

// generate userName
func GenerateRandomUserName(FirstName string) string {

	suffix := make([]byte, 4)

	numbers := "1234567890"
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))

	for i := range suffix {
		suffix[i] = numbers[rng.Intn(10)]
	}

	userName := (FirstName + string(suffix))

	return strings.ToLower(userName)
}

// generate unique string for sku
func GenerateSKU() string {
	sku := make([]byte, 10)

	rand.Read(sku)

	return hex.EncodeToString(sku)
}

// random coupons
func GenerateCouponCode(couponCodeLength int) string {
	// letter for coupons
	letters := `ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890`
	rand.Seed(time.Now().UnixMilli())

	// create a byte array of couponCodeLength
	couponCode := make([]byte, couponCodeLength)

	// loop through the array and randomly pic letter and add to array
	for i := range couponCode {
		couponCode[i] = letters[rand.Intn(len(letters))]
	}
	// convert into string and return the random letter array
	return string(couponCode)
}

func StringToTime(timeString string) (timeValue time.Time, err error) {

	// parse the string time to time
	timeValue, err = time.Parse(time.RFC3339Nano, timeString)

	if err != nil {
		return timeValue, fmt.Errorf("faild to parse given time %v to time variable \nivalid input", timeString)
	}
	return timeValue, err
}

func GenerateRandomString(length int) string {
	sku := make([]byte, length)

	rand.Read(sku)

	return hex.EncodeToString(sku)
}

func RandomInt(min, max int) int {
	rand.Seed(time.Hour.Nanoseconds())

	return rand.Intn(max-min) + min
}

func GetHashedPassword(password string) (hashedPassword string, err error) {

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return hashedPassword, err
	}
	hashedPassword = string(hash)
	return hashedPassword, nil
}

func ComparePasswordWithHashedPassword(actualpassword, hashedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(actualpassword))
	return err
}

// SaveFileLocally saves an uploaded file to the local filesystem
// Returns the relative path within the project for storing in the database
func SaveFileLocally(fileHeader *multipart.FileHeader, baseDir string) (string, error) {
	// Create the base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate unique filename using UUID to avoid conflicts
	ext := filepath.Ext(fileHeader.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Full path where file will be saved
	fullPath := filepath.Join(baseDir, filename)

	// Open the uploaded file
	src, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create the destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy the uploaded file to the destination
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// Return the relative path within the project for storing in database
	return fullPath, nil
}
