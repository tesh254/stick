package tui

import (
    "strings"
    "testing"

    "github.com/tesh254/stick/internal/functions"
)

func TestFunctionRenderer_ExecuteAndRender_Success(t *testing.T) {
    r := NewFunctionRenderer()

    reg := functions.NewRegistry()
    reg.Register("echo", functions.Echo, 0, -1)

    styled, err := r.ExecuteAndRender(reg, "echo", []string{"'hello'"}, nil)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !strings.Contains(styled, "Function Result") || !strings.Contains(styled, "hello") {
        t.Fatalf("styled output missing expected content: %q", styled)
    }
}

func TestFunctionRenderer_ExecuteAndRender_UnknownFunction(t *testing.T) {
    r := NewFunctionRenderer()
    reg := functions.NewRegistry()

    styled, err := r.ExecuteAndRender(reg, "unknown", []string{}, nil)
    if err == nil {
        t.Fatalf("expected error for unknown function, got none")
    }
    if !strings.Contains(err.Error(), "unknown function") {
        t.Fatalf("unexpected error message: %v", err)
    }
    if !strings.Contains(styled, "Function Error") {
        t.Fatalf("expected styled error output, got: %q", styled)
    }
}

func TestFunctionRenderer_ExecuteAndRender_CaseSensitiveCheck(t *testing.T) {
    r := NewFunctionRenderer()
    reg := functions.NewRegistry()
    reg.Register("echo", functions.Echo, 0, -1)

    // With case-sensitive (default), mixed case should be unknown
    _, err := r.ExecuteAndRender(reg, "Echo", []string{"'x'"}, nil)
    if err == nil {
        t.Fatalf("expected case-sensitive unknown function error")
    }

    // With case-insensitive, registry.Call will normalize; this should succeed
    styled, err := r.ExecuteAndRender(reg, "Echo", []string{"'x'"}, &CallOptions{CaseSensitive: false})
    if err != nil {
        t.Fatalf("unexpected error with case-insensitive option: %v", err)
    }
    if !strings.Contains(styled, "Function Result") || !strings.Contains(styled, "x") {
        t.Fatalf("styled output missing expected content: %q", styled)
    }
}