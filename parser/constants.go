package parser

// Literal Pool Estimation Constants
const (
	// EstimatedLiteralsPerPool is the heuristic estimate for the number of literal values
	// typically stored in each literal pool section. Used for sizing pool sections during assembly.
	EstimatedLiteralsPerPool = 16
)

// Macro Processing Constants
const (
	// MaxMacroNestingDepth is the maximum depth for nested macro expansions.
	// Prevents infinite recursion in macro processing.
	MaxMacroNestingDepth = 100
)
