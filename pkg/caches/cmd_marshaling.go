package caches

import (
	"encoding/json"
	"fmt"
)

type CommandEnvelope struct {
	Stuff    any          `json:"stuff,omitempty"` // Arbitrary metadata
	Commands []RawCommand `json:"commands"`
}

// RawCommand handles polymorphic decoding of any command by "type"
type RawCommand struct {
	Command Command
}

func (r RawCommand) MarshalJSON() ([]byte, error) {
	if r.Command == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(r.Command)
}

func (r *RawCommand) UnmarshalJSON(data []byte) error {
	var base struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &base); err != nil {
		return err
	}

	var cmd Command
	switch base.Type {
	case "IF":
		cmd = &CommandIf{}
	case "FOR":
		cmd = &CommandFor{}
	case "REPLACE":
		cmd = &CommandReplace{}
	case "RETURN":
		cmd = &CommandReturn{}
	case "PRINT":
		cmd = &CommandPrint{}
	case "GET":
		cmd = &CommandGet{}
	case "INC":
		cmd = &CommandInc{}
	case "NOOP":
		cmd = &CommandNoop{}
	case "COMMANDS":
		cmd = &CommandGroup{}
	default:
		return fmt.Errorf("unknown command type: %s", base.Type)
	}

	if err := json.Unmarshal(data, cmd); err != nil {
		return err
	}

	r.Command = cmd
	return nil
}

// TODO: This is all very tedious. Find a better way.
//
///
////
/////

func (c *CommandIf) MarshalJSON() ([]byte, error) {
	type Alias CommandIf
	return json.Marshal(struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  c.Type(),
		Alias: (*Alias)(c),
	})
}

func (c *CommandFor) MarshalJSON() ([]byte, error) {
	type Alias CommandFor
	return json.Marshal(struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  c.Type(),
		Alias: (*Alias)(c),
	})
}

func (c *CommandReplace) MarshalJSON() ([]byte, error) {
	type Alias CommandReplace
	return json.Marshal(struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  c.Type(),
		Alias: (*Alias)(c),
	})
}

func (c *CommandReturn) MarshalJSON() ([]byte, error) {
	type Alias CommandReturn
	return json.Marshal(struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  c.Type(),
		Alias: (*Alias)(c),
	})
}

func (c *CommandPrint) MarshalJSON() ([]byte, error) {
	type Alias CommandPrint
	return json.Marshal(struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  c.Type(),
		Alias: (*Alias)(c),
	})
}

func (c *CommandGet) MarshalJSON() ([]byte, error) {
	type Alias CommandGet
	return json.Marshal(struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  c.Type(),
		Alias: (*Alias)(c),
	})
}

func (c *CommandInc) MarshalJSON() ([]byte, error) {
	type Alias CommandInc
	return json.Marshal(struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  c.Type(),
		Alias: (*Alias)(c),
	})
}

func (c *CommandNoop) MarshalJSON() ([]byte, error) {
	type Alias CommandNoop
	return json.Marshal(struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  c.Type(),
		Alias: (*Alias)(c),
	})
}

func (c *CommandGroup) MarshalJSON() ([]byte, error) {
	type Alias CommandGroup
	return json.Marshal(struct {
		Type string `json:"type"`
		*Alias
	}{
		Type:  c.Type(),
		Alias: (*Alias)(c),
	})
}
