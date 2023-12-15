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

import "bytes"
import "strings"
import "strconv"
import "unicode"
import "unicode/utf8"
import "path/filepath"

import lib "github.com/signintech/gopdf"

type Fountain struct {
	Meta struct {
		Source  string `json:"source"`
		Version uint8  `json:"version"`
	} `json:"meta"`

	Title struct {
		has_any   bool
		Title     string `json:"title,omitempty"`
		Credit    string `json:"credit,omitempty"`
		Author    string `json:"author,omitempty"`
		Source    string `json:"source,omitempty"`
		Notes     string `json:"notes,omitempty"`
		DraftDate string `json:"draft_date,omitempty"`
		Copyright string `json:"copyright,omitempty"`
		Revision  string `json:"revision,omitempty"`
		Contact   string `json:"contact,omitempty"`
		Info      string `json:"info,omitempty"`
	} `json:"title"`

	Characters []Character `json:"characters,omitempty"`
	Content    []Section   `json:"content,omitempty"`

	paper    *lib.Rect
	template *Template

	header string
	footer string

	more_tag string
	cont_tag string

	chars_lookup   map[string]int
	counter_lookup map[string]*Counter
}

type Character struct {
	Name       string   `json:"name"`
	Gender     string   `json:"gender"`
	OtherNames []string `json:"other_names,omitempty"`
	Lines      int      `json:"lines_spoken,omitempty"`
}

type Section struct {
	page   int
	skip   bool
	is_raw bool

	pos_x        float64
	pos_y        float64
	total_height float64
	line_height  float64
	para_indent  float64 // applies to first line only; added to margin
	justify      uint8

	Type        Line_Type `json:"type"`
	Text        string    `json:"text,omitempty"`
	SceneNumber string    `json:"scene_number,omitempty"`
	Level       int       `json:"level,omitempty"`

	longest_line int
	lines []Line
}

/*
	cheat sheet
	set     value |  flag
	clear   value &^ flag
	toggle  value ^  flag
	has     value &  flag != 0
*/
type Leaf_Type int
const (
	NORMAL Leaf_Type = 1 << iota
	ESCAPE
	NEWLINE
	VARIABLE

	COUNTER
	COUNTER_ALPHA

	is_balanced

	UNDERLINE
	STRIKEOUT
	HIGHLIGHT

	does_break // causes a font change

	ITALIC
	BOLD
	NOTE

	NO_TYPE
)

type Line struct {
	length int
	style_reset Leaf_Type

	leaves    []Leaf
	underline []int
	strikeout []int
	highlight []int
}

type Leaf struct {
	leaf_type  Leaf_Type
	is_opening bool
	text       string
}

// intermediary structure with extra members for
// processing the leaves
type Inline_Format struct {
	Leaf
	space_width   int
	text_width    int
	could_open    bool
	could_close   bool
	space_only    bool
	counter_reset int
}

type Counter struct {
	_type Leaf_Type
	value int
}

func init_data() *Fountain {
	data := new(Fountain)

	data.Meta.Source  = MEANDER
	data.Meta.Version = DATA_VERSION

	return data
}

type Line_Type uint8
const (
	WHITESPACE Line_Type = iota

	PAGE_BREAK

	HEADER
	FOOTER

	is_printable

	ACTION
	SCENE

	CHARACTER
	DUAL_CHARACTER
	PARENTHETICAL
	DUAL_PARENTHETICAL
	DIALOGUE
	DUAL_DIALOGUE
	LYRIC
	DUAL_LYRIC

	TRANSITION
	SYNOPSIS
	CENTERED

	is_section

	SECTION
	SECTION2
	SECTION3

	TYPE_COUNT // used to set array lengths
)

func (x Line_Type) MarshalJSON() ([]byte, error) {
	buffer := new(bytes.Buffer)
	buffer.Grow(32)

	buffer.WriteRune('"')
	buffer.WriteString(x.String())
	buffer.WriteRune('"')

	return buffer.Bytes(), nil
}

