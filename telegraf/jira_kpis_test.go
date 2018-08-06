package jira_kpis

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJiraDefaultsToLocal(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	m := &Jirakpis{
		Servers: []string{fmt.Sprintf("root@tcp(%s:8080)/", testutil.GetLocalHost())},
	}

	var acc testutil.Accumulator
	err := m.Gather(&acc)
	require.NoError(t, err)

	assert.True(t, acc.HasMeasurement("jira_kpis"))
}
