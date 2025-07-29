/*
	Meander
	A portable Fountain utility for production writing
	Copyright (C) 2022-2023 Harley Denham
*/

package main

import "strings"
import "encoding/json"

func command_merge(config *Config) {
	merged_file, success := merge(config.source_file)
	if !success {
		eprintln("failed to merge file", config.source_file)
		return
	}

	success = write_file(fix_path(config.output_file), []byte(merged_file))
	if !success {
		eprintln("failed to write", config.output_file)
	}
}

/*func command_archive(config *Config) {
	text, success := merge(config.source_file)
	if !success {
		eprintln("failed to merge file", config.source_file)
		return
	}

	const NL rune = '\n'

	content := new(strings.Builder)
	content.Grow(len(text))

	last_relevant_rune := 'x' // this initial value is never relevant
	for {
		c, w := get_rune(content)
		content = content[w:]

		if c == '#' {
			if last_relevant_rune != '\n' {

			}
		}

		buffer.WriteRune(c)
		last_rune = c
	}

	success = write_file(fix_path(config.output_file), []byte(content.String()))
	if !success {
		eprintln("failed to write", config.output_file)
	}
}*/

func merge(source_file string) (string, bool) {
	text, success := load_file_normalise(fix_path(source_file))
	if !success {
		eprintf("%q not found", source_file)
		return "", false
	}

	content := new(strings.Builder)
	content.Grow(len(text))

	for {
		if len(text) == 0 {
			break
		}

		if text[0] == '\n' && len(text) > 9 && text[1] == 'i' {
			content.WriteRune('\n')
			text = text[1:] // newline

			if rune_on_line(text, ':') != 8 || homogenise(text[:7]) != "include" {
				continue
			}

			test_text := extract_to_newline(text)
			file_name := include_path(source_file, strings.TrimSpace(test_text[8:]))

			if child_text, found_file := merge(file_name); found_file {
				content.Grow(len(child_text))
				content.WriteString(consume_title_page(child_text))
			} else {
				content.WriteString(text[:len(test_text)])
			}

			text = text[len(test_text):]
			continue
		}

		the_rune, rune_width := get_rune(text)
		text = text[rune_width:]
		content.WriteRune(the_rune)
	}

	return content.String(), true
}

const DATA_VERSION = 1

func command_data(config *Config) {
	text, success := merge(config.source_file)
	if !success {
		return
	}

	data := init_data(config)
	syntax_parser(config, data, text)

	blob, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		eprintln("failed to marshal", config.output_file)
		return
	}

	success = write_file(config.output_file, blob)
	if !success {
		eprintln("failed to write", config.output_file)
	}
}
