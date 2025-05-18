package caches

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/goodblaster/logos"
	"github.com/stretchr/testify/assert"
)

func TestComplexCommand(t *testing.T) {
	ctx := context.Background()
	cache := New()

	j := `
		{
			"a": "b",
			"num": 1,
			"arr1": [3,4,5,6],
			"job-1234": {
				"domains": {
					"domain-1": {
						"countdown": 0,
						"status": "busy"
					},
					"domain-2": {
						"countdown": 1,
						"status": "busy"
					}
				},	
				"status": "busy"
			}
		}
		`

	m := map[string]any{}
	if err := json.Unmarshal([]byte(j), &m); err != nil {
		logos.Fatal(err)
	}

	_ = cache.Create(ctx, m)

	res := COMMANDS(
		FOR(`${{job-1234/domains/*/countdown}}`,
			IF(`${{job-1234/domains/${{1}}/countdown}} == 0`,
				REPLACE(`job-1234/domains/${{1}}/status`, "complete"),
				INC(`job-1234/domains/${{1}}/countdown`, -1),
			),
		),
		// This second iteration would actually be handled by a cascading trigger.
		FOR(`${{job-1234/domains/*/countdown}}`,
			IF(`${{job-1234/domains/${{1}}/countdown}} == 0`,
				REPLACE(`job-1234/domains/${{1}}/status`, "complete"),
				INC(`job-1234/domains/${{1}}/countdown`, -1),
			),
		),
		RETURN(`current job status is ${{job-1234/status}}`),
	).Do(ctx, cache)

	if assert.NotNil(t, res) {
		lastValue := res.Values[len(res.Values)-1]
		assert.Contains(t, lastValue, "current job status is busy")
	}

	res = COMMANDS(
		IF(`all(${{job-1234/domains/*/status}} == "complete")`,
			REPLACE(`job-1234/status`, "complete"),
			NOOP(),
		),
		IF(`all(${{job-1234/status}} == "complete")`,
			RETURN("job is complete"),
			RETURN(`current job status is ${{job-1234/status}}`),
		),
	).Do(ctx, cache)

	// Last result value should sat "job is complete".
	if assert.NoError(t, res.Error) {
		lastValue := res.Values[len(res.Values)-1]
		assert.Contains(t, lastValue, "job is complete")
	}
}
