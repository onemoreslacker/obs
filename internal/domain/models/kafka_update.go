package models

type KafkaUpdate struct {
	ChatID      int64  `json:"chatId"`
	Url         string `json:"url"`
	Description string `json:"description"`
}
