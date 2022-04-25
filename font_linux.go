package main

import (
	"os"
	"path/filepath"
)

const font_flag_supported = true

var system_dirs = []string {
	filepath.Join(os.Getenv("HOME"), ".fonts"),
	"/usr/share/fonts/truetype",
}