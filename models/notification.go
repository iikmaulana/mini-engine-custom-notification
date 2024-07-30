package models

type NotificationRequest struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Text             string `json:"text"`
	OrganizationID   string `json:"organization_id"`
	UserID           string `json:"user_id"`
	Type             string `json:"type"`
	ReadStatus       int64  `json:"read_status"`
	CreatedAt        string `json:"created_at"`
	SendTo           string `json:"send_to"`
	NotificationType string `json:"notification_type"`
	SkipDB           *bool  `json:"skip_db"`
}
