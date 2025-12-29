package parser

import (
	"os"
	"path/filepath"
)

// ParseFileOptions configures file parsing behavior
type ParseFileOptions struct {
	// Defines are symbols to define for conditional assembly (.ifdef/.ifndef)
	Defines []string
	// EnablePreprocessor enables .include and conditional directives (default: true)
	EnablePreprocessor bool
}

// DefaultParseFileOptions returns the default options for parsing
func DefaultParseFileOptions() ParseFileOptions {
	return ParseFileOptions{
		EnablePreprocessor: true,
	}
}

// ParseFile reads and parses an assembly file with preprocessing support.
// This is the recommended entry point for parsing files, handling:
// - File reading
// - Preprocessing (.include, .ifdef, .ifndef, .else, .endif)
// - Parsing
//
// Returns the parsed program or an error. Check parser.Errors() for additional warnings.
func ParseFile(filePath string, opts ParseFileOptions) (*Program, *Parser, error) {
	// Read the file
	content, err := os.ReadFile(filePath) // #nosec G304 -- user-provided assembly file path
	if err != nil {
		return nil, nil, err
	}

	filename := filepath.Base(filePath)
	source := string(content)

	// Apply preprocessing if enabled
	if opts.EnablePreprocessor {
		baseDir := filepath.Dir(filePath)
		pp := NewPreprocessor(baseDir)

		// Apply defines
		for _, def := range opts.Defines {
			pp.Define(def)
		}

		// Process content (handles .include, .ifdef, etc.)
		processed, err := pp.ProcessContent(source, filename)
		if err != nil {
			return nil, nil, err
		}

		// Check for preprocessor errors
		if len(pp.Errors().Errors) > 0 {
			// Preprocessor errors are fatal
			return nil, nil, pp.Errors().Errors[0]
		}

		source = processed
	}

	// Parse the (possibly preprocessed) source
	p := NewParser(source, filename)
	program, err := p.Parse()
	if err != nil {
		return nil, p, err
	}

	return program, p, nil
}

// ParseFileSimple is a convenience wrapper that uses default options
func ParseFileSimple(filePath string) (*Program, *Parser, error) {
	return ParseFile(filePath, DefaultParseFileOptions())
}
