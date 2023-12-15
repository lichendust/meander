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

import "fmt"
import "sort"
import "strings"
import "strconv"

func line_override(line *Line, style Leaf_Type) {
	if style & UNDERLINE != 0 { line.underline = []int{0, line.length} }
	if style & STRIKEOUT != 0 { line.strikeout = []int{0, line.length} }
	if style & HIGHLIGHT != 0 { line.highlight = []int{0, line.length} }
	line.style_reset = style
}

func style_override(section *Section, style Leaf_Type) {
	for i := range section.lines {
		line := &section.lines[i]
		line_override(line, style)
	}
}

// @note some of the inside_dual_dialogue state
// stuff is a confusing read -- clean up or rework
func paginate(config *Config, data *Fountain) {
	template := data.template

	last_type       := WHITESPACE
	running_height  := MARGIN_TOP
	max_page_height := data.paper.H - MARGIN_BOTTOM
	first_on_page   := true
	page_number     := 1

	data.counter_lookup["page"]  = &Counter{COUNTER, page_number}
	data.counter_lookup["scene"] = &Counter{COUNTER, 0}

	// used for when we have to jump back
	old_running_height   := MARGIN_TOP
	margin_adjust_dual   := float64(0)
	delayed_page_number  := false
	inside_dual_dialogue := 0

	original_content := data.Content
	data.Content = make([]Section, 0, len(data.Content))

	var last_char *Section

	new_page := func() {
		if inside_dual_dialogue != 2 {
			do_header(data, data.footer, FOOTER, page_number)
		}

		running_height = MARGIN_TOP
		page_number += 1
		first_on_page = true

		data.counter_lookup["page"] = &Counter{COUNTER, page_number}

		if inside_dual_dialogue != 2 {
			do_header(data, data.header, HEADER, page_number)
		}
	}

	// initial header/footer, if any
	do_header(data, data.header, HEADER, page_number)
	do_header(data, data.footer, FOOTER, page_number)

	for content_index := range original_content {
		section := &original_content[content_index]

		if section.Type == WHITESPACE && inside_dual_dialogue == 1 {
			continue
		}

		if section.Type == SCENE && config.scenes == SCENE_GENERATE {
			data.counter_lookup["scene"].value += 1
			section.SceneNumber = fmt.Sprintf("%d", data.counter_lookup["scene"].value)
		}

		// we adjust this here so the raw data struct
		// stays clean and free of strange section levels
		// for other commands
		if section.Type == SECTION {
			section.Type += Line_Type(section.Level - 1)
		}

		t := template.types[section.Type]

		section.skip = t.skip

		if running_height >= max_page_height && inside_dual_dialogue != 1 {
			new_page()

			if inside_dual_dialogue == 2 {
				delayed_page_number = false
				first_on_page = false
			}
		}

		// might be too aggressive
		if inside_dual_dialogue == 2 && section.Type == WHITESPACE || section.Level == 0 {
			inside_dual_dialogue = 0

			if running_height < old_running_height {
				running_height = old_running_height
			}

			old_running_height = 0

			if delayed_page_number {
				page_number += 1
				delayed_page_number = false
				data.counter_lookup["page"].value = page_number
			}
		}

		if section.Type >= is_printable {
			if t.skip {
				continue
			}

			if section.Type == CHARACTER || section.Type == DUAL_CHARACTER {
				last_char = section
				switch section.Level {
				case 1:
					inside_dual_dialogue = 1
					old_running_height = running_height
				case 2:
					inside_dual_dialogue = 2
					old_running_height, running_height = running_height, old_running_height
				}
			}

			if inside_dual_dialogue == 2 {
				margin_adjust_dual = data.paper.W / 2 - MARGIN_LEFT / 2
			} else {
				margin_adjust_dual = 0
			}

			if !first_on_page {
				running_height += t.space_above
			}
			first_on_page = false

			switch t.casing {
			case UPPERCASE: section.Text = strings.ToUpper(section.Text)
			case LOWERCASE: section.Text = strings.ToLower(section.Text)
			}

			section.line_height = t.line_height
			section.justify     = t.justify
			section.lines = break_section(data, section, t.width, t.para_indent, t.style != NORMAL)

			if t.trail_height > 0 && running_height >= max_page_height - t.trail_height {
				new_page()
				first_on_page = false

				if inside_dual_dialogue == 2 {
					delayed_page_number = false
				}
			}

			section.pos_y = running_height
			section.page = page_number

			switch t.justify {
			default:
				section.pos_x = template.left_margin + t.margin
			case RIGHT:
				section.pos_x = template.right_margin - t.margin
			case CENTER:
				section.pos_x = template.center_line
			}

			section.pos_x += margin_adjust_dual

			if t.style != NORMAL {
				style_override(section, t.style)
			}

			if !section.is_raw {
				page_break_length := 0

				for i := 1; i <= len(section.lines); i += 1 {
					if running_height + float64(i) * section.line_height > max_page_height {
						page_break_length = i
						break
					}
				}

				if page_break_length > 0 && page_break_length != len(section.lines) {
					copy_section := *section
					copy_section.lines = section.lines[:page_break_length]
					data.Content = append(data.Content, copy_section)

					old_running_height := running_height + float64(page_break_length) * section.line_height
					old_page_number    := page_number

					new_page()
					first_on_page = false

					needs_page_reset := false // dual dialogue has to undo the page bump

					if inside_dual_dialogue == 1 {
						needs_page_reset    = true
						delayed_page_number = true
					}
					if inside_dual_dialogue == 2 {
						delayed_page_number = false
					}

					if section.Type == DIALOGUE || section.Type == DUAL_DIALOGUE {
						dual_offset := Line_Type(0)
						if section.Type == DUAL_DIALOGUE {
							dual_offset = 1
						}

						local_t := template.types[PARENTHETICAL + dual_offset]

						data.Content = append(data.Content, Section{
							pos_x:   MARGIN_LEFT + local_t.margin + margin_adjust_dual,
							pos_y:   old_running_height,
							justify: local_t.justify,
							page:    old_page_number,
							is_raw:  true,
							Type:    PARENTHETICAL + dual_offset,
							Text:    "(more)",
							Level:   section.Level,
						})

						local_t = template.types[CHARACTER + dual_offset]

						new_text := last_char.Text

						// if an identical cont'd tag already exists, we ignore
						if !strings.HasSuffix(new_text, data.cont_tag) {
							new_text += " " + data.cont_tag
						}

						// @todo we don't check if this is raw or not -- do that
						data.Content = append(data.Content, Section{
							pos_x:   MARGIN_LEFT + local_t.margin + margin_adjust_dual,
							pos_y:   running_height,
							justify: local_t.justify,
							page:    page_number,
							is_raw:  true,
							Type:    CHARACTER + dual_offset,
							Text:    new_text,
							Level:   section.Level,
						})

						running_height += local_t.line_height
					}

					section.lines = section.lines[page_break_length:]
					section.pos_y = running_height
					section.page  = page_number

					if needs_page_reset {
						page_number = old_page_number
						data.counter_lookup["page"].value = page_number
					}
				}

				running_height += section.line_height * float64(len(section.lines))

			} else {
				running_height += section.line_height
			}

			data.Content = append(data.Content, *section)
		}

		switch section.Type {
		case HEADER:
			data.header = section.Text

		case FOOTER:
			data.footer = section.Text

		case PAGE_BREAK:
			new_page()

		case WHITESPACE:
			if !first_on_page {
				level := section.Level

				if last_type < is_printable {
					level -= 1
				}

				running_height += template.line_height * float64(level)

				if running_height > max_page_height {
					new_page()
				}
			}
		}

		last_type = section.Type
	}

	// this solves the 'corrupted' order of dual dialogue
	// entries when they break across pages.
	// in a dual-free screenplay, it's an O(n) loop that makes
	// no changes.
	// it still slows things down, but it's unbelievable how
	// much simpler all of the code is made above by doing this.
	sort.Stable(Page_Sorter(data.Content))
}

