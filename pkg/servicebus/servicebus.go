package servicebus

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

func NewSender(connString string, queue string) (*azservicebus.Sender, *azservicebus.Client, error) {

	client, err := azservicebus.NewClientFromConnectionString(
		connString,
		nil,
	)
	if err != nil {
		return nil, nil, err
	}

	sender, err := client.NewSender(queue, nil)
	if err != nil {
		return nil, nil, err
	}

	return sender, client, nil
}

func Close(client *azservicebus.Client) {
	client.Close(context.Background())
}