func syntax_parser(config *Config, data *Fountain, text string) {
	data.chars_lookup   = make(map[string]int, 32)
	data.counter_lookup = make(map[string]*Counter, 32)
	data.Characters     = make([]Character, 0, 32) // we pre-empt needing these

	// only remove newlines in case the first
	// element is indented action
	text = consume_newlines(text)

	{
		// remove boneyards in a single step:
		// it's the only syntax that crosses a
		// line-boundary, so we deal with it now

		copy := new(strings.Builder)
		copy.Grow(len(text))

		eat_spaces    := false
		eat_newlines  := false
		last_rune     := '_'

		is_escaped := false

		for len(text) > 0 {
			if !config.include_notes {
				if text[0] == '[' && len(text) > 1 && text[1] == '[' {
					if is_escaped {
						text = text[2:]
						copy.WriteString("[[")
						is_escaped = false
						continue
					}

					n := rune_pair(text[2:], ']', ']')

					if n < 0 {
						copy.WriteString("[[")
						text = text[2:]
						continue
					}

					text = text[n + 2:]

					eat_newlines = (last_rune == '\n')
					eat_spaces   = (last_rune == ' ')
					continue
				}
			}

			if text[0] == '/' && len(text) > 1 && text[1] == '*' {
				if is_escaped {
					text = text[2:]
					copy.WriteString("/*")
					is_escaped = false
					continue
				}

				n := rune_pair(text[2:], '*', '/')

				if n < 0 {
					copy.WriteString("/*")
					text = text[2:]
					continue
				}

				parse_gender(data, text[2:n])

				text = text[n + 2:]

				eat_newlines = (last_rune == '\n')
				eat_spaces   = (last_rune == ' ')
				continue
			}

			r, width := utf8.DecodeRuneInString(text)
			text = text[width:]

			if r == '\\' {
				last_rune = '\\'

				if is_escaped {
					copy.WriteRune('\\')
					is_escaped = false
					continue
				}

				is_escaped = true
				continue
			}

			if is_escaped {
				copy.WriteRune('\\')
			}
			is_escaped = false

			last_rune = r

			if r == '\n' && eat_newlines {
				continue
			} else {
				eat_newlines = false
			}
			if r == ' ' && eat_spaces {
				continue
			} else {
				eat_spaces = false
			}

			copy.WriteRune(r)
		}

		text = copy.String() // return de-boned string
	}

	// title page mini-parser
	for {
		n, ok := find_title_colon(text)
		if !ok {
			break
		}

		word := homogenise(text[:n])
		text = text[n + 1:]

		word = homogenise(strings.TrimSpace(word))

		title_buffer := strings.Builder{}
		title_buffer.Grow(64)

		break_main_loop := false

		// begin parsing
		for {
			// grab the first line manually
			line := extract_to_newline(text)
			text = text[len(line):] // consume the line

			title_buffer.WriteString(strings.TrimSpace(line))

			if len(text) == 0 {
				break
			}

			if text[0] == '\n' {
				text = text[1:] // consume the newline

				if len(text) == 0 {
					break
				}

				if len(text) > 0 && text[0] == '\n' {
					break_main_loop = true
					break
				}
				if !unicode.IsSpace(rune(text[0])) {
					break
				}

				sub_line := extract_to_newline(text)
				text = text[len(sub_line):]

				title_buffer.WriteRune('\n')
				title_buffer.WriteString(strings.TrimSpace(sub_line))
			}
		}

		sub_line := left_trim(title_buffer.String())

		if sub_line != "" {
			switch word {
			case "title":
				data.Title.Title = sub_line
				data.Title.has_any = true
			case "credit":
				data.Title.Credit = sub_line
				data.Title.has_any = true
			case "author":
				data.Title.Author = sub_line
				data.Title.has_any = true
			case "source":
				data.Title.Source = sub_line
				data.Title.has_any = true
			case "notes":
				data.Title.Notes = sub_line
				data.Title.has_any = true
			case "draftdate":
				data.Title.DraftDate = sub_line
				data.Title.has_any = true
			case "copyright":
				data.Title.Copyright = sub_line
				data.Title.has_any = true
			case "revision":
				data.Title.Revision = sub_line
				data.Title.has_any = true
			case "contact":
				data.Title.Contact = sub_line
				data.Title.has_any = true
			case "info":
				data.Title.Info = sub_line
				data.Title.has_any = true

			case "header":
				data.header = sub_line
			case "footer":
				data.footer = sub_line

			case "conttag":
				data.cont_tag = sub_line
			case "moretag":
				data.more_tag = sub_line

			case "format", "template":
				if len(config.template) == 0 {
					config.template = sub_line
				}
			case "paper":
				if len(config.paper_size) == 0 {
					config.paper_size = sub_line
				}
			}
		}

		if break_main_loop {
			break
		}
	}

	// update any missing configuration by
	// applying defaults
	{
		if data.header == "" {
			data.header = "| #page."
		}
		if data.cont_tag == "" {
			data.cont_tag = DEFAULT_CONT_TAG
		}
		if data.more_tag == "" {
			data.more_tag = DEFAULT_MORE_TAG
		}
		if data.Title.Title == "" {
			data.Title.Title = filepath.Base(config.source_file)
		}
	}

	data.counter_lookup["wordcount"] = &Counter{value: word_count(text)}

	//
	// line parsing
	//
	nodes := make([]Section, 0, 256)

	for {
		if len(left_trim(text)) == 0 {
			break
		}

		count := count_rune(text, '\n')

		if count > 0 {
			if count == 1 {
				text = text[1:]
				continue
			}

			nodes = append(nodes, Section{
				Type:  WHITESPACE,
				Level: count - 1,
			})
			text = text[count:]
			continue
		}

		dirty_line := extract_to_newline(text)
		clean_line := strings.TrimSpace(dirty_line)
		text = text[len(dirty_line):]

		if len(clean_line) == 0 {
			nodes = append(nodes, Section{
				Type:  WHITESPACE,
				Level: 1,
			})
			continue
		}

		// this is just a simple guard-rail for lines that have
		// only a single character; they can't be anything but
		// action by definition and we cut a whole bunch of
		// additional len() checks out from subsequent code
		if len(clean_line) == 1 {
			nodes = append(nodes, Section{
				Type:  ACTION,
				Text:  clean_line,
			})
			continue
		}

		the_type := ACTION
		level    := 0

		switch clean_line[0] {
		case '!':
			the_type   = ACTION
			clean_line = left_trim(clean_line[1:])

			nodes = append(nodes, Section{
				Type:  the_type,
				Text:  clean_line,
			})
			continue

		case '@':
			the_type   = CHARACTER
			clean_line = clean_line[1:]

			if clean_line[len(clean_line) - 1] == '^' {
				clean_line = clean_line[:len(clean_line) - 1]
				level += 1
			}

		case '~':
			n := count_rune(clean_line, '~')
			if n == 2 {
				break
			}

			the_type   = LYRIC
			clean_line = left_trim(clean_line[1:])

		case '=':
			n := count_rune(clean_line, '=')

			if n >= 3 {
				the_type = PAGE_BREAK
			} else if n == 1 {
				the_type   = SYNOPSIS
				clean_line = left_trim(clean_line[1:])
			}

		case '#':
			// if we consider it to be a variable, we action it
			if r, _ := utf8.DecodeRuneInString(clean_line[1:]); !(r == '#' || unicode.IsSpace(r)) {
				the_type = ACTION

			// otherwise it's a section and we treat it
			// accordingly.
			} else {
				n := count_rune(clean_line, '#')

				the_type   = SECTION
				level      = n
				clean_line = left_trim(clean_line[n:])

				if level > 3 {
					level = 3
				}
			}

		case '(':
			if clean_line[len(clean_line) - 1] == ')' {
				if last_node, ok := get_last_section(nodes); ok {
					if is_character_train(last_node.Type) {
						the_type = PARENTHETICAL
					}
				}
			}

		case '>':
			clean_line = left_trim(clean_line[1:])

			the_type = TRANSITION

			if clean_line[len(clean_line) - 1] == '<' {
				the_type   = CENTERED
				clean_line = right_trim(clean_line[:len(clean_line) - 1])
			}

		case '.':
			if left_trim(clean_line[1:])[0] != '.' {
				name, number, ok := get_scene_number(left_trim(clean_line[1:]))
				if !ok {
					name = left_trim(clean_line[1:])
				}

				nodes = append(nodes, Section{
					Type:        SCENE,
					Text:        name,
					SceneNumber: number,
				})
				continue
			}
		}

		// if we're still marked as action
		// check general syntaxes
		if the_type == ACTION {
			if n := strings.IndexRune(clean_line, ':'); n > 0 {
				count := 0

				for _, c := range clean_line[:n] {
					if !unicode.IsLetter(c) {
						break
					}
					count += 1
				}

				if n == count {
					switch homogenise(clean_line[:n]) {
					case "header":
						the_type   = HEADER
						clean_line = left_trim(clean_line[n + 1:])
					case "footer":
						the_type   = FOOTER
						clean_line = left_trim(clean_line[n + 1:])
					}

					nodes = append(nodes, Section{
						Type: the_type,
						Text: clean_line,
					})
					continue
				}
			}

			if is_valid_scene(clean_line) {
				// can we combine this with the scene in the switch above? ^^
				name, number, ok := get_scene_number(left_trim(clean_line))
				if !ok {
					name = clean_line
				}

				nodes = append(nodes, Section{
					Type:        SCENE,
					Text:        name,
					SceneNumber: number,
				})
				continue

			} else if is_valid_transition(clean_line) {
				the_type = TRANSITION

			} else if is_valid_character(clean_line) {
				the_type = CHARACTER

				if clean_line[len(clean_line) - 1] == '^' {
					clean_line = strings.TrimSpace(clean_line[:len(clean_line) - 1])
					level += 1
				}

			} else {
				if last_node, ok := get_last_section(nodes); ok && is_character_train(last_node.Type) {
					the_type = DIALOGUE
				} else {
					clean_line = dirty_line // plain action uses dirty_line
				}
			}
		}

		nodes = append(nodes, Section{
			Type:  the_type,
			Level: level,
			Text:  clean_line,
		})
	}

	var last_char *Section
	any_visible := false

	for i := range nodes {
		node := &nodes[i]

		if !any_visible && node.Type > is_printable && !is_character_train(node.Type) {
			any_visible = true
		}

		if node.Type == CHARACTER {
			if i < len(nodes) - 1 {
				if !is_character_train(nodes[i + 1].Type) {
					node.Type = ACTION
					continue
				}
			} else {
				node.Type = ACTION
				continue
			}

			if node.Level == 1 {
				if last_char.Level == 2 || any_visible {
					node.Level = 0
				} else if last_char.Level == 0 {
					last_char.Level = 1
					node.Level = 2
				}
			}

			last_char = node
			any_visible = false

			name := strings.ToLower(node.Text)

			for i, c := range name {
				if c == '(' {
					name = strings.TrimSpace(name[:i])
					break
				}
			}

			if x, ok := data.chars_lookup[name]; ok {
				c := &data.Characters[x]
				c.Lines += 1
			} else {
				data.chars_lookup[name] = len(data.Characters)
				data.Characters = append(data.Characters, Character{
					Name:   title_case(name),
					Gender: "unknown",
					Lines:  1,
				})
			}
		}
	}

	has_dual := false
	level := 0
	for i := range nodes {
		node := &nodes[i]

		if node.Type == CHARACTER {
			has_dual = node.Level > 0
			if has_dual {
				node.Type += 1
				level = node.Level
			}
			continue
		}

		if is_character_train(node.Type) && has_dual {
			node.Type += 1
			node.Level = level
			continue
		}

		has_dual = false
	}

	data.Content = nodes
}