func do_header(data *Fountain, text string, the_type Line_Type, page_number int) {
	if text == "" {
		return
	}

	y_pos := data.template.header_height

	if the_type == FOOTER {
		y_pos = data.template.footer_height
	}

	split := strings.SplitN(text, "|", 3)

	switch len(split) {
	case 1:
		one := quick_section(data, split[0], LEFT, LINE_HEIGHT, INCH * 5)
		one.pos_x = MARGIN_LEFT
		one.pos_y = y_pos
		one.Type  = the_type
		one.page  = page_number

		data.Content = append(data.Content, *one)

	case 2:
		one := quick_section(data, split[0], LEFT, LINE_HEIGHT, INCH * 3)
		one.pos_x = MARGIN_LEFT
		one.pos_y = y_pos
		one.Type  = the_type
		one.page  = page_number

		two := quick_section(data, split[1], RIGHT, LINE_HEIGHT, INCH * 3)
		two.pos_x = data.paper.W - MARGIN_RIGHT
		two.pos_y = y_pos
		two.Type  = the_type
		two.page  = page_number

		data.Content = append(data.Content, *one)
		data.Content = append(data.Content, *two)

	case 3:
		one := quick_section(data, split[0], LEFT, LINE_HEIGHT, INCH * 2)
		one.pos_x = MARGIN_LEFT
		one.pos_y = y_pos
		one.Type  = the_type
		one.page  = page_number

		two := quick_section(data, split[1], CENTER, LINE_HEIGHT, INCH * 2)
		two.pos_x = data.paper.W / 2
		two.pos_y = y_pos
		two.Type  = the_type
		two.page  = page_number

		three := quick_section(data, split[2], RIGHT, LINE_HEIGHT, INCH * 2)
		three.pos_x = data.paper.W - MARGIN_RIGHT
		three.pos_y = y_pos
		three.Type  = the_type
		three.page  = page_number

		data.Content = append(data.Content, *one)
		data.Content = append(data.Content, *two)
		data.Content = append(data.Content, *three)
	}
}

