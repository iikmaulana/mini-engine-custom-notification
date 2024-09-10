package models

type PromoResult struct {
	ID                   string `json:"id" db:"id"`
	LinkImage            string `json:"link_image" db:"link_image"`
	Title                string `json:"title" db:"title"`
	Description          string `json:"description" db:"description"`
	TypePromo            string `json:"type_promo" db:"type_promo"`
	TargetPromo          string `json:"target_promo" db:"target_promo"`
	StartDate            string `json:"start_date" db:"start_date"`
	EndDate              string `json:"end_date" db:"end_date"`
	StatusPopup          string `json:"status_popup" db:"status_popup"`
	StatusNotif          string `json:"status_notif" db:"status_notif"`
	Frekuensi            string `json:"frekuensi" db:"frekuensi"`
	TimeCronjob          string `json:"time_cronjob" db:"time_cronjob"`
	CreatedAt            string `json:"created_at" db:"created_at"`
	UpdatedAt            string `json:"updated_at" db:"updated_at"`
	PengirimanBerikutnya string `json:"updated_at"`
	Status               string `json:"status"`
}
