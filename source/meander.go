/*
	Meander
	A portable Fountain utility for production writing
	Copyright (C) 2022-2023 Harley Denham
*/

package main

import "os"

import lib "github.com/signintech/gopdf"

const VERSION = "v0.3.0"
const MEANDER = "Meander " + VERSION

func main() {
	config, success := get_arguments()
	if !success {
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
		args := os.Args[2:]
		text := ""
		if len(args) == 0 {
			text = help("help")
		} else {
			text = help(args[0])
		}

		println(MEANDER)
		println(apply_color(text))
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

	scenes            uint8
	include_notes     bool
	include_synopses  bool
	include_sections  bool
	include_gender    bool
	table_of_contents bool

	template_set    bool
	template        Format
	template_string string

	paper_set  bool
	paper_size lib.Rect

	starred_show   bool
	starred_only   bool
	starred_target string

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

func get_arguments() (*Config, bool) {
	const SEE_HELP_RENDER = "see $1meander help render$0 for full usage"

	args := os.Args[1:]

	config := new(Config)

	max     := len(args) - 1
	index   := 0
	patharg := 0

	for {
		if index > max {
			break
		}

		arg := args[index]
		index += 1

		switch arg {
		case "render":
			config.command = COMMAND_RENDER
			continue

		case "merge":
			config.command = COMMAND_MERGE
			continue

		case "data":
			config.command = COMMAND_DATA
			continue

		case "gender":
			config.command = COMMAND_GENDER
			continue

		case "convert":
			config.command = COMMAND_CONVERT
			continue

		case "help":
			config.command = COMMAND_HELP
			return config, true

		case "version":
			config.command = COMMAND_VERSION
			return config, true

		case "credit":
			config.command = COMMAND_CREDIT
			return config, true

		case "fonts":
			config.command = COMMAND_FONTS
			return config, true
		}

		// there shouldn't be any arguments shorter than 2
		// that we aren't expecting as additional values
		if len(arg) < 2 {
			eprintf("error: unknown argument %q", arg)
			return config, false
		}

		if arg[:2] == "--" {
			arg = arg[2:]
		} else if arg[0] == '-' {
			arg = arg[1:]
		} else {
			switch patharg {
			case 0:
				config.source_file = arg
			case 1:
				config.output_file = arg
			default:
				eprintln("error: too many path arguments")
				return config, false
			}
			patharg += 1
			continue
		}

		switch arg {
		case "toc":
			config.table_of_contents = true

		case "notes":
			config.include_notes = true

		case "synopses":
			config.include_synopses = true

		case "sections":
			config.include_sections = true

		case "print-gender", "g":
			config.include_gender = true

		case "stars-only":
			config.starred_only = true
			fallthrough

		case "stars":
			config.starred_show = true

			if index > max {
				continue
			}
			if args[index][0] == '-' {
				continue
			}

			config.starred_target = args[index]
			index += 1

		case "scene", "s":
			if index > max {
				eprintln(apply_color("error: the --scene flag requires a value\n\n    input\n    remove\n    generate\n\n" + SEE_HELP_RENDER))
				return config, false
			}

			x, success := arg_scene_type(args[index])
			if !success {
				eprintln("invalid scene type")
				return config, false
			}
			config.scenes = x
			index += 1

		case "format", "f":
			if index > max {
				eprintln(apply_color("error: the --format flag requires a value\n\n    screenplay\n    stageplay\n    graphicnovel\n    manuscript\n\n" + SEE_HELP_RENDER))
				return config, false
			}

			x, success := set_format(args[index])
			if !success {
				eprintln("error: not a valid template")
				return config, false
			}

			config.template = x
			config.template_set = true
			config.template_string = args[index]
			index += 1

		case "paper", "p":
			if index > max {
				eprintln(apply_color("error: the --paper flag requires a value\n\n    USLetter\n    USLegal\n    A4\n\n" + SEE_HELP_RENDER))
				return config, false
			}

			x, success := set_paper(args[index])
			if !success {
				eprintln("error: invalid paper size")
			}

			config.paper_set = true
			config.paper_size = x
			index += 1

		default:
			eprintf("error: %q flag is unknown", arg)
			return config, false
		}
	}

	if config.source_file == "" {
		eprintln("error: no input file specified!")
		return config, false
	}

	if config.output_file == "" {
		switch config.command {
		case COMMAND_RENDER:
			config.output_file = rewrite_ext(config.source_file, ".pdf")
		case COMMAND_MERGE:
			config.output_file = rewrite_ext(config.source_file, "_merged" + FOUNTAIN_EXT)
		case COMMAND_CONVERT:
			config.output_file = rewrite_ext(config.source_file, FOUNTAIN_EXT)
		case COMMAND_DATA:
			config.output_file = rewrite_ext(config.source_file, ".json")
		}
	}

	return config, true
}
