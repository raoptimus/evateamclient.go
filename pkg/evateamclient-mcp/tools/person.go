package tools

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/raoptimus/evateamclient.go"
)

// PersonTools provides MCP tool handlers for person operations.
type PersonTools struct {
	client *evateamclient.Client
}

// NewPersonTools creates a new PersonTools instance.
func NewPersonTools(client *evateamclient.Client) *PersonTools {
	return &PersonTools{client: client}
}

// PersonListInput represents input for eva_person_list tool.
type PersonListInput struct {
	QueryInput
	OnVacation  *bool `json:"on_vacation,omitempty"`
	DoesNotWork *bool `json:"does_not_work,omitempty"`
}

// PersonList returns a list of persons.
func (p *PersonTools) PersonList(ctx context.Context, input *PersonListInput) (*ListResult, error) {
	qb, err := BuildQuery(evateamclient.EntityPerson, &input.QueryInput)
	if err != nil {
		return nil, WrapError("person_list", err)
	}

	if input.OnVacation != nil {
		qb = qb.Where(sq.Eq{"on_vacation": *input.OnVacation})
	}
	if input.DoesNotWork != nil {
		qb = qb.Where(sq.Eq{"does_not_work": *input.DoesNotWork})
	}

	persons, _, err := p.client.PersonsList(ctx, qb)
	if err != nil {
		return nil, WrapError("person_list", err)
	}

	return &ListResult{
		Items:   toAnySlice(persons),
		HasMore: len(persons) == input.Limit && input.Limit > 0,
	}, nil
}

// PersonGetInput represents input for eva_person_get tool.
type PersonGetInput struct {
	ID     string   `json:"id,omitempty"`
	Login  string   `json:"login,omitempty"`
	Email  string   `json:"email,omitempty"`
	Fields []string `json:"fields,omitempty"`
}

// PersonGet retrieves a single person.
func (p *PersonTools) PersonGet(ctx context.Context, input PersonGetInput) (any, error) {
	var qb *evateamclient.QueryBuilder

	switch {
	case input.ID != "":
		qb = evateamclient.NewQueryBuilder().
			Select(input.Fields...).
			From(evateamclient.EntityPerson).
			Where(sq.Eq{"id": input.ID}).
			Limit(1)
	case input.Login != "":
		qb = evateamclient.NewQueryBuilder().
			Select(input.Fields...).
			From(evateamclient.EntityPerson).
			Where(sq.Eq{"login": input.Login}).
			Limit(1)
	case input.Email != "":
		qb = evateamclient.NewQueryBuilder().
			Select(input.Fields...).
			From(evateamclient.EntityPerson).
			Where(sq.Eq{"email": input.Email}).
			Limit(1)
	default:
		return nil, WrapError("person_get", ErrInvalidInput)
	}

	person, _, err := p.client.PersonQuery(ctx, qb)
	if err != nil {
		return nil, WrapError("person_get", err)
	}

	return person, nil
}

// PersonCountInput represents input for eva_person_count tool.
type PersonCountInput struct {
	OnVacation  *bool `json:"on_vacation,omitempty"`
	DoesNotWork *bool `json:"does_not_work,omitempty"`
}

// PersonCount counts persons.
func (p *PersonTools) PersonCount(ctx context.Context, input PersonCountInput) (*CountResult, error) {
	qb := evateamclient.NewQueryBuilder().From(evateamclient.EntityPerson)

	if input.OnVacation != nil {
		qb = qb.Where(sq.Eq{"on_vacation": *input.OnVacation})
	}
	if input.DoesNotWork != nil {
		qb = qb.Where(sq.Eq{"does_not_work": *input.DoesNotWork})
	}

	count, err := p.client.PersonCount(ctx, qb)
	if err != nil {
		return nil, WrapError("person_count", err)
	}

	return &CountResult{Count: count}, nil
}
