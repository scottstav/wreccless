package logrender

import (
	"testing"
)

func TestParseAssistantText(t *testing.T) {
	line := `{"type":"assistant","content":"Hello world"}`
	events := ParseLine([]byte(line))
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != EventText {
		t.Errorf("expected EventText, got %v", events[0].Type)
	}
	if events[0].Text != "Hello world" {
		t.Errorf("expected 'Hello world', got %q", events[0].Text)
	}
}

func TestParseRealStreamJSON(t *testing.T) {
	line := `{"type":"assistant","message":{"content":[{"type":"text","text":"Real message."},{"type":"tool_use","name":"Bash","id":"x"}]}}`
	events := ParseLine([]byte(line))
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Type != EventText || events[0].Text != "Real message." {
		t.Errorf("event[0]: expected text 'Real message.', got %+v", events[0])
	}
	if events[1].Type != EventTool || events[1].ToolName != "Bash" {
		t.Errorf("event[1]: expected tool 'Bash', got %+v", events[1])
	}
}

func TestParseToolUse(t *testing.T) {
	line := `{"type":"tool_use","name":"Edit"}`
	events := ParseLine([]byte(line))
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != EventTool || events[0].ToolName != "Edit" {
		t.Errorf("expected tool 'Edit', got %+v", events[0])
	}
}

func TestParseResult(t *testing.T) {
	line := `{"type":"result","subtype":"success"}`
	events := ParseLine([]byte(line))
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != EventResult || events[0].SubType != "success" {
		t.Errorf("expected result 'success', got %+v", events[0])
	}
}

func TestParseSystem(t *testing.T) {
	line := `{"type":"system","subtype":"init"}`
	events := ParseLine([]byte(line))
	if len(events) != 0 {
		t.Errorf("expected 0 events for system, got %d", len(events))
	}
}

func TestParseEmptyContent(t *testing.T) {
	line := `{"type":"assistant","content":""}`
	events := ParseLine([]byte(line))
	if len(events) != 0 {
		t.Errorf("expected 0 events for empty content, got %d", len(events))
	}
}

func TestParseInvalidJSON(t *testing.T) {
	events := ParseLine([]byte("not json"))
	if len(events) != 1 || events[0].Type != EventText {
		t.Errorf("expected raw text event for invalid JSON, got %+v", events)
	}
}

func TestRenderPlain(t *testing.T) {
	events := []Event{
		{Type: EventText, Text: "hello"},
		{Type: EventTool, ToolName: "Edit"},
		{Type: EventResult, SubType: "success"},
	}
	out := RenderPlain(events)
	expected := "hello\n[tool: Edit]\n[result: success]\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}
