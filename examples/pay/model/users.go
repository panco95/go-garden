package model

type Order struct {
	ID      int    `gorm:"primaryKey" json:"id"`
	UserId  int    `json:"user_id"`
	OrderId string `json:"order_id"`
}