type Page_Sorter []Section
func (oc Page_Sorter) Len() int           { return len(oc) }
func (oc Page_Sorter) Less(i, j int) bool { return oc[i].page < oc[j].page }
func (oc Page_Sorter) Swap(i, j int)      { oc[i], oc[j] = oc[j], oc[i] }

func break_section(data *Fountain, section *Section, width float64, para_indent int, known_style bool) []Line {
	input := section.Text

	max_width := 0
	if width == 0 {
		max_width = int((data.paper.W - MARGIN_LEFT - MARGIN_RIGHT) / CHAR_WIDTH)
	} else {
		max_width = int(width / CHAR_WIDTH)
	}
	test_width := max_width - para_indent // first line might be indented

	section.para_indent = float64(para_indent) * CHAR_WIDTH

	for _, c := range input {
		if is_format_char(c) {
			known_style = true
			break
		}
	}

	if !known_style {
		length := rune_count(input)
		if length < max_width {
			section.longest_line = length
			section.total_height = section.line_height
			section.is_raw = true
			return nil
		}
	}

	the_list := make([]Inline_Format, 0, 16)

	for {
		if len(input) == 0 {
			break
		}

		space_width := count_whitespace(input)
		input = left_trim_ignore_newlines(input)

		if len(input) == 0 {
			break
		}

		var the_word   string
		var the_type   Leaf_Type
		var byte_width int
		counter_reset := -1

		current_rune := input[0]

		switch current_rune {
		case '\\':
			the_word, byte_width = extract_repeated_rune(input, '\\')
			the_type = ESCAPE

		case '\n':
			the_word = "\n"
			the_type = NEWLINE
			byte_width = 1

		case '*':
			the_word, byte_width = extract_repeated_rune(input, '*')

			switch byte_width {
			case 1:
				the_type = ITALIC
			case 2:
				the_type = BOLD
			case 3:
				the_type = BOLD | ITALIC
			}

		case '$':
			_, keyword_width := extract_ident(input[1:])
			if keyword_width > 0 {
				byte_width = keyword_width + 1
				the_word   = input[:byte_width]
				the_type   = VARIABLE
			} else {
				byte_width = 1
				the_type   = NORMAL
				the_word   = "$"
			}

		case '#':
			_, keyword_width := extract_ident(input[1:])
			if keyword_width > 0 {
				byte_width = keyword_width + 1
				the_word   = input[:byte_width]
				the_type   = COUNTER

				if len(input) > byte_width && input[byte_width] == ':' {
					x, w := extract_letters_or_numbers(input[byte_width + 1:])
					byte_width += w + 1

					is_numbers := is_all_numbers(x)
					is_letters := is_all_letters(x)

					if is_numbers {
						v, _ := strconv.Atoi(x)
						counter_reset = v
					} else if is_letters {
						the_type = COUNTER_ALPHA
						counter_reset = alphabet_to_int(x)
					}

					the_word = input[:byte_width]
				}

			} else {
				byte_width = 1
				the_type   = NORMAL
				the_word   = "#"
			}

		case '_':
			the_word, byte_width = extract_repeated_rune(input, '_')
			the_type = UNDERLINE

			if byte_width > 1 {
				the_type = NORMAL
			}

		case '+':
			the_word, byte_width = extract_repeated_rune(input, '+')
			the_type = HIGHLIGHT

			if byte_width > 1 {
				the_type = NORMAL
			}

		case '~':
			the_word, byte_width = extract_repeated_rune(input, '~')
			the_type = STRIKEOUT

			if byte_width != 2 {
				the_type = NORMAL
			}

		case '[':
			the_word, byte_width = extract_repeated_rune(input, '[')
			the_type = NOTE

			if byte_width < 2 {
				the_type = NORMAL
			}

		case ']':
			the_word, byte_width = extract_repeated_rune(input, ']')
			the_type = NOTE

			if byte_width < 2 {
				the_type = NORMAL
			}

		default:
			the_word, byte_width = non_token_word(input)
			the_type = NORMAL
		}

		if the_type != NORMAL {
			known_style = true
		}

		var format Inline_Format

		format.leaf_type     = the_type
		format.text          = the_word
		format.text_width    = rune_count(the_word)
		format.space_width   = space_width
		format.counter_reset = counter_reset

		if the_type == ESCAPE {
			format.space_only = true
		}
		if the_type == NOTE {
			format.is_opening = (the_word[0] == '[')
		}

		the_list = append(the_list, format)

		input = input[byte_width:]
	}

	if known_style {
		length := len(the_list) - 1
		last_type := NO_TYPE

		for i := range the_list {
			entry := &the_list[i]

			if entry.leaf_type == ESCAPE {
				n := len(entry.text)

				if n % 2 == 0 {
					entry.leaf_type = NORMAL
					entry.space_only = false
					entry.text = entry.text[n / 2:]
					last_type = entry.leaf_type
					continue
				}

				if n > 1 {
					entry.leaf_type = NORMAL
					entry.space_only = false
					entry.text = entry.text[(n + 1) / 2 - 1:]
				}

				if len(the_list[i:]) > 1 {
					target := &the_list[i + 1]
					target.leaf_type = NORMAL
				}

				last_type = entry.leaf_type
				continue
			}

			if entry.leaf_type < is_balanced {
				last_type = entry.leaf_type
				continue
			}

			has_spaces_before := false

			if entry.space_width  == 0 {
				entry.could_close = true
			} else {
				has_spaces_before = true
			}

			test := i + 1

			if test > length {
				last_type = entry.leaf_type
				continue
			}

			// if the next chunk has no leading spaces
			// (we're smushed against it) then we
			// _could_ be an opening
			if the_list[test].space_width == 0 {
				entry.could_open = true
			} else if has_spaces_before {
				// if we're floating (spaces before and after)
				// we have to be normal text, because fountain
				// doesn't allow for us to exist.
				entry.leaf_type = NORMAL
			}

			// if we're marked as closing
			if entry.leaf_type != NORMAL && entry.could_close && last_type != NORMAL {
				entry.could_close = false
			}
		}

		// balance the states of bold/italics etc.
		var state [NO_TYPE]struct{
			live bool
			last *Inline_Format
		}

		for i := range the_list {
			entry := &the_list[i]

			switch entry.leaf_type {
			default:
				s := &state[entry.leaf_type]
				s.live = inline_balance(entry, s.live)
				if entry.is_opening {
					s.last = entry
				}

				/*if entry.leaf_type == NOTE {
					fmt.Println(entry.is_opening)
				}*/

			case BOLD | ITALIC:
				b := &state[BOLD]
				i := &state[ITALIC]

				x := inline_balance(entry, b.live && i.live)
				b.live = x
				i.live = x

				if entry.is_opening {
					b.last = entry
					i.last = entry
				}

			case COUNTER, COUNTER_ALPHA:
				word := homogenise(entry.text[1:])
				switch word {
				case "page", "scene", "wordcount":
					entry.text = fmt.Sprintf("%d", data.counter_lookup[word].value)

				default:
					// this is here because it's the only way to
					// preserve the original text when escaped
					// because I didn't write a good parser.
					sub_word, _ := extract_ident(word)
					sub_word     = homogenise(sub_word)

					x, exists := data.counter_lookup[sub_word]
					if !exists {
						x = new(Counter)
						x._type = entry.leaf_type
						x.value = 1

						data.counter_lookup[sub_word] = x
					}

					if entry.counter_reset >= 0 {
						x.value = entry.counter_reset
					}

					if x._type == COUNTER_ALPHA {
						if x.value < 1 { x.value = 1 }

						entry.text = alphabetical_increment(x.value)
					} else {
						entry.text = fmt.Sprintf("%d", x.value)
					}

					x.value += 1
				}

				entry.leaf_type  = NORMAL
				entry.text_width = rune_count(entry.text)

			case VARIABLE:
				switch homogenise(entry.text[1:]) {
				case "title":
					entry.text = clean_string(data.Title.Title)
				case "author":
					entry.text = clean_string(data.Title.Author)
				case "source":
					entry.text = clean_string(data.Title.Source)
				case "notes":
					entry.text = clean_string(data.Title.Notes)
				case "draftdate":
					entry.text = clean_string(data.Title.DraftDate)
				case "copyright":
					entry.text = clean_string(data.Title.Copyright)
				case "revision":
					entry.text = clean_string(data.Title.Revision)
				case "contact":
					entry.text = clean_string(data.Title.Contact)
				case "info":
					entry.text = clean_string(data.Title.Info)
				case "date":
					entry.text = nsdate("dd/MM/yyyy") // @todo
				}

				entry.leaf_type  = NORMAL
				entry.text_width = rune_count(entry.text)

			case NORMAL:
				continue
			}
		}

		// if any of these tracts of styling are still 'live' at
		// the end of a line, the thing that started that
		// styling is unbalanced, so we revert it
		for i := range state {
			if state[i].live {
				state[i].last.leaf_type = NORMAL
			}
		}
	}

	line_stack := make([]Line, 0, len(the_list) / 2)
	leaf_stack := make([]Leaf, 0, 8)

	line_buffer := new(strings.Builder)
	line_buffer.Grow(max_width)

	underline_range := make([]int, 0, 4)
	strikeout_range := make([]int, 0, 4)
	highlight_range := make([]int, 0, 4)

	// we track these through the lines
	line_length  := 0
	current_type := NORMAL
	cancel_space := false

	for _, entry := range the_list {
		if entry.leaf_type != NORMAL {
			if entry.leaf_type > does_break {
				if line_buffer.Len() > 0 {
					leaf_stack = append(leaf_stack, Leaf{
						leaf_type:  current_type,
						is_opening: entry.is_opening,
						text:       line_buffer.String(),
					})

					line_buffer.Reset()
					line_buffer.Grow(max_width)
				}
			}

			current_type ^= entry.leaf_type // toggle the incoming style on/off

			switch entry.leaf_type {
			case UNDERLINE: underline_range = append(underline_range, line_length + entry.space_width)
			case STRIKEOUT: strikeout_range = append(strikeout_range, line_length + entry.space_width)
			case HIGHLIGHT: highlight_range = append(highlight_range, line_length + entry.space_width)
			}
		}

		if line_length + entry.space_width > test_width || entry.leaf_type == NEWLINE || entry.leaf_type == NORMAL && line_length + entry.text_width > test_width {
			if line_buffer.Len() > 0 {
				leaf_stack = append(leaf_stack, Leaf{
					leaf_type:  current_type,
					is_opening: entry.is_opening,
					text:       line_buffer.String(),
				})

				line_buffer.Reset()
				line_buffer.Grow(max_width)
			}

			if current_type & UNDERLINE != 0 { underline_range = append(underline_range, line_length) }
			if current_type & STRIKEOUT != 0 { strikeout_range = append(strikeout_range, line_length) }
			if current_type & HIGHLIGHT != 0 { highlight_range = append(highlight_range, line_length) }

			line_stack = append(line_stack, Line{
				length:    line_length,
				leaves:    leaf_stack,
				underline: underline_range,
				strikeout: strikeout_range,
				highlight: highlight_range,
			})

			test_width = max_width
			leaf_stack = make([]Leaf, 0, 8)

			underline_range = make([]int, 0, 4)
			strikeout_range = make([]int, 0, 4)
			highlight_range = make([]int, 0, 4)

			if current_type & UNDERLINE != 0 { underline_range = append(underline_range, 0) }
			if current_type & STRIKEOUT != 0 { strikeout_range = append(strikeout_range, 0) }
			if current_type & HIGHLIGHT != 0 { highlight_range = append(highlight_range, 0) }

			cancel_space = true
			line_length = 0
		}

		if entry.space_width > 0 && !cancel_space {
			line_buffer.WriteString(strings.Repeat(" ", entry.space_width))
			line_length += entry.space_width
		}

		if entry.leaf_type == NORMAL {
			cancel_space = false

			line_buffer.WriteString(entry.text)
			line_length += entry.text_width
		}
	}

	if line_buffer.Len() > 0 {
		leaf_stack = append(leaf_stack, Leaf{
			leaf_type: current_type,
			text:      line_buffer.String(),
		})
	}
	if len(leaf_stack) > 0 {
		line_stack = append(line_stack, Line{
			length:    line_length,
			leaves:    leaf_stack,
			underline: underline_range,
			strikeout: strikeout_range,
			highlight: highlight_range,
		})
	}

	for _, line := range line_stack {
		if line.length > section.longest_line {
			section.longest_line = line.length
		}
	}

	section.total_height = float64(len(section.lines)) * section.line_height
	return line_stack
}

func inline_balance(entry *Inline_Format, check bool) bool {
	if entry.could_open && entry.could_close {
		if check {
			return false
		} else {
			entry.is_opening = true
			return true
		}
	} else if entry.could_open {
		if check {
			entry.leaf_type = NORMAL
			return check
		} else {
			entry.is_opening = true
			return true
		}
	} else if entry.could_close {
		if check {
			return false
		} else {
			entry.leaf_type = NORMAL
		}
	}
	return check
}