package main

import (
	"embed"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/lookbusy1344/arm-emulator/parser"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Parse command line flags
	flag.Parse()

	// Create application
	app := NewApp()

	// Load initial file if specified (with preprocessing support)
	if flag.NArg() > 0 {
		filePath := flag.Arg(0)
		// #nosec G304 -- filePath comes from command-line argument, user-controlled by design
		source, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to read file %s: %v", filePath, err)
		}

		// Run preprocessor to handle .include, .ifdef, etc.
		baseDir := filepath.Dir(filePath)
		pp := parser.NewPreprocessor(baseDir)
		processedSource, err := pp.ProcessContent(string(source), filepath.Base(filePath))
		if err != nil {
			log.Fatalf("Preprocessing error in %s: %v", filePath, err)
		}
		if len(pp.Errors().Errors) > 0 {
			log.Fatalf("Preprocessing error: %v", pp.Errors().Errors[0])
		}

		if err := app.LoadProgramFromSource(processedSource, filePath, 0x8000); err != nil {
			log.Fatalf("Failed to load program: %v", err)
		}
	}

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "ARM Emulator",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
