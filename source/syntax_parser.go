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
	"unicode"
	"unicode/utf8"
)

func get_last_node(nodes []*syntax_node) (*syntax_node, bool) {
	if len(nodes) > 0 {
		return nodes[len(nodes)-1], true
	}
	return nil, false
}

func is_character_train(node_type uint8) bool {
	switch node_type {
	case CHARACTER, PARENTHETICAL, DIALOGUE, LYRIC:
		return true
	}
	return false
}

func assign_dual_dialogue(original *syntax_node, nodes []*syntax_node) {
	if original.level == 0 {
		return
	}

	if len(nodes) > 0 {
		for i := len(nodes) - 1; i >= 0; i-- {
			target := nodes[i]

			if target.node_type == CHARACTER {
				if target.level == 1 {
					original.level = 0 // don't set because there's already one above
					break
				}
				target.level = 1
				break
			}

			if !(is_character_train(target.node_type) || target.node_type == WHITESPACE) {
				original.level = 0 // don't set because can't have other stuff between
				break
			}
		}
	}
}

func find_title_colon(input string) (int, bool) {
	for i, c := range input {
		if c == '\n' {
			return 0, false
		}
		if c == ':' {
			return i, true
		}
	}
	return 0, false
}

func syntax_parser(config *config) (*fountain_content, bool) {
	text, ok := syntax_preprocessor(config.source_file, config)

	if !ok {
		return nil, false
	}

	if text == "" {
		return nil, false
	}

	// estimate final token count for allocation
	token_count := 0

	for _, c := range text {
		if c == '\n' {
			token_count++
		}
	}

	// prepare the elements that make up the "content" object
	title := make(map[string]string, 16)
	nodes := make([]*syntax_node, 0, token_count)

	first := true

	// title page mini-parser
	for {
		n, ok := find_title_colon(text)

		// if there's no ":", we're done with the title page
		if !ok {
			break
		}

		// get the leading word before the ":"
		word := strings.ToLower(text[:n])

		// consume the word and colon
		text = text[n+1:]

		if config.revision && first {
			// because of the need to always skip the diff char
			// to check if we're at the end of the title page
			// below, we end up returning to this point for
			// each title entry with the diff char already
			// removed - we only need to do it the first time
			word = word[1:]
			first = false
		}

		word = strings.TrimSpace(word)

		// make a teeny buffer for reconstituting multi-line entries
		sub_line_buffer := strings.Builder{}
		sub_line_buffer.Grow(64)

		// that "loop:" syntax in Go is too
		// nuclear so we need a delayed
		// nested loop break
		break_main_loop := false

		// begin parsing
		for {
			// grab the first line manually
			line := extract_to_newline(text)
			text = text[len(line):] // consume the line

			sub_line_buffer.WriteString(strings.TrimSpace(line))

			// the first character is now the newline (or eof)
			// because "extract_to_newline" stops before the
			// first one it finds

			if len(text) == 0 {
				break
			}

			if text[0] == '\n' {
				text = text[1:] // consume the newline

				if len(text) == 0 {
					break
				}

				// eat the diff char
				if config.revision {
					text = text[1:]
				}

				// this means we've hit a double newline;
				// title pages in Fountain must be all
				// smushed together
				if len(text) > 0 && text[0] == '\n' {
					break_main_loop = true
					break
				}

				// we're now into multi-line entries

				// if there's no leading space this is not a
				// multi-line entry, so break to finish up this
				// entry and go to the next one
				if text[0] != ' ' {
					break
				}

				// we know there's a leading space, we're into
				// multi-line, so grab that line and consume
				// it
				sub_line := extract_to_newline(text)
				text = text[len(sub_line):]

				/*if config.revision {
					sub_line = sub_line[1:]
				}*/

				// write it into the sub_line buffer,
				// _removing_ that leading whitespace,
				// because we don't want it
				sub_line_buffer.WriteRune('\n')
				sub_line_buffer.WriteString(strings.TrimSpace(sub_line))
			}
		}

		// get the final version of the sub_line and make sure
		// it has no leading newline (side effect of the
		// WriteRune directly above) it's cheaper to remove
		// here than to make a logical exception within the loop
		sub_line := consume_whitespace(sub_line_buffer.String())

		// if the entire sub_line is empty, the user didn't fill
		// in the value so don't register it we're nice about
		// it though, we don't write it back into their
		// screenplay as an action
		if sub_line != "" {
			title[word] = sub_line
		}

		// if the inner loop detected the end
		// of the title page, we're done here
		if break_main_loop {
			break
		}
	}

	// only remove newlines in case the first
	// element something like indented "action".
	text = consume_newlines(text)

	for {
		// eof
		if len(consume_whitespace(text)) == 0 {
			break
		}

		// get the line
		line := extract_to_newline(text)
		text = text[len(line):] // consume it

		// remove the left-behind newline with eof safety
		if len(text) > 0 && text[0] == '\n' {
			text = text[1:]
		}

		is_revised := false
		dirty_line := line

		if config.revision {
			if len(dirty_line) > 0 {
				char := rune(dirty_line[0])
				chop := false

				switch char {
				case '+':
					chop = true
					is_revised = true
				case ' ':
					chop = true
				case '-', '\\':
					continue
				}

				if chop {
					dirty_line = dirty_line[1:]
				}
			}
		}

		clean_line := strings.TrimSpace(dirty_line)

		// just whitespace
		if clean_line == "" {
			// if the last line was whitespace, level the
			// existing one up instead of writing a new
			// one
			if last_node, ok := get_last_node(nodes); ok {
				if last_node.node_type == WHITESPACE {
					last_node.level++
					continue
				}
			}

			nodes = append(nodes, &syntax_node{
				node_type: WHITESPACE,
				level:     1,
			})
			continue
		}

		// single characters on lines
		// are _actions_, so don't check
		// forces if the line is only one char.
		if len(clean_line) > 1 {
			// handle "forced" syntaxes
			switch clean_line[0] {
			case '!':
				nodes = append(nodes, &syntax_node{
					node_type: ACTION,
					revised:   is_revised,
					raw_text:  clean_line[1:],
				})
				continue

			case '@':
				level := uint8(0)

				if clean_line[len(clean_line)-1] == '^' {
					clean_line = clean_line[:len(clean_line)-1]
					level++
				}

				the_node := &syntax_node{
					node_type: CHARACTER,
					level:     level,
					revised:   is_revised,
					raw_text:  consume_whitespace(clean_line[1:]),
				}

				assign_dual_dialogue(the_node, nodes)

				nodes = append(nodes, the_node)
				continue

			case '~':
				n := count_rune(clean_line, '~')

				if n == 2 {
					break
				}

				nodes = append(nodes, &syntax_node{
					node_type: LYRIC,
					revised:   is_revised,
					raw_text:  clean_line[1:],
				})
				continue

			case '=':
				n := count_rune(clean_line, '=')

				// "===" is a page-break
				if n >= 3 {
					nodes = append(nodes, &syntax_node{
						node_type: PAGE_BREAK,
					})
					continue
				}

				// otherwise "= words" is a synopsis
				nodes = append(nodes, &syntax_node{
					node_type: SYNOPSIS,
					raw_text:  consume_whitespace(clean_line[n:]),
				})
				continue

			// sections
			case '#':
				n := count_rune(clean_line, '#')

				x := uint8(n) - 1
				if x > 2 {
					x = 2
				} // clamp sections to 3

				nodes = append(nodes, &syntax_node{
					node_type: SECTION + x,
					level:     x,
					revised:   is_revised,
					raw_text:  consume_whitespace(clean_line[n:]),
				})
				continue

			// scenes
			case '.':
				// a leading stop "." forces a scene
				// however, a leading ellipsis "...",
				// should _not_ be considered a scene

				// so we need to check(ignoring spaces) for
				// repeating stops, for which two in
				// sequence are _enough_.
				if consume_whitespace(clean_line[1:])[0] == '.' {
					break
				}

				// this is safe because of the >1 check that
				// the entire switch is wrapped in
				clean_line = clean_line[1:]

				if scene, scene_number, ok := has_scene_number(clean_line); ok {
					// insert cleaned scene + number
					nodes = append(nodes, &syntax_node{
						node_type: SCENE,
						revised:   is_revised,
						raw_text:  scene,
					})
					nodes = append(nodes, &syntax_node{
						node_type: SCENE_NUMBER,
						raw_text:  scene_number,
					})
				} else {
					// insert just the whole line scene
					nodes = append(nodes, &syntax_node{
						node_type: SCENE,
						revised:   is_revised,
						raw_text:  consume_whitespace(clean_line),
					})
				}
				continue

			// parenthetical
			case '(':
				if clean_line[len(clean_line)-1] != ')' {
					break // action
				}

				if last_node, ok := get_last_node(nodes); ok {
					if !is_character_train(last_node.node_type) {
						break // action
					}
				}

				nodes = append(nodes, &syntax_node{
					node_type: PARENTHETICAL,
					revised:   is_revised,
					raw_text:  clean_line,
				})
				continue

			// transitions + centered
			case '>':
				clean_line = consume_whitespace(clean_line[1:])

				// if line ends with a matching angle bracket
				// we're a "centered" item
				if clean_line[len(clean_line)-1] == '<' {
					clean_line = consume_ending_whitespace(clean_line[:len(clean_line)-1])

					nodes = append(nodes, &syntax_node{
						node_type: CENTERED,
						revised:   is_revised,
						raw_text:  clean_line,
					})
					continue
				}

				nodes = append(nodes, &syntax_node{
					node_type: TRANSITION,
					revised:   is_revised,
					raw_text:  clean_line,
				})
				continue

			case '{':
				if len(line[1:]) > 0 && line[1:][0] == '{' {
					n := rune_pair(line[2:], '}', '}')

					if n < 0 {
						break
					}

					line = strings.TrimSpace(line[2:n])

					// do line-spacing clean-up that matches
					// preprocessor behaviour (different way of
					// handling the subtractive newlines on
					// notes and boneyards)
					if len(nodes) > 0 {
						last_node := nodes[len(nodes)-1]

						if last_node.node_type == WHITESPACE {
							for i, c := range text {
								if i == int(last_node.level) {
									break
								}

								if c != '\n' {
									break
								}

								text = text[1:]
							}
						}
					}

					if config.revision {
						line = line[1:]
					}

					if template, ok := macro(line, "header"); ok {
						nodes = append(nodes, &syntax_node{
							node_type: HEADER,
							raw_text:  template,
						})
						continue
					}

					if template, ok := macro(line, "footer"); ok {
						nodes = append(nodes, &syntax_node{
							node_type: FOOTER,
							raw_text:  template,
						})
						continue
					}

					if template, ok := macro(line, "pagenumber"); ok {
						nodes = append(nodes, &syntax_node{
							node_type: PAGE_NUMBER,
							raw_text:  template,
						})
						continue
					}

					// we're just removing these for now because
					// we don't handle them yet.
					line = strings.ToLower(line)

					if strings.HasPrefix(line, "toc") {
						continue
					}
					if strings.HasPrefix(line, "watermark") {
						continue
					}
					if strings.HasPrefix(line, "endnote") {
						continue
					}

					// @todo this isn't all of them
				}
			}
		}

		// scene headings
		if is_valid_scene(clean_line) {
			if scene, scene_number, ok := has_scene_number(clean_line); ok {
				// insert cleaned scene + number
				nodes = append(nodes, &syntax_node{
					node_type: SCENE,
					revised:   is_revised,
					raw_text:  scene,
				})
				nodes = append(nodes, &syntax_node{
					node_type: SCENE_NUMBER,
					raw_text:  scene_number,
				})
			} else {
				// insert just the whole line scene
				nodes = append(nodes, &syntax_node{
					node_type: SCENE,
					revised:   is_revised,
					raw_text:  clean_line,
				})
			}
			continue
		}

		// transitions
		if is_valid_transition(clean_line) {
			nodes = append(nodes, &syntax_node{
				node_type: TRANSITION,
				revised:   is_revised,
				raw_text:  clean_line,
			})
			continue
		}

		// characters
		if is_valid_character(clean_line) {
			level := uint8(0)

			if clean_line[len(clean_line)-1] == '^' {
				clean_line = consume_ending_whitespace(clean_line[:len(clean_line)-1])
				level++
			}

			the_node := &syntax_node{
				node_type: CHARACTER,
				level:     level,
				revised:   is_revised,
				raw_text:  clean_line,
			}

			assign_dual_dialogue(the_node, nodes)

			nodes = append(nodes, the_node)
			continue
		}

		// actions
		{
			// if we get to this point, the line is most
			// likely just "action". however, "dialogue"
			// is the other non-identifiable syntax,
			// which depends on whether the previous
			// tokens are part of a "character train"
			// (which is "character", "dialogue" or
			// "parenthetical")

			expected_type := ACTION

			// change expected type in light of a "character train".
			if node, ok := get_last_node(nodes); ok {
				if is_character_train(node.node_type) {
					expected_type = DIALOGUE
					dirty_line = clean_line
				}
			}

			nodes = append(nodes, &syntax_node{
				node_type: expected_type,
				revised:   is_revised,
				raw_text:  dirty_line,
			})
		}
	}

	nodes = syntax_validator(nodes)

	return &fountain_content{
		title: title,
		nodes: nodes,
	}, true
}

