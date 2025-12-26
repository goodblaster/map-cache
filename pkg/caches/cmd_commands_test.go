package caches

import (
	"context"
	"encoding/json"
	"testing"

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
		t.Fatal(err)
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
		assert.Contains(t, res.Value, "current job status is busy")
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
		assert.Contains(t, res.Value, "job is complete")
	}
}

func Test_MultipleGet(t *testing.T) {
	ctx := context.Background()
	cache := New()

	// Create some test data
	testData := map[string]any{
		"key1":           "value1",
		"key2":           "value2",
		"key3/innerKey1": "innerValue1",
		"key4/1":         "item2",
	}

	err := cache.Create(ctx, testData)
	if !assert.NoError(t, err, "Failed to create test data in cache") {
		return
	}

	res := COMMANDS(
		GET("key1"),
		GET("key2"),
		GET("key3/innerKey1"),
		GET("key4/1"),
	).Do(ctx, cache)

	if assert.Len(t, res.Value, 4, "Expected 4 items") {
		vals := res.Value.([]any)
		assert.EqualValues(t, "value1", vals[0], "Expected value1 for key1")
		assert.EqualValues(t, "value2", vals[1], "Expected value2 for key2")
		assert.EqualValues(t, "innerValue1", vals[2], "Expected innerValue1 for key3/innerKey1")
		assert.EqualValues(t, "item2", vals[3], "Expected item2 for key4/1")
	}
}
