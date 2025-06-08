package caches

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Big(t *testing.T) {
	t.Skip("Skipping big test for performance reasons")

	ctx := context.Background()
	cache := New()
	err := cache.Create(ctx, map[string]any{
		"job-1234": map[string]any{
			"domains": map[string]any{},
			"status":  "running",
		},
	})
	assert.NoError(t, err)

	const N = 100 // number of domains
	const M = 100 // countdowns per domain

	for i := 1; i <= N; i++ {
		domain := map[string]any{
			"status":    "running",
			"countdown": M,
		}
		err = cache.Create(ctx, map[string]any{
			fmt.Sprintf("job-1234/domains/domain-%d", i): domain,
		})
		assert.NoError(t, err)
	}

	// Insert the trigger commands
	var cmd RawCommand
	err = json.Unmarshal([]byte(countdownTriggerCommand), &cmd)
	assert.NoError(t, err)

	_, err = cache.CreateTrigger(ctx, "job-1234/domains/*/countdown", cmd.Command)
	assert.NoError(t, err)

	err = json.Unmarshal([]byte(statusTriggerCommand), &cmd)
	_, err = cache.CreateTrigger(ctx, "job-1234/domains/*/status", cmd.Command)
	assert.NoError(t, err)

	start := time.Now()
	wg := sync.WaitGroup{}
	wg.Add(N)
	for i := 1; i <= N; i++ {
		go func(i int) {
			defer wg.Done()
			id := fmt.Sprintf("%d", i)
			for j := 0; j < M; j++ {
				// Print every 10th iteration
				//if j%10 == 0 {
				//	fmt.Print(i, j, " \n")
				//}
				cmd := COMMANDS(
					INC(fmt.Sprintf("job-1234/domains/domain-%d/countdown", i), -1),
					RETURN("${{job-1234/status}}"),
				)

				cache.Acquire(id)
				res := cmd.Do(ctx, cache)
				if res.Error != nil {
					fmt.Printf("Error processing command for domain-%d: %v\n", i, res.Error)
					cache.Release(id)
					return
				}
				cache.Release(id)
				status, ok := res.Value.(string)
				if ok && status == "complete" {
					fmt.Printf("completed by thread %d \n", i)
				}
			}
		}(i)
	}
	wg.Wait()
	fmt.Printf("Processed %d countdowns in %s\n", M, time.Since(start))

	status, err := cache.Get(ctx, "job-1234/status")
	assert.NoError(t, err)
	assert.Equal(t, "complete", status)
}

const countdownTriggerCommand = `{
  "type": "IF",
  "condition": "${{job-1234/domains/${{1}}/countdown}} <= 0",
  "if_true": {
	"type": "REPLACE",
	"key": "job-1234/domains/${{1}}/status",
	"value": "complete"
  },
  "if_false": {
	"type": "NOOP"
  }
}`

const statusTriggerCommand = `{
  "type": "IF",
  "condition": "all(${{job-1234/domains/*/status}} == \"complete\")",
  "if_true": {
	"type": "REPLACE",
	"key": "job-1234/status",
	"value": "complete"
  },
  "if_false": {
	"type": "NOOP"
  }
}`
