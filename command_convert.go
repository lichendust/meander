package main

import (
	"os"
	"fmt"
	"path/filepath"
)

func command_convert(config *config) {
	ext := filepath.Ext(config.source_file)

	switch ext {
	case ".fdx":
		command_convert_final_draft(config)
	case ".highland":
		command_convert_highland(config)
	case ".ftn", ".fountain":
		fmt.Fprintf(os.Stderr, "convert: %q is already a Fountain file\n", config.source_file)
	default:
		fmt.Fprintf(os.Stderr, "convert: no handler for filetype %q\n", ext)
	}
}