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
	"time"
	"strings"
	"strconv"
	"unicode"
	"unicode/utf8"
	"path/filepath"

	lib "github.com/signintech/gopdf"
)

type Fountain struct {
	Meta struct {
		Source  string       `json:"source"`
		Version uint8        `json:"version"`
	} `json:"meta"`

	Title struct {
		has_any   bool
		Title     string     `json:"title,omitempty"`
		Credit    string     `json:"credit,omitempty"`
		Author    string     `json:"author,omitempty"`
		Source    string     `json:"source,omitempty"`
		Notes     string     `json:"notes,omitempty"`
		DraftDate string     `json:"draft_date,omitempty"`
		Copyright string     `json:"copyright,omitempty"`
		Revision  string     `json:"revision,omitempty"`
		Contact   string     `json:"contact,omitempty"`
		Info      string     `json:"info,omitempty"`
	} `json:"title"`

	Characters []Character   `json:"characters,omitempty"`
	Content    []Syntax_Node `json:"content,omitempty"`

	paper    *lib.Rect
	template *Template
}

func init_data() *Fountain {
	data := Fountain{}

	data.Meta.Source  = MEANDER
	data.Meta.Version = 1

	data.paper    = lib.PageSizeA4
	data.template = &SCREENPLAY

	return &data
}

const (
	WHITESPACE uint8 = iota
	PAGE_BREAK
	HEADER
	FOOTER
	PAGE_NUMBER
	SCENE_NUMBER

	ACTION
	LIST
	SCENE
	CHARACTER
	PARENTHETICAL
	DIALOGUE
	LYRIC
	TRANSITION
	SYNOPSIS

	JUSTIFY_CENTER
	JUSTIFY_RIGHT

	SECTION
	SECTION2
	SECTION3

	TYPE_COUNT
)

type Character struct {
	hash       uint32
	Name       string   `json:"name"`
	Gender     string   `json:"gender"`
	OtherNames []string `json:"other_names,omitempty"`
	Lines      int      `json:"lines_spoken,omitempty"`
}

type Syntax_Node struct {
	page   int
	pos_x  float64
	pos_y  float64
	width  float64
	height float64

	visible bool
	hash    uint32

	Type    uint8   `json:"type"`
	Level   uint8	`json:"-"`
	Revised bool    `json:"revised,omitempty"`

	Text string  `json:"text,omitempty"`
}

type Syntax_Line struct {

}

type Syntax_Leaf struct {

}

var format_chars = map[rune]bool{
	'*':  true,
	'+':  true,
	'~':  true,
	'_':  true,
	']':  true,
	'[':  true,
	'"':  true,
	'\'': true,
	'\\': true,
}

const (
	NORMAL int = 1 << iota
	ITALIC
	BOLD
	BOLDITALIC
	UNDERLINE
	STRIKEOUT
	HIGHLIGHT
	NOTE
	QUOTE
	DOUBLE_QUOTE
	ESCAPE
)

/*
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
*/

func syntax_parser(config *config, data *Fountain, text string) {
	// title page mini-parser
	first := false

	for {
		n, ok := find_title_colon(text)
		if !ok {
			break
		}

		word := homogenise(text[:n])
		text = text[n + 1:]

		if config.revision && first {
			// because of the need to always skip the diff char
			// to check if we're at the end of the title page
			// below, we end up returning to this point for
			// each title entry with the diff char already
			// removed - we only need to do it the first time
			word = word[1:]
			first = false
		}

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

				// eat the diff char
				if config.revision {
					text = text[1:]
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

				// eat the diff char
				if config.revision {
					sub_line = sub_line[1:]
				}

				title_buffer.WriteRune('\n')
				title_buffer.WriteString(strings.TrimSpace(sub_line))
			}
		}

		sub_line := consume_whitespace(title_buffer.String())

		if sub_line != "" {
			switch word {
			case "format":
				data.template = set_template(sub_line)
			case "paper":
				data.paper = set_paper(sub_line)
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
			}
		}

		if break_main_loop {
			break
		}
	}

	if data.Title.Title == "" {
		data.Title.Title = filepath.Base(config.source_file)
	}

	// only remove newlines in case the first
	// element something like indented "action".
	text = consume_newlines(text)

	{
		// remove boneyards in a single step:
		// it's the only syntax that crosses a
		// line-boundary, so we deal with it now

		copy := strings.Builder{}
		copy.Grow(len(text))

		is_escaped    := false
		eat_spaces    := false
		eat_newlines  := false
		newline_count := 0
		last_rune     := '_'

		for len(text) > 0 {
			if text[0] == '\\' {
				if is_escaped {
					copy.WriteRune('\\')
				}
				is_escaped = !is_escaped
				text = text[1:]
				continue
			}

			if text[0] == '/' && len(text) > 1 && text[1] == '*' {
				n := rune_pair(text[2:], '*', '/')

				if n < 0 {
					copy.WriteString("/*")
					text = text[2:]
					continue
				}

				parse_gender(data, text[2:n]) // (tests and performs it)
				text = text[n + 2:]

				eat_newlines = (last_rune == '\n')
				eat_spaces   = (last_rune == ' ')
			}

			r, width := utf8.DecodeRuneInString(text)
			text = text[width:]

			last_rune = r

			if r == '\n' && eat_newlines && newline_count < 2 {
				newline_count += 1
				continue
			} else {
				newline_count = 0
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
}

func parse_gender(data *Fountain, text string) {
	text = strings.TrimSpace(text)

	if !(len(text) > 8) {
		return
	}
	if strings.ToLower(text[:8]) != "[gender." {
		return
	}

	if data.Characters == nil {
		data.Characters = make([]Character, 0, count_rune(text, '\n'))
	}

	current_gender := ""

	for len(text) > 0 {
		line := extract_to_newline(text)
		text = text[len(line):]
		line = strings.TrimSpace(line)
		text = consume_whitespace(text)

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

		if len(names) == 0 {
			names = nil
		}

		data.Characters = append(data.Characters, Character{
			hash:       uint32_from_string(name),
			Name:       name,
			Gender:     current_gender,
			OtherNames: names,
		})
	}
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

	if found_colon && valid_title_page[line] {
		return true
	}

	return false
}

var valid_title_page = map[string]bool{
	// fountain
	"title":      true,
	"credit":     true,
	"author":     true,
	"source":     true,
	"contact":    true,
	"revision":   true,
	"copyright":  true,
	"draftdate":  true, // we homogenise spaces
	"notes":      true,

	// meander
	"format": true,
	"paper":  true,

	"conttag": true,
	"moretag": true,
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

const default_timestamp = "d MMM yyyy"

var ns_magic_convert = map[string]string{
	"M":    "1",
	"MM":   "01",
	"MMM":  "Jan",
	"MMMM": "January",
	"d":    "2",
	"dd":   "02",
	"E":    "Mon",
	"EEEE": "Monday",
	"h":    "3",
	"hh":   "03",
	"HH":   "15",
	"a":    "PM",
	"m":    "4",
	"mm":   "04",
	"s":    "5",
	"ss":   "05",
	"SSS":  ".000",
}

// this conversion system obviously isn't
// perfect, but it supports many common
// formatters and fills in the gaps in Go's
// magic numbers to be tighter to the base
// Unicode spec
func nsdate(input string) string {
	final := strings.Builder{}

	t := time.Now()

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

				if x, ok := ns_magic_convert[repeat]; ok {
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