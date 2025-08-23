package commands

import (
	"testing"

	"github.com/ODudek/go-aris/resp"
)

func TestPingCommand(t *testing.T) {
	cmd := &PingCommand{}
	result, err := cmd.Execute()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	strVal, ok := result.(*resp.StrValue)
	if !ok {
		t.Errorf("Expected StrValue, got %T", result)
	}

	if strVal.GetString() != "PONG" {
		t.Errorf("Expected 'PONG', got '%s'", strVal.GetString())
	}
}
