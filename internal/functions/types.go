package functions

type Functions struct {
}

type Func func(args []string) (string, error)

type Registry struct {
	funcs  map[string]Func
	minArg map[string]int
	maxArg map[string]int
}

// GetFunctions returns a map of all registered function names
func (r *Registry) GetFunctions() map[string]Func {
	return r.funcs
}

// GetMinArgs returns the minimum number of arguments for a function
func (r *Registry) GetMinArgs(name string) (int, bool) {
	min, exists := r.minArg[name]
	return min, exists
}

// GetMaxArgs returns the maximum number of arguments for a function
func (r *Registry) GetMaxArgs(name string) (int, bool) {
	max, exists := r.maxArg[name]
	return max, exists
}

// ParseResult contains the result of parsing a function call
type ParseResult struct {
	FunctionName string
	Arguments    []string
	HasFunction  bool
	OriginalText string
	Error        error
}

type Parser struct{}
