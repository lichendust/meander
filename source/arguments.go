/*
	Meander
	A portable Fountain utility for production writing
	Copyright (C) 2022-2023 Harley Denham

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"os"
	"strings"
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

// defaults
const (
	// always all lowercase
	default_template = "screenplay"
	default_paper    = "a4"

	fountain_extension = ".fountain"
	fountain_short_ext = ".ftn"
)

// config is the central location for all
// user input
type config struct {
	command uint8

	scenes     uint8
	template   string
	paper_size string

	include_notes        bool
	include_synopses     bool
	include_sections     bool
	write_gender         bool
	include_gender       bool
	json_keep_formatting bool

	revision     bool
	revision_tag string

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

			case "data":
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

		if a == "" {
			patharg++
			continue
		}

		switch a {
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

		case "preserve-markdown":
			conf.json_keep_formatting = true // @docs
			continue

		case "notes":
			conf.include_notes = true
			continue

		case "synopses":
			conf.include_synopses = true
			continue

		case "sections":
			conf.include_sections = true
			continue

		case "update-gender", "u":
			conf.write_gender = true
			continue

		case "print-gender", "g":
			conf.include_gender = true
			continue

		case "scene", "s":
			if b != "" {
				if x, ok := arg_scene_type[b]; ok {
					conf.scenes = x
				} else {
					eprintf("invalid scene flag: %q", b)
				}
				counter ++
			} else {
				eprintln("args: missing scene mode")
				has_errors = true
			}
			continue

		case "format", "f":
			if b != "" {
				conf.template = strings.ToLower(b)
				counter ++

				if _, ok := template_store[b]; !ok {
					eprintf("args: %q not a template", b)
					has_errors = true
				}

			} else {
				eprintln("args: missing format")
				has_errors = true
			}
			continue

		case "paper", "p":
			if b != "" {
				b = strings.ToLower(b)

				conf.paper_size = b
				counter ++

				if _, ok := paper_store[b]; !ok {
					eprintf("args: %q is not a supported paper size", b)
					has_errors = true
				}
			} else {
				eprintln("args: missing paper size")
				has_errors = true
			}
			continue

		default:
			eprintf("args: %q flag is unknown", a)
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
			eprintln("args: too many path arguments")
			has_errors = true
		}

		patharg++
	}

	if conf.source_file == "" {
		if x := find_default_file(); x != "" {
			conf.source_file = x
		} else {
			eprintln("args: no input file specified or detected!")
			has_errors = true
		}
	}

	if conf.output_file == "" {
		switch conf.command {
		case COMMAND_RENDER:
			conf.output_file = rewrite_ext(conf.source_file, ".pdf")

		case COMMAND_MERGE:
			conf.output_file = rewrite_ext(conf.source_file, "_merged.fountain")

		case COMMAND_CONVERT:
			conf.output_file = rewrite_ext(conf.source_file, fountain_extension)

		case COMMAND_JSON:
			conf.output_file = rewrite_ext(conf.source_file, ".json")
		}
	}

	return conf, !has_errors
}