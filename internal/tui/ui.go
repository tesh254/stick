package tui

import (
	"fmt"
	"strings"

	"github.com/tesh254/stick/internal/functions"
)

// CallOptions controls validation behavior for function calls.
type CallOptions struct {
	// CaseSensitive enforces exact match on function names before executing.
	// When true, the provided name must match a key in the registry exactly.
	// Defaults to true.
	CaseSensitive bool
}

// FunctionRenderer exposes a unified API to render function calls and results
// using the project's standardized styles.
type FunctionRenderer struct{}

// NewFunctionRenderer creates a renderer that uses the shared styles.
func NewFunctionRenderer() *FunctionRenderer {
	return &FunctionRenderer{}
}

// ExecuteAndRender validates, executes, and renders a function call.
// It returns a fully styled string representing either the result or the error.
//
// - registry: the function registry to execute against.
// - name: the function name as typed by the user.
// - args: the raw string arguments (unquoted processing is handled in functions).
// - opts: optional call options (defaults applied when nil).
func (r *FunctionRenderer) ExecuteAndRender(registry *functions.Registry, name string, args []string, opts *CallOptions) (string, error) {
	if registry == nil {
		return r.renderFunctionOrToolResult(name, strings.Join(args, ", "), "", true), fmt.Errorf("registry is nil")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return r.renderFunctionOrToolResult(name, strings.Join(args, ", "), "", true), fmt.Errorf("function name is empty")
	}

	if opts == nil {
		opts = &CallOptions{CaseSensitive: true}
	}

	// Validate existence with case sensitivity if requested.
	if opts.CaseSensitive {
		funcs := registry.GetFunctions()
		if _, ok := funcs[name]; !ok {
			// Keep error semantics consistent with the TUI layer expectations.
			return r.renderFunctionOrToolResult(name, strings.Join(args, ", "), "", true), fmt.Errorf("unknown function: %s", name)
		}
	}

	// Execute via the registry (which internally validates min/max args).
	result, err := registry.Call(name, args)
	if err != nil {
		// Render a styled error block and return the error.
		return r.renderFunctionOrToolResult(name, strings.Join(args, ", "), err.Error(), true), err
	}

	// Successfully executed; render the styled result.
	return r.renderFunctionOrToolResult(name, strings.Join(args, ", "), result, false), nil
}

// RenderFunctionName renders a styled header and the function name/args.
func (r *FunctionRenderer) RenderFunctionName(name string, args []string) string {
	return renderFunctionName(name, strings.Join(args, ", "))
}

// renderFunctionOrToolResult is a thin wrapper around the shared renderer.
func (r *FunctionRenderer) renderFunctionOrToolResult(name string, args string, result string, isError bool) string {
	return renderFunctionOrToolResult(name, args, result, isError)
}

// Documentation:
//
// Typical usage in the TUI when a function call is parsed:
//
//  renderer := NewFunctionRenderer()
//  styled, err := renderer.ExecuteAndRender(registry, functionName, arguments, nil)
//  if err != nil {
//      // Append styled to the viewport along with any user-prefix for errors
//  } else {
//      // Append styled to the viewport for successful result display
//  }
//
// This unified API ensures consistent styling and error handling for all
// function calls throughout the terminal UI.
