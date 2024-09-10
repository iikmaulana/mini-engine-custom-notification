package models

type CustomNotificationResult struct {
	Id                   string `json:"id" db:"id"`
	LinkImageWeb         string `json:"link_image" db:"link_image"`
	Title                string `json:"title" db:"title"`
	Description          string `json:"description" db:"description"`
	Category             string `json:"category" db:"category"`
	LinkWeb              string `json:"link_web" db:"link_web"`
	TypeNotification     string `json:"type_notification" db:"type_notification"`
	Frekuensi            string `json:"frekuensi" db:"frekuensi"`
	StartDate            string `json:"start_date" db:"start_date"`
	EndDate              string `json:"end_date" db:"end_date"`
	TimeCronjob          string `json:"time_cronjob" db:"time_cronjob"`
	PengirimanBerikutnya string `json:"pengiriman_berikutnya"`
	Status               string `json:"status"`
	CreatedAt            string `json:"created_at" db:"created_at"`
	UpdatedAt            string `json:"updated_at" db:"updated_at"`
}
