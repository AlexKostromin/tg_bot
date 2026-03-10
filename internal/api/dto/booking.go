// internal/api/dto/booking.go

package dto

type UpdateBookingStatus struct {
	Status  string `json:"status"  validate:"required,oneof=pending confirmed cancelled completed"`
	Comment string `json:"comment,omitempty"`
}

type BookingResponse struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	UserName    string `json:"user_name"`
	UserPhone   string `json:"user_phone"`
	SlotID      int    `json:"slot_id"`
	SlotDate    string `json:"slot_date"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	SubjectName string `json:"subject_name"`
	Status      string `json:"status"`
	Comment     string `json:"comment,omitempty"`
	BookedAt    string `json:"booked_at"`
}

type BookingListResponse struct {
	Items interface{} `json:"items"`
	Total int         `json:"total"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
}
