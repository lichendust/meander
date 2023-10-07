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

import "os"

const VERSION = "v0.2.0"
const MEANDER = "Meander " + VERSION

func main() {
	config, ok := get_arguments()
	if !ok {
		return
	}

	switch config.command {
	case COMMAND_RENDER:
		command_render(config)

	case COMMAND_MERGE:
		command_merge(config)

	case COMMAND_DATA:
		command_data(config)

	case COMMAND_GENDER:
		command_gender(config)

	case COMMAND_CONVERT:
		command_convert(config)

	case COMMAND_FONTS:
		export_fonts()

	case COMMAND_VERSION:
		println(MEANDER)

	case COMMAND_CREDIT:
		println(MEANDER)
		println(help("credit") + "\n" + LICENSE_TEXT)

	case COMMAND_HELP:
		println(MEANDER)

		args := os.Args[2:]
		if len(args) == 0 {
			println(apply_color(help("help")))
			return
		}

		println(apply_color(help(args[0])))
	}
}

const FOUNTAIN_EXT = ".fountain"

const (
	COMMAND_RENDER uint8 = iota
	COMMAND_MERGE
	COMMAND_GENDER
	COMMAND_DATA
	COMMAND_CONVERT
	COMMAND_HELP
	COMMAND_VERSION
	COMMAND_CREDIT
	COMMAND_FONTS
)

const (
	SCENE_INPUT uint8 = iota // use text input
	SCENE_REMOVE             // no scene numbers
	SCENE_GENERATE           // create new numbers
)

// config is the central location for all user input
type Config struct {
	command uint8

	scenes     uint8
	template   string
	paper_size string

	include_notes     bool
	include_synopses  bool
	include_sections  bool
	write_gender      bool
	include_gender    bool
	table_of_contents bool
	raw_convert       bool

	source_file string
	output_file string
}

func arg_scene_type(x string) (uint8, bool) {
	switch x {
	case "generate":
		return SCENE_GENERATE, true
	case "remove":
		return SCENE_REMOVE, true
	case "input":
		return SCENE_INPUT, true
	}
	return SCENE_INPUT, false
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
func get_arguments() (*Config, bool) {
	args := os.Args[1:]

	counter := 0
	patharg := 0

	has_errors := false

	conf := new(Config)

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
				conf.command = COMMAND_DATA
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

			case "credit":
				conf.command = COMMAND_CREDIT
				return conf, true // exit immediately

			case "fonts":
				conf.command = COMMAND_FONTS
				return conf, true // exit immediately
			}
		}

		a, b := pull_argument(args[counter:])

		counter += 1

		switch a {
		case "":
			// continue to below

		case "version":
			conf.command = COMMAND_VERSION
			return conf, true

		case "help", "h":
			// psychological failsafe —
			// the user is most likely
			// to try "--help" or "-h" first
			conf.command = COMMAND_HELP
			return conf, true

		case "notes", "n":
			conf.include_notes = true
			continue

		case "synopses":
			conf.include_synopses = true
			continue

		case "sections":
			conf.include_sections = true
			continue

		case "raw":
			conf.raw_convert = true
			continue

		case "update-gender", "u":
			conf.write_gender = true
			continue

		case "print-gender", "g":
			conf.include_gender = true
			continue

		case "scene", "s":
			if b != "" {
				if x, ok := arg_scene_type(b); ok {
					conf.scenes = x
				} else {
					eprintf("invalid scene flag: %q", b)
				}
				counter += 1
			} else {
				eprintln("args: missing scene mode")
				has_errors = true
			}
			continue

		case "format", "f":
			if b != "" {
				counter += 1

				if _, ok := is_valid_format(b); ok {
					conf.template = b
					continue
				}
			}

			eprintln("args: bad format")
			has_errors = true
			continue

		case "paper", "p":
			if b != "" {
				conf.paper_size = b
				counter += 1

				if set_paper(conf.paper_size) != nil {
					continue
				}
			}

			eprintln("args: bad paper size")
			has_errors = true
			continue

		default:
			eprintf("args: %q flag is unknown", a)
			has_errors = true

			if b != "" {
				counter += 1
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

		patharg += 1
	}

	if conf.source_file == "" {
		eprintln("args: no input file specified or detected!")
		has_errors = true
	}

	if conf.output_file == "" {
		switch conf.command {
		case COMMAND_RENDER:
			conf.output_file = rewrite_ext(conf.source_file, ".pdf")
		case COMMAND_MERGE:
			conf.output_file = rewrite_ext(conf.source_file, "_merged" + FOUNTAIN_EXT)
		case COMMAND_CONVERT:
			conf.output_file = rewrite_ext(conf.source_file, FOUNTAIN_EXT)
		case COMMAND_DATA:
			conf.output_file = rewrite_ext(conf.source_file, ".json")
		}
	}

	return conf, !has_errors
}