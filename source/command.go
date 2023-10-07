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

import "strings"
import "unicode/utf8"
import "encoding/json"

func command_merge(config *Config) {
	merged_file, ok := merge(config.source_file)
	if !ok {
		eprintln("failed to merge file", config.source_file)
		return
	}

	ok = write_file(fix_path(config.output_file), []byte(merged_file))
	if !ok {
		eprintln("failed to write", config.output_file)
	}
}

func merge(source_file string) (string, bool) {
	text, ok := load_file_normalise(fix_path(source_file))
	if !ok {
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

		the_rune, rune_width := utf8.DecodeRuneInString(text)
		text = text[rune_width:]
		content.WriteRune(the_rune)
	}

	return content.String(), true
}

const DATA_VERSION = 1

func command_data(config *Config) {
	text, ok := merge(config.source_file)
	if !ok {
		return
	}

	data := init_data()
	syntax_parser(config, data, text)

	blob, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		eprintln("failed to marshal", config.output_file)
		return
	}

	ok = write_file(config.output_file, blob)
	if !ok {
		eprintln("failed to write", config.output_file)
	}
}