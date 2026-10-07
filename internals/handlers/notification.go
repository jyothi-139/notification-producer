package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"notification-producer/internals/models"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/gin-gonic/gin"
)

func CreateNotification(db *sql.DB) gin.HandlerFunc {

	return func(c *gin.Context) {

		var notification models.Notification

		if err := c.ShouldBindJSON(&notification); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		notification.ID = int(time.Now().UnixMilli())

		_, err := db.Exec(`
			INSERT INTO notifications
			(
				channel,
				recipient,
				phone_number,
				subject,
				message,
				status
			)
			VALUES ($1,$2,$3,$4,$5,$6)
		`,
			notification.Channel,
			notification.Recipient,
			notification.PhoneNumber,
			notification.Subject,
			notification.Message,
			"PENDING",
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		body, err := json.Marshal(notification)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		connectionString := os.Getenv("SERVICEBUS_CONNECTION_STRING")
		queueName := os.Getenv("SERVICEBUS_QUEUE_NAME")

		client, err := azservicebus.NewClientFromConnectionString(
			connectionString,
			nil,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		defer client.Close(context.Background())

		sender, err := client.NewSender(
			queueName,
			nil,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		defer sender.Close(context.Background())

		err = sender.SendMessage(
			context.Background(),
			&azservicebus.Message{
				Body: body,
			},
			nil,
		)

		if err != nil {

			_, _ = db.Exec(`
				UPDATE notifications
				SET status='FAILED'
				WHERE id=$1
			`,
				notification.ID,
			)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		_, _ = db.Exec(`
			UPDATE notifications
			SET status='QUEUED'
			WHERE id=$1
		`,
			notification.ID,
		)

		c.JSON(http.StatusOK, gin.H{
			"id":      notification.ID,
			"status":  "QUEUED",
			"message": "Notification Published Successfully",
		})
	}
}
