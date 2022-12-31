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
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
	"path/filepath"
)

type prepro_data struct {
	counter_panel   int
	counter_figure  int
	counter_series  int
	counter_chapter int
}

func macro(text, keyword string) (string, bool) {
	if strings.HasPrefix(strings.ToLower(text), keyword) {
		for _, c := range text {
			if c == '\n' {
				return "", false
			}
		}

		text = strings.TrimSpace(text[len(keyword):])

		if len(text) == 0 {
			return "", true
		}

		if text[0] == ':' {
			text = strings.TrimSpace(text[1:])
		} else {
			return "", true
		}

		return strings.TrimSpace(text), true
	}

	return "", false
}

func syntax_preprocessor(source_file string, config *config) (string, bool) {
	data := prepro_data{}
	return prepro_recurse(source_file, config, &data)
}

func prepro_recurse(source_file string, config *config, data *prepro_data) (string, bool) {
	path := fix_path(source_file)
	text := ""

	{
		ok := false

		if config.revision {
			text, ok = load_file_tag(path, config.revision_tag, config.revision_system)
		} else {
			text, ok = load_file_normalise(path)
		}

		if !ok {
			eprintf("%q not found", filepath.ToSlash(source_file))
			return "", false
		}
	}

	newline_count := 0

	buffer := strings.Builder{}
	buffer.Grow(len(text))

	last_element_removable := false
	is_escaped := false

	for len(text) > 0 {
		// strip boneyards
		if text[0] == '\\' {
			if is_escaped {
				buffer.WriteRune('\\')
				text = text[1:]
				is_escaped = false
				continue
			}

			text = text[1:]
			is_escaped = true
			continue
		}

		if text[0] == '/' && len(text) > 1 && text[1] == '*' {
			n := rune_pair(text[2:], '*', '/')

			if n < 0 {
				buffer.WriteString(text[:2])
				text = text[2:]
				continue
			}

			text = text[n+2:]

			if last_element_removable {
				newline_count += 1
			}

			if newline_count > 0 {
				for len(text) > 0 {
					the_rune, rune_width := utf8.DecodeRuneInString(text)

					if !unicode.IsSpace(the_rune) {
						break
					}

					if the_rune == '\n' {
						text = text[rune_width:]

						write_diff := true

						if newline_count > 0 {
							write_diff = false
							newline_count--
						} else {
							buffer.WriteRune(the_rune)
						}

						if config.revision {
							the_rune, rune_width := utf8.DecodeRuneInString(text)
							text = text[rune_width:]

							if write_diff {
								buffer.WriteRune(the_rune)
							}
						}
						continue
					}

					buffer.WriteRune(the_rune)
					text = text[rune_width:]
				}
			}

			last_element_removable = true
			continue
		}

		// strip notes
		if !config.include_notes {
			if text[0] == '[' && len(text) > 1 && text[1] == '[' {
				if is_escaped {
					buffer.WriteString("[[")
					text = text[2:]
					is_escaped = false
					continue
				}

				n := rune_pair(text[2:], ']', ']')

				if n < 0 {
					buffer.WriteString(text[:2])
					text = text[2:]
					continue
				}

				text = text[n+2:]

				if last_element_removable {
					newline_count += 1
				}

				if newline_count > 0 {
					for len(text) > 0 {
						the_rune, rune_width := utf8.DecodeRuneInString(text)

						if !unicode.IsSpace(the_rune) {
							break
						}

						if the_rune == '\n' {
							text = text[rune_width:]

							write_diff := true

							if newline_count > 0 {
								write_diff = false
								newline_count--
							} else {
								buffer.WriteRune(the_rune)
							}

							if config.revision {
								the_rune, rune_width := utf8.DecodeRuneInString(text)
								text = text[rune_width:]

								if write_diff {
									buffer.WriteRune(the_rune)
								}
							}
							continue
						}

						buffer.WriteRune(the_rune)
						text = text[rune_width:]
					}
				}

				last_element_removable = true
				continue
			}
		}

		// handle directives
		if text[0] == '{' && len(text) > 1 && text[1] == '{' {
			if is_escaped {
				buffer.WriteString("{{")
				text = text[2:]
				is_escaped = false
				continue
			}

			n := rune_pair(text[2:], '}', '}')

			if n < 0 {
				buffer.WriteString(text[:2])
				text = text[2:]
				continue
			}

			test_text := text[2:n]

			if path, ok := macro(test_text, "include"); ok {
				path = include_path(source_file, path)

				if child_text, ok := prepro_recurse(path, config, data); ok {
					if newline_count < 2 {
						buffer.WriteRune('\n')
					}

					child_text = consume_title_page(child_text)

					if config.revision {
						switch child_text[0] {
						case '\\', '+', '-':
						default:
							child_text = " " + child_text
						}
					}

					buffer.WriteString(child_text)
				} else {
					buffer.WriteString(text[:n+2])
				}

				text = text[n+2:]
				newline_count = 0
				continue
			}

			if template, ok := macro(test_text, "timestamp"); ok {
				if template == "" {
					template = default_timestamp
				}

				buffer.WriteString(nsdate(template))
				text = text[n+2:]
				continue
			}

			if x, ok := macro(test_text, "series"); ok {
				data.counter_series += 1

				if x != "" {
					i, err := strconv.Atoi(x)
					if err != nil {
						i = 1 // when in rome
					}

					data.counter_series = i
				}

				buffer.WriteString(strconv.Itoa(data.counter_series))
				text = text[n+2:]
				continue
			}

			if x, ok := macro(test_text, "chapter"); ok {
				data.counter_chapter += 1

				if x != "" {
					i, err := strconv.Atoi(x)
					if err != nil {
						i = 1 // when in rome
					}

					data.counter_chapter = i
				}

				buffer.WriteString(strconv.Itoa(data.counter_chapter))
				text = text[n+2:]
				continue
			}

			if x, ok := macro(test_text, "panel"); ok {
				data.counter_panel += 1

				if x != "" {
					i, err := strconv.Atoi(x)
					if err != nil {
						i = 1 // when in rome
					}

					data.counter_panel = i
				}

				buffer.WriteString(strconv.Itoa(data.counter_panel))
				text = text[n+2:]
				continue
			}

			if x, ok := macro(test_text, "figure"); ok {
				data.counter_figure += 1

				if x != "" {
					i, err := strconv.Atoi(x)
					if err != nil {
						i = 1 // when in rome
					}

					data.counter_figure = i
				}

				buffer.WriteString(strconv.Itoa(data.counter_figure))
				text = text[n+2:]
				continue
			}
		}

		if is_escaped {
			buffer.WriteRune('\\')
			is_escaped = false
		}

		last_element_removable = false

		the_rune, rune_width := utf8.DecodeRuneInString(text)

		text = text[rune_width:]

		if unicode.IsSpace(the_rune) {
			if the_rune == '\n' {
				newline_count += 1
			}

			if config.revision {
				buffer.WriteRune(the_rune)

				the_rune, rune_width = utf8.DecodeRuneInString(text)
				text = text[rune_width:]
			}
		} else {
			newline_count = 0
		}

		buffer.WriteRune(the_rune)
	}

	text = strings.TrimSpace(buffer.String())

	if config.revision {
		switch text[0] {
		case '\\', '+', '-':
		default:
			text = " " + text
		}
	}

	return text, true
}