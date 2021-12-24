package model

type User struct {
	ID       int    `gorm:"primaryKey" json:"id"`
	Username string `json:"username"`
}
