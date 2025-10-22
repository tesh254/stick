package functions

import (
	"fmt"
	"strings"
)

func NewRegistry() *Registry {
	return &Registry{
		funcs:  make(map[string]Func),
		minArg: make(map[string]int),
		maxArg: make(map[string]int),
	}
}

func (r *Registry) Register(name string, fn Func, minArgs, maxArgs int) {
	key := strings.ToLower(name)
	r.funcs[key] = fn

	if minArgs > 0 {
		r.minArg[key] = minArgs
	}
	if maxArgs >= 0 {
		r.maxArg[key] = maxArgs
	}
}

func (r *Registry) Call(name string, args []string) (string, error) {
	key := strings.ToLower(name)
	fn, ok := r.funcs[key]

	if !ok {
		return "", fmt.Errorf("unknown function: %s", name)
	}
	if min, ok := r.minArg[key]; ok && len(args) < min {
		return "", fmt.Errorf("%s expects at least %d args, got %d", name, min, len(args))
	}
	if max, ok := r.maxArg[key]; ok && max >= 0 && len(args) > max {
		return "", fmt.Errorf("%s expects at most %d args, got %d", name, max, len(args))
	}
	return fn(args)
}
