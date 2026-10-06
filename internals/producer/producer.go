package producer

import (
	"context"
	"database/sql"
	"encoding/json"

	"notification-producer/internals/models"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

func Publish(
	db *sql.DB,
	sender *azservicebus.Sender,
	n models.Notification,
) error {

	_, err := db.Exec(`
		INSERT INTO notifications
		(
			channel,
			recipient,
			subject,
			message,
			status
		)
		VALUES ($1,$2,$3,$4,$5)
	`,
		n.Channel,
		n.Recipient,
		n.Subject,
		n.Message,
		"PENDING",
	)

	if err != nil {
		return err
	}

	payload, err := json.Marshal(n)
	if err != nil {
		return err
	}

	msg := &azservicebus.Message{
		Body: payload,
	}

	return sender.SendMessage(
		context.Background(),
		msg,
		nil,
	)
}
