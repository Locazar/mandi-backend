package handler

// This file is intentionally left blank as a placeholder for future notification handler implementations.
import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	usercaseInterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"google.golang.org/api/option"
)

var fcmClient *messaging.Client

type NotificationHandler struct {
	// Add necessary fields here, such as services or configurations
	notificationUsecase usercaseInterfaces.NotificationUseCase
}

func NewNotificationHandler(notificationUsecase usercaseInterfaces.NotificationUseCase) *NotificationHandler {
	InitFirebase()
	return &NotificationHandler{
		notificationUsecase: notificationUsecase,
	}
}

// Example method for sending a notification

// SaveNotification godoc
//
//	@summary 	api for sending notification
//	@Security	BearerAuth
//	@id			SaveNotification
//	@tags		Notification
//	@Param		input	body	request.Notification{}	true	"inputs"
//	@Router		/notifications/ [post]
//	@Success	200	{object}	response.Response{}	"Successfully sent notification"
//	@Failure	400	{object}	response.Response{}	"invalid input"
func (h *NotificationHandler) SaveNotification(ctx *gin.Context) {
	// Implementation for sending notification
	var body request.Notification
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// Call the usecase to send notification
	err := h.notificationUsecase.SaveNotification(ctx.Request.Context(), body)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to send notification"})
		return
	}

	// Example: send notification using FCM
	var Token = "device_registration_token_here" // Replace with actual device token

	notificationData := request.Notification{
		SenderID:     body.SenderID,
		SenderType:   body.SenderType,
		ReceiverID:   body.ReceiverID,
		ReceiverType: body.ReceiverType,
		Title:        body.Title,
		Body:         body.Body,
		Status:       "sent",
		CreatedAt:    time.Now(),
	}

	// Fetch receiver's active tokens
	notifications, err := h.notificationUsecase.SendNotificationToDevice(ctx.Request.Context(), notificationData)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to send notification to device"})
		return
	}

	fmt.Printf("Fetched %d notifications for sending\n", len(notifications))

	sendNotificationDataToDevice(Token, notificationData)
	ctx.JSON(200, gin.H{"message": "Notification sent successfully"})

}

// GetNotificationsBy godoc
//
//	@summary 	api for getting notifications with filters
//	@Security	BearerAuth
//	@id			GetNotificationsBy
//	@tags		Notification
//	@Param		filter	query	request.GetNotification	false	"filter"
//	@Param		pagination	query	request.Pagination	false	"pagination"
//	@Router		/notifications/ [get]
//	@Success	200	{object}	response.Response{}	"Successfully got notifications"
//	@Failure	400	{object}	response.Response{}	"invalid input"
func (c *NotificationHandler) GetNotificationsBy(ctx *gin.Context) {
	// Implementation for getting notifications
	var filter request.GetNotification
	var pagination request.Pagination
	ctx.JSON(200, gin.H{"message": "Get Notifications"})

	err := ctx.ShouldBindQuery(&request.Pagination{})
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// Call the usecase to get notifications
	notificationData, err := c.notificationUsecase.GetNotificationsBy(ctx.Request.Context(), filter, pagination)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to get notifications"})
		return
	}

	// Use a token placeholder or obtain the real device token(s) as needed.
	var token = "device_registration_token_here"

	// Send notifications to devices for each returned notification.
	for _, n := range notificationData {
		reqNotif := request.Notification{
			Title: n.Title,
			Body:  n.Body,
		}
		sendNotificationDataToDevice(token, reqNotif)
	}
}

func InitFirebase() {
	// Initialize Firebase app and messaging client here
	ctx := context.Background()
	creds := os.Getenv("FIREBASE_CONFIG")
	if creds == "" {
		log.Println("FIREBASE_CONFIG env var is empty; skipping Firebase initialization")
		return
	}
	opt := option.WithCredentialsJSON([]byte(creds))
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("error initializing Firebase app: %v", err)
	}
	fcmClient, err = app.Messaging(ctx)
	if err != nil {
		log.Fatalf("error getting Messaging client: %v", err)
	}
}

// MarkNotificationAsRead godoc
//
//	@summary 	api for marking notification as read
//	@Security	BearerAuth
//	@id			MarkNotificationAsRead
//	@tags		Notification
//	@Param		notification_id	path	uint	true	"Notification ID"
//	@Router		/notifications/{notification_id}/read [put]
//	@Success	200	{object}	response.Response{}	"Successfully marked notification as read"
//	@Failure	400	{object}	response.Response{}	"invalid input"
func (h *NotificationHandler) MarkNotificationAsRead(ctx *gin.Context) {
	notificationIDParam := ctx.Param("notification_id")
	var notificationID uint
	_, err := fmt.Sscan(notificationIDParam, &notificationID)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid notification ID"})
		return
	}
	// Implementation for marking notification as read
	ctx.JSON(200, gin.H{"message": "Mark Notification As Read"})

	err = h.notificationUsecase.MarkNotificationAsRead(ctx.Request.Context(), notificationID)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to mark notification as read"})
		return
	}

	ctx.JSON(200, gin.H{"message": "Notification marked as read successfully"})

}

// sendNotificationToDevice sends a notification to a device using FCM.
func sendNotificationDataToDevice(token string, notificationData request.Notification) {
	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: notificationData.Title,
			Body:  notificationData.Body,
		},
	}

	// Send a message to the device corresponding to the provided registration token.
	response, err := fcmClient.Send(context.Background(), message)
	if err != nil {
		log.Printf("Failed to send message: %v\n", err)
		return
	}
	// Response is a message ID string.
	log.Printf("Successfully sent message: %s\n", response)
}

// GenerateFCMToken godoc
//
//	@summary 	api for generating FCM token
//	@Security	BearerAuth
//	@id			GenerateFCMToken
//	@tags		Notification
//	@Param		input	body	request.NotificationDeviceToken{}	true	"inputs"
//	@Router		/notifications/generateFCMToken [post]
//	@Success	200	{object}	response.Response{}	"Successfully generated FCM token"
//	@Failure	400	{object}	response.Response{}	"invalid input"
func (h *NotificationHandler) GenerateFCMToken(ctx *gin.Context) {

	var req request.NotificationDeviceToken

	err := h.notificationUsecase.GenerateFCMToken(ctx, req)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}

	ctx.JSON(200, gin.H{"message": "Notification sent successfully"})
}
