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
	"strings"
	"unicode/utf8"
)

func command_merge_document(config *config) {
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
	text, ok := load_file(fix_path(source_file))
	if !ok {
		eprintf("%q not found", source_file)
		return "", false
	}

	content := strings.Builder{}
	content.Grow(len(text))

	for {
		if len(text) == 0 {
			break
		}

		if text[0] == '{' && len(text) > 1 && text[1] == '{' {
			n := rune_pair(text[2:], '}', '}')

			if n < 0 {
				content.WriteString(text[:2])
				text = text[2:]
				continue
			}

			test_text := text[2:n]

			if path, ok := macro(test_text, "include"); ok {
				path = include_path(source_file, path)

				if child_text, ok := merge(path); ok {
					content.Grow(len(child_text))
					content.WriteString(consume_title_page(child_text))
				} else {
					content.WriteString(text[:n + 2])
				}

				text = text[n + 2:]
				continue
			}
		}

		the_rune, rune_width := utf8.DecodeRuneInString(text)
		text = text[rune_width:]
		content.WriteRune(the_rune)
	}

	return content.String(), true
}