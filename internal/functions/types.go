package functions

type Functions struct {
}

type Func func(args []string) (string, error)

type Registry struct {
	funcs  map[string]Func
	minArg map[string]int
	maxArg map[string]int
}

type Parser struct{}