// post-parse clean ups and checks
func syntax_validator(nodes []*syntax_node) []*syntax_node {
	if len(nodes) > 0 {
		// remove leading whitespace on file
		if nodes[0].node_type == WHITESPACE {
			nodes = nodes[1:]
		}

		if last_node, ok := get_last_node(nodes); ok {
			switch last_node.node_type {
			case WHITESPACE:
				nodes = nodes[:len(nodes)-1]

			// trailing character can't be a character
			case CHARACTER:
				last_node.node_type = ACTION
			}
		}

		// reset any false positives for characters
		for i, node := range nodes {
			if node.node_type != CHARACTER {
				continue
			}

			if len(nodes[i:]) > 1 {
				forward_node := nodes[i+1]

				// if two characters are back to back
				// reset the first one
				if forward_node.node_type == CHARACTER {
					forward_node.node_type = DIALOGUE
					continue
				}

				// if the lines following it are anything
				// other than valid character/dialogue content
				// reset it.
				if !is_character_train(forward_node.node_type) {
					node.node_type = ACTION
				}
			} else {
				// if it's the last thing in the file
				// reset it.
				node.node_type = ACTION
			}
		}
	}

	return nodes
}

func is_valid_scene(line string) bool {
	word := ""

	for i, c := range line {
		if c == '.' {
			word = line[:i]
			break
		}

		if c >= utf8.RuneSelf {
			if unicode.IsSpace(c) {
				word = line[:i]
				break
			}
			continue
		}

		if ascii_space[c] == 1 {
			word = line[:i]
			break
		}
	}

	if len(word) > 0 {
		return valid_scene[strings.ToLower(clean_string(word))]
	}

	return false
}

