package handler

type CreateTaskRequest struct {
	UserID  int    `json:"user_id" binding:"required,min=1"`
	Type    string `json:"type" binding:"required,oneof=email sms push"`
	Payload string `json:"payload" binding:"required"`
}
