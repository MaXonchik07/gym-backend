package models

import (
	"time"
)

type Class struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Duration     string `json:"duration"`
	Capacity     int    `json:"capacity"`
	Difficulty   string `json:"difficulty"`
	Instructor   string `json:"instructor"`
}

type Booking struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	ClassID    string    `json:"class_id"`
	ClassName  string    `json:"class_name"`
	Instructor string    `json:"instructor"`
	Date       string    `json:"date"`
	Time       string    `json:"time"`
	CreatedAt  time.Time `json:"created_at"`
}