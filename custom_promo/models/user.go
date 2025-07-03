package models

type UserResult struct {
	OrganizationId string `json:"organization_id" db:"organization_id"`
	UserId         string `json:"user_id" db:"user_id"`
}
