package tui

import (
    "testing"
)

// Test that user input and function calls produce consistent display/storage updates.
func TestDualSlice_UserAndFunctionCall(t *testing.T) {
    m := initialModel()

    // Simulate user input
    m.textarea.SetValue("hello world")
    m.handleEnterKey()

    if len(m.messages) == 0 || len(m.storageMessages) == 0 {
        t.Fatalf("expected non-empty slices after user input")
    }
    if len(m.messages) != len(m.storageMessages) {
        t.Fatalf("display and storage slices should match in count; got %d vs %d", len(m.messages), len(m.storageMessages))
    }

    // Simulate function call
    m.textarea.SetValue("add(2, 3)")
    m.handleEnterKey()

    if len(m.messages) != len(m.storageMessages) {
        t.Fatalf("after function, display/storage slices diverged: %d vs %d", len(m.messages), len(m.storageMessages))
    }
}