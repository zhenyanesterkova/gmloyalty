package order

import "time"

type Order struct {
	UploadTime time.Time `json:"uploaded_at"`
	Status     string    `json:"status"`
	Number     int64     `json:"number"`
	Accrual    float64   `json:"accrual"`
}
