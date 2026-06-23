package main

import (
	"fmt"

	"noisecheck/internal/viewer"
)

type viewerOptions struct {
	addr     string
	showHelp bool
}

func parseViewerFlags(args []string) (viewerOptions, error) {
	a := newNcFlagSet("nc viewer")

	opts := viewerOptions{}
	a.StringVar(&opts.addr, "addr", "localhost:5483", "listen address")

	if err := a.Parse(args); err != nil {
		return opts, fmt.Errorf("parse flags: %w", err)
	}

	opts.showHelp = a.showHelp
	return opts, nil
}

func runViewer(args []string) error {
	opts, err := parseViewerFlags(args)
	if err != nil {
		return err
	}
	if opts.showHelp {
		printViewerUsage()
		return nil
	}

	fmt.Printf("NoiseCheck Viewer starting on http://%s\n", opts.addr)
	return viewer.StartServer(opts.addr)
}

func printViewerUsage() {
	fmt.Println(`Session history WebUI viewer.

Usage:
  nc viewer [flags]
  nc v [flags]              (alias)

Flags:
  --addr <address>           listen address (default: localhost:5483)

Examples:
  nc viewer                     # start on default port
  nc viewer --addr :3000        # bind to all interfaces on port 3000`)
}
