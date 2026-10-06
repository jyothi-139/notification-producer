package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"notification-producer/internals/models"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/joho/godotenv"
)

func main() {

	_ = godotenv.Load()

	connectionString := os.Getenv("SERVICEBUS_CONNECTION_STRING")
	queueName := os.Getenv("SERVICEBUS_QUEUE_NAME")

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		dbHost,
		dbPort,
		dbUser,
		dbPassword,
		dbName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Database Connection Error:", err)
		return
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		fmt.Println("Database Ping Error:", err)
		return
	}

	reader := bufio.NewReader(os.Stdin)

	notification := models.Notification{
		ID: int(time.Now().UnixMilli()),
	}

	fmt.Print("Enter Channel (EMAIL/SMS/BOTH): ")
	channel, _ := reader.ReadString('\n')
	channel = strings.ToUpper(strings.TrimSpace(channel))

	notification.Channel = channel

	switch channel {

	case "EMAIL":

		fmt.Print("Enter Recipient Email: ")
		email, _ := reader.ReadString('\n')
		notification.Recipient = strings.TrimSpace(email)

		fmt.Print("Enter Subject: ")
		subject, _ := reader.ReadString('\n')
		notification.Subject = strings.TrimSpace(subject)

		fmt.Print("Enter Message: ")
		message, _ := reader.ReadString('\n')
		notification.Message = strings.TrimSpace(message)

	case "SMS":

		fmt.Print("Enter Phone Number: ")
		phone, _ := reader.ReadString('\n')
		notification.PhoneNumber = strings.TrimSpace(phone)

		notification.Recipient = notification.PhoneNumber

		fmt.Print("Enter Message: ")
		message, _ := reader.ReadString('\n')
		notification.Message = strings.TrimSpace(message)

	case "BOTH":

		fmt.Print("Enter Recipient Email: ")
		email, _ := reader.ReadString('\n')
		notification.Recipient = strings.TrimSpace(email)

		fmt.Print("Enter Phone Number: ")
		phone, _ := reader.ReadString('\n')
		notification.PhoneNumber = strings.TrimSpace(phone)

		fmt.Print("Enter Subject: ")
		subject, _ := reader.ReadString('\n')
		notification.Subject = strings.TrimSpace(subject)

		fmt.Print("Enter Message: ")
		message, _ := reader.ReadString('\n')
		notification.Message = strings.TrimSpace(message)

	default:
		fmt.Println("Invalid Channel")
		return
	}

	_, err = db.Exec(`
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
		notification.Channel,
		notification.Recipient,
		notification.Subject,
		notification.Message,
		"PENDING",
	)

	if err != nil {
		fmt.Println("Notification Insert Error:", err)
		return
	}

	fmt.Println("✅ Notification Saved To Database")

	body, err := json.Marshal(notification)
	if err != nil {
		fmt.Println("JSON Marshal Error:", err)
		return
	}

	client, err := azservicebus.NewClientFromConnectionString(
		connectionString,
		nil,
	)
	if err != nil {
		fmt.Println("Service Bus Connection Error:", err)
		return
	}
	defer client.Close(context.Background())

	sender, err := client.NewSender(queueName, nil)
	if err != nil {
		fmt.Println("Sender Creation Error:", err)
		return
	}
	defer sender.Close(context.Background())

	msg := &azservicebus.Message{
		Body: body,
	}

	err = sender.SendMessage(
		context.Background(),
		msg,
		nil,
	)

	if err != nil {
		fmt.Println("Message Send Error:", err)

		_, _ = db.Exec(`
			UPDATE notifications
			SET status='FAILED'
			WHERE recipient=$1
		`,
			notification.Recipient,
		)

		return
	}

	_, _ = db.Exec(`
		UPDATE notifications
		SET status='QUEUED'
		WHERE recipient=$1
	`,
		notification.Recipient,
	)

	fmt.Println()
	fmt.Println("Generated Notification:")
	fmt.Println(string(body))
	fmt.Println()

	fmt.Println("✅ Notification Published Successfully")
	fmt.Println("✅ Notification Saved In notifications Table")
	fmt.Println("Notification ID:", strconv.Itoa(notification.ID))
}
