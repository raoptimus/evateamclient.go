package evateamclient

import (
	"context"

	"github.com/raoptimus/evateamclient/models"
)

// ALL readable fields from ACL [response]
const (
	// Core identity fields (always readable)

	PersonFieldID    = "id"
	PersonFieldName  = "name"
	PersonFieldCode  = "code"
	PersonFieldLogin = "login"

	// Contact info (readonly)

	PersonFieldEmail          = "email"
	PersonFieldEmail2         = "email_2"
	PersonFieldPhone          = "phone"
	PersonFieldPhoneInternal  = "phone_internal"
	PersonFieldPhoneMobile    = "phone_mobile"
	PersonFieldPhone2         = "phone_2"
	PersonFieldPhoneAssistant = "phone_assistant"
	PersonFieldSkype          = "skype"
	PersonFieldTelegram       = "telegram"
	PersonFieldWhatsApp       = "whatsapp"
	PersonFieldSlack          = "slack"
	PersonFieldZoom           = "zoom"
	PersonFieldLinkedin       = "linkedin"
	PersonFieldFacebook       = "facebook"
	PersonFieldInstagram      = "instagram"
	PersonFieldVK             = "vk"
	PersonFieldOK             = "ok"

	// Personal info

	PersonFieldFirstName    = "first_name"
	PersonFieldLastName     = "last_name"
	PersonFieldSecondName   = "second_name"
	PersonFieldWorkPosition = "work_position"
	PersonFieldBirthday     = "birthday"

	// Status

	PersonFieldOnVacation     = "on_vacation"
	PersonFieldDoesNotWork    = "does_not_work"
	PersonFieldAvatarFilename = "avatar_filename"

	// Client info

	PersonFieldClientJobName    = "client_job_name"
	PersonFieldClientDepartment = "client_department"
	PersonFieldClientDivision   = "client_division"
	PersonFieldClientOffice     = "client_office"

	// System

	PersonFieldProjectID    = "project_id"
	PersonFieldParentID     = "parent_id"
	PersonFieldCmfOwnerID   = "cmf_owner_id"
	PersonFieldCreatedAt    = "cmf_created_at"
	PersonFieldModifiedAt   = "cmf_modified_at"
	PersonFieldDeletedLogin = "deleted_login"
	PersonFieldOldLogin     = "old_login"
	PersonFieldExtID        = "ext_id"
	PersonFieldOrderNo      = "orderno"

	// DENY fields (hidden)
	// activity_id, api_token_hash, auth_*, password_*, tokens, etc.
)

// DefaultPersonFields - Production ready (all readonly fields)
var DefaultPersonFields = []string{
	PersonFieldID,
	PersonFieldName,
	PersonFieldCode,
	PersonFieldLogin,
	PersonFieldEmail,
	PersonFieldOnVacation,
	PersonFieldDoesNotWork,
	PersonFieldCreatedAt,
}

// Person retrieves single user by ID or email/login.
func (c *Client) Person(
	ctx context.Context,
	userID string,
	fields []string,
) (*models.Person, *models.Meta, error) {
	if len(fields) == 0 {
		fields = DefaultPersonFields
	}

	kwargs := map[string]any{
		"filter": []any{"id", "==", userID},
		"fields": fields,
		"slice":  []int{0, 1},
	}

	users, meta, err := c.Persons(ctx, kwargs)
	if err != nil {
		return nil, nil, err
	}

	if len(users) == 0 {
		return nil, meta, nil
	}

	return &users[0], meta, nil
}

// Persons retrieves users with custom filters.
func (c *Client) Persons(
	ctx context.Context,
	kwargs map[string]any,
) ([]models.Person, *models.Meta, error) {
	if len(kwargs) == 0 {
		kwargs = make(map[string]any)
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfPerson.list",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.PersonListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}
