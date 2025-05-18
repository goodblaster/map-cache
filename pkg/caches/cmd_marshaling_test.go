package caches

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshaling(t *testing.T) {
	var input CommandEnvelope
	err := json.Unmarshal([]byte(bigCommandJson), &input)
	assert.NoError(t, err)

	b, err := json.MarshalIndent(&input, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
	input.Commands = nil
	err = json.Unmarshal(b, &input)
	assert.NoError(t, err)
}

var simpleNoopJson = `{"commands": [{"type": "NOOP"}]}`
var simplePrintJson = `{"commands": [{"type": "PRINT", "messages": ["message 1"]}]}`
var simpleForJson = `{"commands": [{"type": "FOR", "loop_expr": "true", "commands": []}]}`

var simpleGetJson = `{"commands": [{"type": "GET", "keys": ["key 1"]}]}`
var simpleIfJson = `{"commands": [{"type": "IF", "condition": "true", "if_true": {"type":"NOOP"}, "if_false": {"type":"NOOP"}}]}`
var simpleIncJson = `{"commands": [{"type": "INC", "key": "key 1", "value": 1}]}`
var simpleReplaceJson = `{"commands": [{"type": "REPLACE", "key": "key 1", "value": 1}]}`
var simpleReturnJson = `{"commands": [{"type": "RETURN", "values": ["value 1"]}]}`

var bigCommandJson = `
{
  "commands": [
	{
	  "type": "FOR",
	  "loop_expr": "${{job-1234/domains/*/countdown}}",
	  "commands": [
		{
		  "type": "IF",
		  "condition": "${{job-1234/domains/${{1}}/countdown}} == 0",
		  "if_true": 
			{
			  "type": "REPLACE",
			  "key": "job-1234/domains/${{1}}/status",
			  "value": "complete"
			}
		  ,
		  "if_false": 
			{
			  "type": "INC",
			  "key": "job-1234/domains/${{1}}/countdown",
			  "value": -1
			}
		  
		}
	  ]
	},
	{
	  "type": "FOR",
	  "loop_expr": "${{job-1234/domains/*/countdown}}",
	  "commands": [
		{
		  "type": "IF",
		  "condition": "${{job-1234/domains/${{1}}/countdown}} == 0",
		  "if_true": 
			{
			  "type": "REPLACE",
			  "path": "job-1234/domains/${{1}}/status",
			  "value": "complete"
			}
		  ,
		  "if_false": 
			{
			  "type": "INC",
			  "path": "job-1234/domains/${{1}}/countdown",
			  "value": -1
			}
		  
		}
	  ]
	},
	{
	  "type": "RETURN",
	  "value": "current job status is ${{job-1234/status}}"
	},
	{
	  "type": "IF",
	  "condition": "all(${{job-1234/domains/*/status}} == \"complete\")",
	  "if_true": 
		{
		  "type": "REPLACE",
		  "path": "job-1234/status",
		  "value": "complete"
		}
	  ,
	  "if_false": 
		{
		  "type": "NOOP"
		}
	  
	},
	{
	  "type": "IF",
	  "condition": "all(${{job-1234/status}} == \"complete\")",
	  "if_true": 
		{
		  "type": "RETURN",
		  "value": "job is complete"
		}
	  ,
	  "if_false": 
		{
		  "type": "RETURN",
		  "value": "current job status is ${{job-1234/status}}"
		}
	  
	}
  ]
}
`
