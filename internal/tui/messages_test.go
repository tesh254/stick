package tui

import (
    "strings"
    "testing"
)

// newTestModel creates a minimal model with a populated function registry
func newTestModel() model {
    m := model{}
    m.functionRegistry = setupFunctionRegistry()
    return m
}

func TestProcessFunctionCall_SingleWordNonMatching(t *testing.T) {
    m := newTestModel()
    result, err := m.processFunctionCall("hello")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != "" {
        t.Fatalf("expected empty result for non-matching single word, got %q", result)
    }
}

func TestProcessFunctionCall_SingleWordMatchingIsPlainText(t *testing.T) {
    m := newTestModel()
    result, err := m.processFunctionCall("echo")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != "" {
        t.Fatalf("expected plain text behavior for single word matching function, got %q", result)
    }
}

func TestProcessFunctionCall_CaseSensitivity(t *testing.T) {
    m := newTestModel()
    // Registered functions are lowercase; mixed case should be treated as plain text
    result, err := m.processFunctionCall("Echo")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != "" {
        t.Fatalf("expected plain text for case mismatch, got %q", result)
    }
}

func TestProcessFunctionCall_ValidEmptyCall(t *testing.T) {
    m := newTestModel()
    result, err := m.processFunctionCall("echo()")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !strings.Contains(result, "Function Result") {
        t.Fatalf("expected styled output to include 'Function Result', got %q", result)
    }
}

func TestProcessFunctionCall_ValidParameterizedCall(t *testing.T) {
    m := newTestModel()
    result, err := m.processFunctionCall("echo('text')")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !strings.Contains(result, "Function Result") || !strings.Contains(result, "text") {
        t.Fatalf("styled output missing expected content, got %q", result)
    }
}

func TestProcessFunctionCall_MalformedMissingClosingParen(t *testing.T) {
    m := newTestModel()
    _, err := m.processFunctionCall("echo(")
    if err == nil {
        t.Fatalf("expected error for missing closing parenthesis")
    }
    if !strings.Contains(err.Error(), "missing closing ')'") {
        t.Fatalf("expected missing closing ')' error, got: %v", err)
    }
}

func TestProcessFunctionCall_MultiWordPlainText(t *testing.T) {
    m := newTestModel()
    result, err := m.processFunctionCall("hello my name is erick")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != "" {
        t.Fatalf("expected empty result for multi-word plain text, got %q", result)
    }
}

func TestProcessFunctionCall_MultiWordWithFunctionCall(t *testing.T) {
    m := newTestModel()
    result, err := m.processFunctionCall("add(2, 2) please")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !strings.Contains(result, "Function Result") || !strings.Contains(result, "4") {
        t.Fatalf("styled output missing expected content, got %q", result)
    }
}

func TestProcessFunctionCall_SpaceBeforeParenIsPlainText(t *testing.T) {
    m := newTestModel()
    result, err := m.processFunctionCall("echo ('text')")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != "" {
        t.Fatalf("expected plain text (no function) when space precedes '(', got %q", result)
    }
}

func TestProcessFunctionCall_OpenParenTriggersDetection(t *testing.T) {
    m := newTestModel()
    _, err := m.processFunctionCall("echo(")
    if err == nil {
        t.Fatalf("expected error for missing closing parenthesis when '(' is present")
    }
    if !strings.Contains(err.Error(), "missing closing ')'") {
        t.Fatalf("expected missing closing ')' error, got: %v", err)
    }
}