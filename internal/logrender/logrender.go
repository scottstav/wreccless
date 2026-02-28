package logrender

import (
	"encoding/json"
	"fmt"
	"strings"
)

// EventType classifies a parsed log event.
type EventType int

const (
	EventText   EventType = iota // Plain text output from assistant
	EventTool                    // Tool invocation
	EventResult                  // Result marker
)

// Event is a single parsed log event.
type Event struct {
	Type     EventType
	Text     string // for EventText
	ToolName string // for EventTool
	SubType  string // for EventResult
}

// ParseLine parses an NDJSON log line into zero or more events.
// Handles both the real stream-json format (.message.content[]) and the
// simplified test format (.content string).
func ParseLine(line []byte) []Event {
	var raw map[string]interface{}
	if err := json.Unmarshal(line, &raw); err != nil {
		return []Event{{Type: EventText, Text: string(line)}}
	}

	typ, _ := raw["type"].(string)
	switch typ {
	case "assistant":
		content := raw["content"]
		if msg, ok := raw["message"].(map[string]interface{}); ok {
			content = msg["content"]
		}
		return parseContent(content)
	case "tool_use":
		name, _ := raw["name"].(string)
		return []Event{{Type: EventTool, ToolName: name}}
	case "result":
		sub, _ := raw["subtype"].(string)
		return []Event{{Type: EventResult, SubType: sub}}
	default:
		return nil
	}
}

func parseContent(content interface{}) []Event {
	switch c := content.(type) {
	case string:
		if c == "" {
			return nil
		}
		return []Event{{Type: EventText, Text: c}}
	case []interface{}:
		var events []Event
		for _, item := range c {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			itemType, _ := m["type"].(string)
			switch itemType {
			case "text":
				if text, ok := m["text"].(string); ok && text != "" {
					events = append(events, Event{Type: EventText, Text: text})
				}
			case "tool_use":
				if name, ok := m["name"].(string); ok {
					events = append(events, Event{Type: EventTool, ToolName: name})
				}
			}
		}
		return events
	}
	return nil
}

// RenderPlain renders events as plain text, matching the original CLI output format.
func RenderPlain(events []Event) string {
	var b strings.Builder
	for _, e := range events {
		switch e.Type {
		case EventText:
			fmt.Fprintf(&b, "%s\n", e.Text)
		case EventTool:
			fmt.Fprintf(&b, "[tool: %s]\n", e.ToolName)
		case EventResult:
			fmt.Fprintf(&b, "[result: %s]\n", e.SubType)
		}
	}
	return b.String()
}
