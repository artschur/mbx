package models

import "time"

type WhatsappBodyDTO struct {
	To          string `json:"to"`
	Body        string `json:"body"`
	TimeFromNow string `json:"time_from_now"`
}
type WhatsappBody struct {
	To          string     `json:"to"`
	Body        string     `json:"body"`
	TimeFromNow *time.Time `json:"time_from_now"`
}
