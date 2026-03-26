package repository

import (
	"context"
	"testing"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type NotificationRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo interfaces.NotificationRepository
}

func (suite *NotificationRepositoryTestSuite) SetupTest() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(suite.T(), err)

	// Migrate the schema
	err = db.AutoMigrate(&domain.Notification{}, &domain.NotificationDeviceToken{})
	assert.NoError(suite.T(), err)

	suite.db = db
	suite.repo = NewNotificationRepository(db)
}

func (suite *NotificationRepositoryTestSuite) TearDownTest() {
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()
}

func (suite *NotificationRepositoryTestSuite) TestSaveNotification() {
	notification := domain.Notification{
		SenderType:   "user",
		ReceiverType: "user",
		Type:         "general",
		SenderID:     1,
		Title:        "Test Notification",
		Message:      "This is a test",
		Body:         "Test body",
		ReceiverID:   2,
		Status:       "sent",
		CreatedAt:    "2023-01-01T00:00:00Z",
		UpdatedAt:    "2023-01-01T00:00:00Z",
	}

	err := suite.repo.SaveNotification(context.Background(), notification)
	assert.NoError(suite.T(), err)
}

func (suite *NotificationRepositoryTestSuite) TestGetNotification() {
	// Save a notification first
	notification := domain.Notification{
		SenderType:   "user",
		ReceiverType: "user",
		Type:         "general",
		SenderID:     1,
		Title:        "Test Notification",
		Message:      "This is a test",
		Body:         "Test body",
		ReceiverID:   2,
		Status:       "sent",
		CreatedAt:    "2023-01-01T00:00:00Z",
		UpdatedAt:    "2023-01-01T00:00:00Z",
	}

	err := suite.repo.SaveNotification(context.Background(), notification)
	assert.NoError(suite.T(), err)

	filter := request.Notification{ReceiverID: 2}
	notifications, err := suite.repo.GetNotification(context.Background(), filter)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), notifications, 1)
	assert.Equal(suite.T(), "Test Notification", notifications[0].Title)
}

func (suite *NotificationRepositoryTestSuite) TestMarkNotificationAsRead() {
	// Save a notification first
	notification := domain.Notification{
		SenderType:   "user",
		ReceiverType: "user",
		Type:         "general",
		SenderID:     1,
		Title:        "Test Notification",
		Message:      "This is a test",
		Body:         "Test body",
		ReceiverID:   2,
		IsRead:       false,
		Status:       "sent",
		CreatedAt:    "2023-01-01T00:00:00Z",
		UpdatedAt:    "2023-01-01T00:00:00Z",
	}

	err := suite.repo.SaveNotification(context.Background(), notification)
	assert.NoError(suite.T(), err)

	// Get the ID (assuming auto increment starts at 1)
	err = suite.repo.MarkNotificationAsRead(context.Background(), 1)
	assert.NoError(suite.T(), err)
}

func (suite *NotificationRepositoryTestSuite) TestGenerateFCMToken() {
	token := request.NotificationDeviceToken{
		OwnerID:   "user123",
		OwnerType: "user",
		Token:     "fcm_token_123",
		Platform:  "android",
	}

	err := suite.repo.GenerateFCMToken(context.Background(), token)
	assert.NoError(suite.T(), err)
}

func (suite *NotificationRepositoryTestSuite) TestGetDeviceTokens() {
	token := request.NotificationDeviceToken{
		OwnerID:   "user123",
		OwnerType: "user",
		Token:     "fcm_token_123",
		Platform:  "android",
	}

	err := suite.repo.GenerateFCMToken(context.Background(), token)
	assert.NoError(suite.T(), err)

	tokens, err := suite.repo.GetDeviceTokens(context.Background(), "user123", "user")
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), tokens, 1)
	assert.Equal(suite.T(), "fcm_token_123", tokens[0].Token)
}

func TestNotificationRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationRepositoryTestSuite))
}