func parse_gender(data *Fountain, text string) {
	text = strings.TrimSpace(text)

	if !(len(text) > 8) {
		return
	}
	if strings.ToLower(text[:8]) != "[gender." {
		return
	}
	data.Characters   = make([]Character, 0, 32) // we pre-empt needing these
	current_gender := ""

	for len(text) > 0 {
		line := extract_to_newline(text)
		text = text[len(line):]
		line = strings.TrimSpace(line)
		text = left_trim(text)

		if line == "" {
			continue
		}

		if line[0] == '[' {
			if line[len(line) - 1] != ']' {
				return
			}

			line = strings.ToLower(strings.TrimSpace(line[1:len(line) - 1]))

			if !strings.HasPrefix(line, "gender.") {
				return
			}

			current_gender = line[7:]
			continue
		}

		names := strings.Split(line, "|")
		for i, entry := range names {
			names[i] = strings.TrimSpace(entry)
		}
		name := names[0]
		names = names[1:]

		n := len(data.Characters)

		if len(names) > 0 {
			for _, x := range names {
				data.chars_lookup[strings.ToLower(x)] = n
			}
		} else {
			names = nil
		}

		data.chars_lookup[strings.ToLower(name)] = n
		data.Characters = append(data.Characters, Character{
			Name:       name,
			Gender:     current_gender,
			OtherNames: names,
		})
	}
}

