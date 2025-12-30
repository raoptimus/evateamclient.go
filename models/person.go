package models

import "time"

// Person represents a person/user in EVA system
type Person struct {
	ID             string     `json:"id"`
	ClassName      string     `json:"class_name"`
	Name           string     `json:"name"`
	Code           string     `json:"code"` // login
	Login          string     `json:"login"`
	Email          *string    `json:"email,omitempty"`
	WorkPosition   *string    `json:"work_position,omitempty"`
	Phone          *string    `json:"phone,omitempty"`
	PhoneMobile    *string    `json:"phone_mobile,omitempty"`
	Telegram       *string    `json:"telegram,omitempty"`
	ProjectID      *string    `json:"project_id,omitempty"`
	OnVacation     *bool      `json:"on_vacation,omitempty"`
	DoesNotWork    *bool      `json:"does_not_work,omitempty"`
	CreatedAt      *time.Time `json:"cmf_created_at,omitempty"`
	AvatarFilename *string    `json:"avatar_filename,omitempty"`

	// ACL metadata
	//ACLFields map[string]string `json:"_acl_fields,omitempty"`
	//ACLObj    string            `json:"_acl_obj,omitempty"`
}

// PersonResponse for Person.get.
type PersonResponse struct {
	JSONRPC string `json:"jsonrpc,omitempty"`
	Result  Person `json:"result,omitempty"`
	Meta    Meta   `json:"meta,omitempty"`
}

// PersonListResponse for Person.list.
type PersonListResponse struct {
	JSONRPC string   `json:"jsonrpc,omitempty"`
	Result  []Person `json:"result,omitempty"`
	Meta    Meta     `json:"meta,omitempty"`
}
