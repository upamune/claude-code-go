package claude

import (
	"encoding/json"
	"fmt"
)

// ParseMessage parses a JSON line from the Claude Code CLI into a Message
func ParseMessage(line string) (Message, error) {
	// First, parse to determine the message type
	var baseMsg struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal([]byte(line), &baseMsg); err != nil {
		return nil, &ParseError{
			Line:    line,
			Message: err.Error(),
		}
	}

	// Parse based on the message type
	switch baseMsg.Type {
	case "user":
		var msg UserMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			return nil, &ParseError{
				Line:    line,
				Message: fmt.Sprintf("failed to parse user message: %v", err),
			}
		}
		return &msg, nil

	case "assistant":
		var msg AssistantMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			return nil, &ParseError{
				Line:    line,
				Message: fmt.Sprintf("failed to parse assistant message: %v", err),
			}
		}
		return &msg, nil

	case "result":
		var msg ResultMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			return nil, &ParseError{
				Line:    line,
				Message: fmt.Sprintf("failed to parse result message: %v", err),
			}
		}
		return &msg, nil

	case "system":
		var msg SystemMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			return nil, &ParseError{
				Line:    line,
				Message: fmt.Sprintf("failed to parse system message: %v", err),
			}
		}
		return &msg, nil

	case "permission_request":
		var msg PermissionRequestMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			return nil, &ParseError{
				Line:    line,
				Message: fmt.Sprintf("failed to parse permission request message: %v", err),
			}
		}
		return &msg, nil

	default:
		return nil, &ParseError{
			Line:    line,
			Message: fmt.Sprintf("unknown message type: %s", baseMsg.Type),
		}
	}
}
