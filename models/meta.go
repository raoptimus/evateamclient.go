package models

import "encoding/json"

// CmfFieldMeta describes a single field metadata from meta.CmfProject.fields.
type CmfFieldMeta struct {
	TEXKOMGroupByAllow   bool             `json:"TEXKOM_group_by_allow,omitempty"`
	APIAllow             bool             `json:"api_allow,omitempty"`
	Auto                 bool             `json:"auto,omitempty"`
	Caption              string           `json:"caption,omitempty"`
	ClassFullName        string           `json:"class_full_name,omitempty"`
	ClassName            string           `json:"class_name,omitempty"`
	ColumnHistory        bool             `json:"column_history,omitempty"`
	Comment              string           `json:"comment,omitempty"`
	Custom               bool             `json:"custom,omitempty"`
	FullsearchIndex      bool             `json:"fullsearch_index,omitempty"`
	IndexUsing           []string         `json:"index_using,omitempty"`
	LogLevel             *int             `json:"log_level,omitempty"`
	MaxLength            *int             `json:"max_length,omitempty"`
	NoACL                *bool            `json:"no_acl,omitempty"`
	Nullable             *bool            `json:"nullable,omitempty"`
	OptionsListParams    *json.RawMessage `json:"options_list_params,omitempty"`
	OptionsListQueryAll  bool             `json:"options_list_query_all,omitempty"`
	Placeholder          string           `json:"placeholder,omitempty"`
	Readonly             *bool            `json:"readonly,omitempty"`
	UIFORMVisible        bool             `json:"ui_form_visible,omitempty"`
	Virtual              bool             `json:"virtual,omitempty"`
	VirtualCacheTimelife *int             `json:"virtual_cache_timelife,omitempty"`
	Visible              bool             `json:"visible,omitempty"`
	Widget               string           `json:"widget,omitempty"`
}

// CmfProjectClassMeta describes meta.CmfProject object.
type CmfProjectClassMeta struct {
	FieldsOrder             []string                `json:"fields_order,omitempty"`
	ClassName               string                  `json:"class_name,omitempty"`
	CmfVer                  bool                    `json:"cmf_ver,omitempty"`
	CodePrefix              string                  `json:"code_prefix,omitempty"`
	Custom                  bool                    `json:"custom,omitempty"`
	DefaultOptionsFilter    *json.RawMessage        `json:"default_options_filter,omitempty"`
	EnableDeletePerm        bool                    `json:"enable_delete_perm,omitempty"`
	EnableEditPerm          bool                    `json:"enable_edit_perm,omitempty"`
	FullSearch              bool                    `json:"full_search,omitempty"`
	FullSearchPreloadFields []string                `json:"full_search_preload_fields,omitempty"`
	Icon                    string                  `json:"icon,omitempty"`
	LogicalDelete           bool                    `json:"logical_delete,omitempty"`
	MenuTreeOrderno         int                     `json:"menu_tree_orderno,omitempty"`
	MenuTreeParentID        string                  `json:"menu_tree_parent_id,omitempty"`
	Ordering                []string                `json:"ordering,omitempty"`
	OrdernoPartitionBy      []string                `json:"orderno_partition_by,omitempty"`
	Readonly                bool                    `json:"readonly,omitempty"`
	SmartNotify             bool                    `json:"smart_notify,omitempty"`
	UIFORM                  json.RawMessage         `json:"ui_form,omitempty"`
	UIGroupFields           []json.RawMessage       `json:"ui_group_fields,omitempty"`
	UIName                  string                  `json:"ui_name,omitempty"`
	VerboseName             string                  `json:"verbose_name,omitempty"`
	VerboseNamePlural       string                  `json:"verbose_name_plural,omitempty"`
	Fields                  map[string]CmfFieldMeta `json:"fields,omitempty"`
	TEXKOMNoCache           bool                    `json:"TEXKOM_no_cache,omitempty"`
	ACLParentField          *string                 `json:"acl_parent_field,omitempty"`
	CacheClusterFields      []string                `json:"cache_cluster_fields,omitempty"`
	CacheInmemory           bool                    `json:"cache_inmemory,omitempty"`
}

// CmfMeta is the root meta object.
type CmfMeta struct {
	CmfProject CmfProjectClassMeta `json:"CmfProject,omitempty"`
}
