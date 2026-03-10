// internal/api/dto/slot.go

package dto

type CreateSlotRequest struct {
	TutorID      int    `json:"tutor_id"       validate:"required"`
	SubjectID    int    `json:"subject_id"     validate:"required"`
	ClassGroupID int    `json:"class_group_id" validate:"required"`
	SlotDate     string `json:"slot_date"      validate:"required,datetime=2006-01-02"`
	StartTime    string `json:"start_time"     validate:"required,datetime=15:04"`
	EndTime      string `json:"end_time"       validate:"required,datetime=15:04"`
}

type BulkCreateSlotRequest struct {
	Slots []CreateSlotRequest `json:"slots" validate:"required,dive"`
}

type SlotResponse struct {
	ID           int    `json:"id"`
	TutorID      int    `json:"tutor_id"`
	TutorName    string `json:"tutor_name"`
	SubjectID    int    `json:"subject_id"`
	SubjectName  string `json:"subject_name"`
	ClassGroupID int    `json:"class_group_id"`
	GroupName    string `json:"group_name"`
	SlotDate     string `json:"slot_date"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	IsAvailable  bool   `json:"is_available"`
}

type SlotListResponse struct {
	Items interface{} `json:"items"`
	Total int         `json:"total"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
}
