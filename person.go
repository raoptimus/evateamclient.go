package evateamclient

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/raoptimus/evateamclient/models"
)

// Person field constants for type-safe queries
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

var (
	// DefaultPersonFields - standard projection for single person queries
	DefaultPersonFields = []string{
		PersonFieldID,
		PersonFieldName,
		PersonFieldCode,
		PersonFieldLogin,
		PersonFieldEmail,
		PersonFieldOnVacation,
		PersonFieldDoesNotWork,
		PersonFieldCreatedAt,
	}

	// DefaultPersonListFields - optimized for LIST queries (lighter payload)
	DefaultPersonListFields = []string{
		PersonFieldID,
		PersonFieldName,
		PersonFieldCode,
		PersonFieldLogin,
		PersonFieldEmail,
		PersonFieldOnVacation,
		PersonFieldDoesNotWork,
	}
)

// Person retrieves a single person by ID (backward compatible)
// Example:
//
//	person, meta, err := client.Person(ctx, "Person:uuid-here", nil)
func (c *Client) Person(
	ctx context.Context,
	userID string,
	fields []string,
) (*models.Person, *models.Meta, error) {
	qb := NewQueryBuilder().
		Select(fields...).
		From(EntityPerson).
		Where(sq.Eq{PersonFieldID: userID}).
		Limit(1)

	return c.PersonQuery(ctx, qb)
}

// PersonQuery executes query using REAL Squirrel API
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  Select("id", "name", "email").
//	  From(evateamclient.EntityPerson).
//	  Where(sq.Eq{"login": "john.doe"})
//	person, meta, err := client.PersonQuery(ctx, qb)
func (c *Client) PersonQuery(ctx context.Context, qb *QueryBuilder) (*models.Person, *models.Meta, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return nil, nil, err
	}

	// Apply default fields if none specified
	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultPersonFields
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfPerson.get",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.PersonResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return &resp.Result, &resp.Meta, nil
}

// PersonsList retrieves list using REAL Squirrel
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  Select("id", "name", "email").
//	  From(evateamclient.EntityPerson).
//	  Where(sq.Eq{"on_vacation": false}).
//	  Where(sq.Eq{"does_not_work": false}).
//	  OrderBy("name").
//	  Offset(0).Limit(100)
//	persons, meta, err := client.PersonsList(ctx, qb)
func (c *Client) PersonsList(
	ctx context.Context,
	qb *QueryBuilder,
) ([]models.Person, *models.Meta, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return nil, nil, err
	}

	// Apply default fields if none specified
	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultPersonListFields
	}

	method, err := qb.ToMethod(false)
	if err != nil {
		return nil, nil, err
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  method,
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp models.PersonListResponse
	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Result, &resp.Meta, nil
}

// PersonCount counts using REAL Squirrel
// Example:
//
//	qb := evateamclient.NewQueryBuilder().
//	  From(evateamclient.EntityPerson).
//	  Where(sq.Eq{"does_not_work": false})
//	count, err := client.PersonCount(ctx, qb)
func (c *Client) PersonCount(
	ctx context.Context,
	qb *QueryBuilder,
) (int, error) {
	kwargs, err := qb.ToKwargs()
	if err != nil {
		return 0, err
	}

	reqBody := &RPCRequest{
		JSONRPC: "2.2",
		Method:  "CmfPerson.count",
		CallID:  newCallID(),
		Kwargs:  kwargs,
	}

	var resp struct {
		JSONRPC string `json:"jsonrpc"`
		Result  int    `json:"result"`
	}

	if err := c.doRequest(ctx, reqBody, &resp); err != nil {
		return 0, err
	}

	return resp.Result, nil
}

// Backward compatible methods (using old API)

// Persons retrieves users with custom filters (backward compatible, deprecated)
// Recommended: use PersonsList with NewQueryBuilder() instead
func (c *Client) Persons(
	ctx context.Context,
	kwargs map[string]any,
) ([]models.Person, *models.Meta, error) {
	if len(kwargs) == 0 {
		kwargs = make(map[string]any)
	}

	if _, hasFields := kwargs["fields"]; !hasFields {
		kwargs["fields"] = DefaultPersonListFields
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
