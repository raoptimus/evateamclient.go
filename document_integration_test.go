package evateamclient

import (
	"context"
	"os"
	"strings"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const integrationProjectCode = "epud"

func newIntegrationClient(t *testing.T) *Client {
	t.Helper()

	c, err := NewClient(&Config{
		BaseURL:  os.Getenv("EVA_API_URL"),
		APIToken: os.Getenv("EVA_API_TOKEN"),
		Debug:    true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { c.Close() })

	return c
}

func getIntegrationProjectID(t *testing.T, c *Client) string {
	t.Helper()

	project, _, err := c.Project(context.Background(), integrationProjectCode, DefaultProjectFields)
	require.NoError(t, err)
	require.NotNil(t, project)
	require.NotEmpty(t, project.ID)

	return project.ID
}

func TestIntegration_DocumentsList_AllFields(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(AllBasicAndRelationFields...).
		From(EntityDocument).
		Where(sq.Eq{DocumentFieldProjectID: projectID}).
		OrderBy("-" + DocumentFieldCmfCreatedAt).
		Limit(5)

	docs, meta, err := c.DocumentsList(ctx, qb)
	require.NoError(t, err)
	require.NotNil(t, meta)
	require.NotEmpty(t, docs, "project %s should have documents", integrationProjectCode)

	for _, doc := range docs {
		assert.True(t, strings.HasPrefix(doc.ID, "CmfDocument:"), "ID should start with CmfDocument:, got %s", doc.ID)
		assert.NotEmpty(t, doc.Name)
		assert.Equal(t, projectID, doc.ProjectID)
		assert.False(t, doc.CmfCreatedAt.IsZero(), "CmfCreatedAt should not be zero")
	}
}

func TestIntegration_DocumentsList_DefaultFields(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(DefaultDocumentListFields...).
		From(EntityDocument).
		Where(sq.Eq{DocumentFieldProjectID: projectID}).
		Limit(5)

	docs, meta, err := c.DocumentsList(ctx, qb)
	require.NoError(t, err)
	require.NotNil(t, meta)
	require.NotEmpty(t, docs)

	for _, doc := range docs {
		assert.True(t, strings.HasPrefix(doc.ID, "CmfDocument:"))
		assert.NotEmpty(t, doc.Code)
		assert.NotEmpty(t, doc.Name)
		assert.Equal(t, projectID, doc.ProjectID)
		assert.False(t, doc.CmfCreatedAt.IsZero(), "CmfCreatedAt should be parsed")
	}
}

func TestIntegration_Document_ByCode(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	// Get a document code from the list first
	qb := NewQueryBuilder().
		Select(DocumentFieldID, DocumentFieldCode).
		From(EntityDocument).
		Where(sq.Eq{DocumentFieldProjectID: projectID}).
		Limit(1)

	docs, _, err := c.DocumentsList(ctx, qb)
	require.NoError(t, err)
	require.NotEmpty(t, docs, "need at least one document")

	docCode := docs[0].Code
	require.NotEmpty(t, docCode)

	// Fetch single document by code with default fields
	doc, meta, err := c.Document(ctx, docCode, DefaultDocumentFields)
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.NotNil(t, meta)

	assert.Equal(t, docCode, doc.Code)
	assert.NotEmpty(t, doc.ID)
	assert.NotEmpty(t, doc.Name)
	assert.False(t, doc.CmfCreatedAt.IsZero(), "CmfCreatedAt should be parsed as time.Time")
	assert.False(t, doc.CmfModifiedAt.IsZero(), "CmfModifiedAt should be parsed as time.Time")
}

func TestIntegration_Document_AllRelationFields(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	// Get a document code
	qb := NewQueryBuilder().
		Select(DocumentFieldCode).
		From(EntityDocument).
		Where(sq.Eq{DocumentFieldProjectID: projectID}).
		Limit(1)

	docs, _, err := c.DocumentsList(ctx, qb)
	require.NoError(t, err)
	require.NotEmpty(t, docs)

	// Fetch with ** (all basic + relation fields)
	doc, meta, err := c.Document(ctx, docs[0].Code, AllBasicAndRelationFields)
	require.NoError(t, err, "deserialization with ** fields should not fail")
	require.NotNil(t, doc)
	require.NotNil(t, meta)

	assert.NotEmpty(t, doc.ID)
	assert.NotEmpty(t, doc.Name)
}

func TestIntegration_DocumentQuery_WithBuilder(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(
			DocumentFieldID,
			DocumentFieldCode,
			DocumentFieldName,
			DocumentFieldText,
			DocumentFieldProjectID,
			DocumentFieldCmfCreatedAt,
			DocumentFieldCmfModifiedAt,
		).
		From(EntityDocument).
		Where(sq.Eq{DocumentFieldProjectID: projectID}).
		Limit(1)

	doc, meta, err := c.DocumentQuery(ctx, qb)
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.NotNil(t, meta)

	assert.True(t, strings.HasPrefix(doc.ID, "CmfDocument:"))
	assert.NotEmpty(t, doc.Code)
	assert.NotEmpty(t, doc.Name)
	assert.Equal(t, projectID, doc.ProjectID)
	assert.False(t, doc.CmfCreatedAt.IsZero())
	assert.False(t, doc.CmfModifiedAt.IsZero())
}

func TestIntegration_DocumentCount(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	qb := NewQueryBuilder().
		Select(DocumentFieldID).
		From(EntityDocument).
		Where(sq.Eq{DocumentFieldProjectID: projectID})

	count, err := c.DocumentCount(ctx, qb)
	require.NoError(t, err)
	assert.Greater(t, count, 0, "project %s should have documents", integrationProjectCode)
}

func TestIntegration_ProjectDocuments(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	docs, meta, err := c.ProjectDocuments(ctx, projectID, DefaultDocumentListFields)
	require.NoError(t, err)
	require.NotNil(t, meta)
	require.NotEmpty(t, docs)

	for _, doc := range docs {
		assert.True(t, strings.HasPrefix(doc.ID, "CmfDocument:"))
		assert.Equal(t, projectID, doc.ProjectID)
	}
}

func TestIntegration_ProjectDocuments_AllFields(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	docs, meta, err := c.ProjectDocuments(ctx, projectID, AllBasicAndRelationFields)
	require.NoError(t, err, "deserialization with ** fields should not fail")
	require.NotNil(t, meta)
	require.NotEmpty(t, docs)

	// Validate time fields are properly parsed
	for _, doc := range docs {
		assert.NotEmpty(t, doc.ID)
		assert.Equal(t, projectID, doc.ProjectID)
		assert.False(t, doc.CmfCreatedAt.IsZero(), "CmfCreatedAt should be parsed")
	}
}

func TestIntegration_DocumentsList_ChildDocuments(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	// Find documents with a parent (child documents)
	qb := NewQueryBuilder().
		Select(
			DocumentFieldID,
			DocumentFieldCode,
			DocumentFieldName,
			DocumentFieldParentID,
			DocumentFieldProjectID,
		).
		From(EntityDocument).
		Where(sq.Eq{DocumentFieldProjectID: projectID}).
		Limit(200)

	allDocs, _, err := c.DocumentsList(ctx, qb)
	require.NoError(t, err)
	require.NotEmpty(t, allDocs)

	// Find a document that has a document parent (not project parent)
	var childDoc *struct {
		code     string
		parentID string
	}
	for _, doc := range allDocs {
		if doc.ParentID != "" && strings.HasPrefix(doc.ParentID, "CmfDocument:") {
			childDoc = &struct {
				code     string
				parentID string
			}{code: doc.Code, parentID: doc.ParentID}
			break
		}
	}

	if childDoc == nil {
		t.Skip("no child documents found in project, skipping parent-child test")
	}

	// Verify parent document exists
	parentQB := NewQueryBuilder().
		Select(DocumentFieldID, DocumentFieldCode, DocumentFieldName).
		From(EntityDocument).
		Where(sq.Eq{DocumentFieldID: childDoc.parentID}).
		Limit(1)

	parentDoc, _, err := c.DocumentQuery(ctx, parentQB)
	require.NoError(t, err, "parent document should exist")
	require.NotNil(t, parentDoc)
	assert.Equal(t, childDoc.parentID, parentDoc.ID)

	// List all children of the parent
	childrenQB := NewQueryBuilder().
		Select(DocumentFieldID, DocumentFieldCode, DocumentFieldName, DocumentFieldParentID).
		From(EntityDocument).
		Where(sq.Eq{DocumentFieldParentID: childDoc.parentID}).
		Limit(50)

	children, _, err := c.DocumentsList(ctx, childrenQB)
	require.NoError(t, err)
	require.NotEmpty(t, children, "parent should have at least one child")

	for _, child := range children {
		assert.Equal(t, childDoc.parentID, child.ParentID,
			"all children should reference the same parent")
	}

	// Verify our original child doc is in the list
	found := false
	for _, child := range children {
		if child.Code == childDoc.code {
			found = true
			break
		}
	}
	assert.True(t, found, "original child document should be in the children list")
}

func TestIntegration_Documents_Deprecated(t *testing.T) {
	c := newIntegrationClient(t)
	projectID := getIntegrationProjectID(t, c)
	ctx := context.Background()

	kwargs := map[string]any{
		"filter": [][]any{
			{DocumentFieldProjectID, "==", projectID},
		},
		"slice": []int{0, 5},
	}

	docs, meta, err := c.Documents(ctx, kwargs)
	require.NoError(t, err)
	require.NotNil(t, meta)
	require.NotEmpty(t, docs)

	for _, doc := range docs {
		assert.True(t, strings.HasPrefix(doc.ID, "CmfDocument:"))
	}
}
