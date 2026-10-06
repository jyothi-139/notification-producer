package models

type Notification struct {
	ID          int    `json:"id"`
	Channel     string `json:"channel"`
	Recipient   string `json:"recipient"`
	PhoneNumber string `json:"phoneNumber"`
	Subject     string `json:"subject"`
	Message     string `json:"message"`
}
