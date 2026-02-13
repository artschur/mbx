package models

type SentMessage struct {
	ID           string `json:"id"`
	To           string `json:"to"`
	Body         string `json:"body"`
	DateSent     string `json:"date_sent,omitempty"`
	ErrorCode    int    `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
	Status       string `json:"status,omitempty"`
	Price        string `json:"price,omitempty"`
	PriceUnit    string `json:"price_unit,omitempty"`
}
