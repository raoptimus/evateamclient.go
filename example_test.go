package evateamclient

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStats(t *testing.T) {
	c, err := NewClient(&Config{
		BaseURL:  os.Getenv("EVA_API_URL"),
		APIToken: os.Getenv("EVA_API_TOKEN"),
		Debug:    true,
		Timeout:  0,
	})
	require.NoError(t, err)
	defer c.Close()

	report, err := c.SprintExecutorsKPI(context.Background(), &SprintExecutorsKPIParams{
		ProjectCode: "epud",
		SprintCode:  "SPR-001838",
	})
	require.NoError(t, err)
	require.NotNil(t, report)
}
