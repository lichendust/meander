package main

import (
	"os"
	"fmt"
	"strings"
	"strconv"
	"unicode"
	"unicode/utf8"
	"path/filepath"
)

var (
	counter_panel    int
	counter_figure   int
	counter_series   int
	counter_chapter  int
)

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
	path := fix_path(source_file)
	text := ""

	{
		ok := false

		if config.revision {
			text, ok = load_file_git_tag(path, config.revision_tag)
		} else {
			text, ok = load_file_normalise(path)
		}

		if !ok {
			fmt.Fprintf(os.Stderr, "%q not found\n", filepath.ToSlash(source_file))
			return "", false
		}
	}

	newline_count := 0

	buffer := strings.Builder {}
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

			text = text[n + 2:]

			if last_element_removable {
				newline_count++
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

				text = text[n + 2:]

				if last_element_removable {
					newline_count++
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

				if child_text, ok := syntax_preprocessor(path, config); ok {
					if newline_count < 2 {
						buffer.WriteRune('\n')
					}

					child_text = consume_title_page(child_text)

					if config.revision {
						switch child_text[0] {
						case '\\', '+', '-':
						default: child_text = " " + child_text
						}
					}

					buffer.WriteString(child_text)
				} else {
					buffer.WriteString(text[:n + 2])
				}

				text = text[n + 2:]
				newline_count = 0
				continue
			}

			if template, ok := macro(test_text, "timestamp"); ok {
				if template == "" {
					template = default_timestamp
				}

				buffer.WriteString(nsdate(template))
				text = text[n + 2:]
				continue
			}

			if x, ok := macro(test_text, "series"); ok {
				counter_series++

				if x != "" {
					i, err := strconv.Atoi(x)

					if err != nil {
						i = 1 // when in rome
					}

					counter_series = i
				}

				buffer.WriteString(strconv.Itoa(counter_series))
				text = text[n+2:]
				continue
			}

			if x, ok := macro(test_text, "chapter"); ok {
				counter_chapter++

				if x != "" {
					i, err := strconv.Atoi(x)

					if err != nil {
						i = 1 // when in rome
					}

					counter_chapter = i
				}

				buffer.WriteString(strconv.Itoa(counter_chapter))
				text = text[n+2:]
				continue
			}

			if x, ok := macro(test_text, "panel"); ok {
				counter_panel++

				if x != "" {
					i, err := strconv.Atoi(x)

					if err != nil {
						i = 1 // when in rome
					}

					counter_panel = i
				}

				buffer.WriteString(strconv.Itoa(counter_panel))
				text = text[n+2:]
				continue
			}

			if x, ok := macro(test_text, "figure"); ok {
				counter_figure++

				if x != "" {
					i, err := strconv.Atoi(x)

					if err != nil {
						i = 1 // when in rome
					}

					counter_figure = i
				}

				buffer.WriteString(strconv.Itoa(counter_figure))
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
				newline_count++
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
		default: text = " " + text
		}
	}

	return text, true
}