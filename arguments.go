package main

import (
	"os"
	"fmt"
	"runtime"
	"strings"
	"path/filepath"
)

const (
	COMMAND_RENDER uint8 = iota
	COMMAND_MERGE
	COMMAND_GENDER
	COMMAND_JSON
	COMMAND_CONVERT
	COMMAND_HELP
	COMMAND_VERSION
)

// defines how to handle scenes
const (
	SCENE_INPUT uint8 = iota    // use text input
	SCENE_REMOVE                // no scene numbers
	SCENE_GENERATE              // create new numbers
)

// config is the central location for all
// user input
type config struct {
	command    uint8

	scenes      uint8
	template    string
	paper_size  string

	include_notes    bool
	include_synopses bool
	include_sections bool
	write_gender     bool

	revision     bool
	revision_tag string

	font_name   string
	source_file string
	output_file string
}

var arg_scene_type = map[string]uint8 {
	"generate": SCENE_GENERATE,
	"remove":   SCENE_REMOVE,
	"input":    SCENE_INPUT,
}

// extracts arguments in the array as
// either --bool or --name <data>
func pull_argument(args []string) (string, string) {
	if len(args) == 0 {
		return "", ""
	}

	if len(args[0]) >= 1 {
		n := count_rune(args[0], '-')
		a := args[0]

		if n > 0 {
			a = a[n:]
		} else {
			return "", ""
		}

		if len(args[1:]) >= 1 {
			b := args[1]

			if len(b) > 0 && b[0] != '-' {
				return a, b
			}
		}

		return a, ""
	}

	return "", ""
}

// process the user input
func get_arguments() (*config, bool) {
	args := os.Args[1:]

	counter := 0
	patharg := 0

	has_errors := false

	conf := &config {}

	for {
		args = args[counter:]

		if len(args) == 0 {
			break
		}

		counter = 0

		if len(args) > 0 {
			switch args[0] {
			case "render":
				conf.command = COMMAND_RENDER
				args = args[1:]
				continue

			case "merge":
				conf.command = COMMAND_MERGE
				args = args[1:]
				continue

			case "json":
				conf.command = COMMAND_JSON
				args = args[1:]
				continue

			case "gender":
				conf.command = COMMAND_GENDER
				args = args[1:]
				continue

			case "convert":
				conf.command = COMMAND_CONVERT
				args = args[1:]
				continue

			case "help":
				conf.command = COMMAND_HELP
				return conf, true // exit immediately

			case "version":
				conf.command = COMMAND_VERSION
				return conf, true // exit immediately
			}
		}

		a, b := pull_argument(args[counter:])

		counter ++

		switch a {
		case "":
			break

		case "revision", "r":
			conf.revision = true

			if b != "" {
				conf.revision_tag = b
				counter ++
			}
			continue

		case "version":
			conf.command = COMMAND_VERSION
			return conf, true

		case "help", "h":
			// psychological failsafe —
			// the user is most likely
			// to try "--help" or "-h" first
			conf.command = COMMAND_HELP
			return conf, true

		case "notes":
			conf.include_notes = true
			continue

		case "synopses":
			conf.include_synopses = true
			continue

		case "sections":
			conf.include_sections = true
			continue

		case "update-table":
			conf.write_gender = true
			continue

		case "scene", "s":
			if b != "" {
				if x, ok := arg_scene_type[b]; ok {
					conf.scenes = x
				} else {
					fmt.Fprintf(os.Stderr, "invalid scene flag: %q\n", b)
				}
				counter ++
			} else {
				fmt.Fprintln(os.Stderr, "args: missing scene mode")
				has_errors = true
			}
			continue

		case "format", "f":
			if b != "" {
				conf.template = strings.ToLower(b)
				counter ++

				if _, ok := template_store[b]; !ok {
					fmt.Fprintf(os.Stderr, "args: %q not a template\n", b)
					has_errors = true
				}

			} else {
				fmt.Fprintln(os.Stderr, "args: missing format")
				has_errors = true
			}
			continue

		case "paper", "p":
			if b != "" {
				b = strings.ToLower(b)

				conf.paper_size = b
				counter ++

				if _, ok := paper_store[b]; !ok {
					fmt.Fprintf(os.Stderr, "args: %q is not a supported paper size\n", b)
					has_errors = true
				}
			} else {
				fmt.Fprintln(os.Stderr, "args: missing paper size")
				has_errors = true
			}
			continue

		case "font":
			if !font_flag_supported {
				fmt.Fprintf(os.Stderr, "args: --font searching not supported on %s\n", strings.Title(runtime.GOOS))
				has_errors = true
			}

			if b != "" {
				conf.font_name = b
				counter ++
			} else {
				fmt.Fprintln(os.Stderr, "args: missing font selection")
				has_errors = true
			}
			continue

		default:
			fmt.Fprintf(os.Stderr, "args: %q flag is unknown\n", a)
			has_errors = true

			if b != "" {
				counter ++
			}
		}

		switch patharg {
		case 0:
			conf.source_file = args[0]
		case 1:
			conf.output_file = args[0]
		default:
			fmt.Fprintln(os.Stderr, "args: too many path arguments")
			has_errors = true
		}

		patharg++
	}

	if conf.source_file == "" {
		if x := find_default_file(); x != "" {
			conf.source_file = x
		} else {
			fmt.Fprintln(os.Stderr, "args: no input file specified or detected!")
			has_errors = true
		}
	}

	if conf.output_file == "" {
		ext := filepath.Ext(conf.source_file)
		raw := conf.source_file[:len(conf.source_file) - len(ext)]

		switch conf.command {
			case COMMAND_RENDER:
				conf.output_file = raw + ".pdf"

			case COMMAND_MERGE:
				conf.output_file = raw + "_merged" + ext

			case COMMAND_CONVERT:
				conf.output_file = raw + ".fountain"

			case COMMAND_JSON:
				conf.output_file = raw + ".json"
		}
	}

	return conf, !has_errors
}