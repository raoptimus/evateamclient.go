/**
 * This file is part of the raoptimus/evateamclient.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/evateamclient.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/evateamclient.go
 */

package evateamclient

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIntegration_Stats(t *testing.T) {
	c := newIntegrationClient(t)

	report, err := c.SprintExecutorsKPI(context.Background(), &SprintExecutorsKPIParams{
		ProjectCode: "epud",
		SprintCode:  "SPR-001838",
	})
	require.NoError(t, err)
	require.NotNil(t, report)
}
