package models

type LogicType struct {
	ID           string            `json:"id"`
	ClassName    string            `json:"class_name"`
	Name         string            `json:"name"` // "Task"
	Code         string            `json:"code"` // "task.agile:task"
	CmfModelName string            `json:"cmf_model_name"`
	ParentID     *string           `json:"parent_id,omitempty"`
	ProjectID    *string           `json:"project_id,omitempty"`
	CmfOwnerID   string            `json:"cmf_owner_id"`
	ACLFields    map[string]string `json:"_acl_fields,omitempty"`
	ACLObj       string            `json:"_acl_obj,omitempty"`
}
