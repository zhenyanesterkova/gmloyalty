package order

import "time"

const (
	StatusNew        = "NEW"
	StatusProcessing = "PROCESSING"
	StatusInvalid    = "INVALID"
	StatusProcessed  = "PROCESSED"
)

type Order struct {
	UploadTime time.Time `json:"uploaded_at"`
	Status     string    `json:"status"`
	Number     string    `json:"number"`
	Accrual    float64   `json:"accrual,omitempty"`
	UserID     int       `json:"-"`
}
