package models

type Todo struct {
	ID uint `gorm:"primaryKey"`

	Text string
	Done bool

	UserID uint
}

type User struct {
	ID uint `gorm:"primaryKey"`

	Name string
}