func get_last_section(nodes []Section) (*Section, bool) {
	if len(nodes) > 0 {
		return &nodes[len(nodes) - 1], true
	}
	return nil, false
}

func consume_title_page(input string) string {
	if n := strings.IndexRune(input, '\n'); n > -1 {
		if is_title_element(input[:n]) {
			if n := rune_pair(input, '\n', '\n'); n > -1 {
				return strings.TrimSpace(input[n:])
			}
		}
	}

	return input
}

func is_valid_character(line string) bool {
	for i, c := range line {
		if !is_format_char(c) {
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

func is_title_element(line string) bool {
	found_colon := false

	for i, c := range line {
		if c == ':' {
			line = strings.TrimSpace(homogenise(line[:i]))
			found_colon = true
			break
		}
	}

	if found_colon && lang_title_page(line) {
		return true
	}

	return false
}

func is_valid_transition(line string) bool {
	for i := len(line) - 1; i >= 0; i-- {
		c := line[i]
		if ascii_space[c] == 1 {
			return lang_transition(strings.ToLower(line[i+1:]))
		}
	}
	return false
}

// this conversion system obviously isn't
// perfect, but it supports many common
// formatters and fills in the gaps in Go's
// magic numbers to be tighter to the base
// Unicode spec
func nsdate(input string) string {
	final := strings.Builder{}
	final.Grow(len(input) * 2)

	t := now()

	const default_timestamp = "d MMM yyyy"

	nsconvert := func(x string) (string, bool) {
		switch x {
		case "M":    return "1", true
		case "MM":   return "01", true
		case "MMM":  return "Jan", true
		case "MMMM": return "January", true
		case "d":    return "2", true
		case "dd":   return "02", true
		case "E":    return "Mon", true
		case "EEEE": return "Monday", true
		case "h":    return "3", true
		case "hh":   return "03", true
		case "HH":   return "15", true
		case "a":    return "PM", true
		case "m":    return "4", true
		case "mm":   return "04", true
		case "s":    return "5", true
		case "ss":   return "05", true
		case "SSS":  return ".000", true
		}
		return "", false
	}

	input = strings.TrimSpace(input)

	for {
		if len(input) == 0 {
			break
		}

		for _, c := range input {
			if unicode.IsLetter(c) {
				n := count_rune(input, c)
				repeat := input[:n]

				// years
				if c == 'y' {
					switch n {
					case 1:
						final.WriteString(strconv.Itoa(t.Year()))
					case 2:
						final.WriteString(t.Format("06"))
					default:
						y := strconv.Itoa(t.Year())
						final.WriteString(strings.Repeat("0", clamp(n-len(y))))
						final.WriteString(y)
					}
					input = input[n:]
					break
				}

				// H - unpadded hour
				if c == 'H' && n == 1 {
					final.WriteString(strconv.Itoa(t.Hour()))
				}
				// MMMMM - single letter month
				if c == 'M' && n == 5 {
					final.WriteString(t.Month().String()[:1])
				}
				// EEEEE - single letter week
				if c == 'E' && n == 5 {
					final.WriteString(t.Weekday().String()[:1])
				}
				// EEEEEE - two letter week
				if c == 'E' && n == 6 {
					final.WriteString(t.Weekday().String()[:2])
				}

				if x, ok := nsconvert(repeat); ok {
					final.WriteString(t.Format(x))
				} else {
					return nsdate(default_timestamp) // just chuck the default back
				}

				input = input[n:]
				break
			}

			final.WriteRune(c)
			input = input[1:]
		}
	}

	return final.String()
}

func clamp(m int) int {
	if m < 0 {
		return 0
	}
	return m
}

func get_scene_number(text string) (string, string, bool) {
	if text[len(text) - 1] == '#' {
		n := 0
		t := text[:len(text) - 1]

		for i := len(t) - 1; i > 0; i-- {
			the_rune, _ := utf8.DecodeLastRuneInString(t[:i + 1])

			if unicode.IsSpace(the_rune) {
				break
			}
			if the_rune == '#' {
				n = i
				break
			}
		}

		if n != 0 {
			return strings.TrimSpace(text[:n]), t[n + 1:], true
		}
	}

	return "", "", false
}

func is_character_train(node_type Line_Type) bool {
	switch node_type {
	case CHARACTER, PARENTHETICAL, DIALOGUE, LYRIC:
		return true
	}
	return false
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
		return lang_scene(strings.ToLower(clean_string(word)))
	}

	return false
}

func copy_without_style(incoming Section) Section {
	new_lines := make([]Line, len(incoming.lines))

	for i, line := range incoming.lines {
		new_lines[i].length      = line.length
		new_lines[i].style_reset = NORMAL

		new_lines[i].leaves = make([]Leaf, len(line.leaves))

		for x, leaf := range line.leaves {
			leaf.leaf_type = NORMAL
			new_lines[i].leaves[x] = leaf
		}

		new_lines[i].underline = []int{}
		new_lines[i].strikeout = []int{}
		new_lines[i].highlight = []int{}
	}

	incoming.lines = new_lines

	return incoming
}

const LINE_TYPE_NAMES = "whitespacepage_breakheaderfooteris_printableactionscenecharacterdual_characterparentheticaldual_parentheticaldialoguedual_dialoguelyricdual_lyrictransitionsynopsiscenteredis_sectionsectionsection2section3type_count"

var LINE_TYPE_INDICES = [...]uint8{0, 10, 20, 26, 32, 44, 50, 55, 64, 78, 91, 109, 117, 130, 135, 145, 155, 163, 171, 181, 188, 196, 204, 214}

func (i Line_Type) String() string {
	return LINE_TYPE_NAMES[LINE_TYPE_INDICES[i]:LINE_TYPE_INDICES[i+1]]
}