func is_valid_character(line string) bool {
	for i, c := range line {
		if !format_chars[c] {
			line = line[i:]
			break
		}
	}

	// characters must start with a letter
	if !unicode.IsLetter(rune(line[0])) {
		return false
	}

	has_letters := false
	first_char := true
	copy := line

	for len(copy) > 0 {
		c, rune_width := utf8.DecodeRuneInString(copy)

		if c == '(' && !first_char {
			for i, c := range copy {
				if c == ')' {
					copy = copy[i:]
					break
				}
			}
		}

		if unicode.IsLetter(c) {
			has_letters = true

			if !unicode.IsUpper(c) {
				return false
			}
		}

		copy = copy[rune_width:]
		first_char = false
	}

	return has_letters
}

func is_valid_transition(line string) bool {
	for i := len(line) - 1; i >= 0; i-- {
		c := line[i]
		if ascii_space[c] == 1 {
			return valid_transition[strings.ToLower(line[i+1:])]
		}
	}
	return false
}

func is_title_element(line string) (string, string, bool) {
	if !strings.Contains(line, ":") {
		return "", "", false
	}

	t := strings.SplitN(line, ":", 2)

	name, value := strings.ToLower(t[0]), t[1]

	if valid_title_page[name] {
		return name, value, true
	}

	return "", "", false
}

func has_scene_number(text string) (string, string, bool) {
	if text[len(text)-1] == '#' {
		n := 0
		t := text[:len(text)-1]

		for i := len(t) - 1; i > 0; i-- {
			the_rune, _ := utf8.DecodeLastRuneInString(t[:i+1])

			if unicode.IsSpace(the_rune) {
				break
			}

			if the_rune == '#' {
				n = i
				break
			}
		}

		if n != 0 {
			return strings.TrimSpace(text[:n]), t[n+1:], true
		}
	}

	return "", "", false
}

// quickly discards the title page for included files
func consume_title_page(input string) string {
	if n := strings.IndexRune(input, '\n'); n > -1 {
		if _, _, ok := is_title_element(input[:n]); ok {
			if n := rune_pair(input, '\n', '\n'); n > -1 {
				return strings.TrimSpace(input[n:])
			}
		}
	}
	return input
}