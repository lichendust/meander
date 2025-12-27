/*
	Meander
	A portable Fountain utility for production writing
	Copyright (C) 2022-2023 Harley Denham
*/

package main

import "strings"
import "strconv"
import "unicode"
import "unicode/utf8"

import lib "github.com/signintech/gopdf"

const (
	PICA float64 = 12
	INCH float64 = PICA * 6

	LINE_HEIGHT = PICA

	MARGIN_TOP    = INCH
	MARGIN_LEFT   = INCH * 1.5
	MARGIN_RIGHT  = INCH
	MARGIN_BOTTOM = INCH
)

const (
	LEFT uint8 = iota
	RIGHT
	CENTER

	NONE uint8 = iota
	UPPERCASE
	LOWERCASE
)

type Color struct {
	R, G, B uint8
}

type Format uint8
const (
	SCREENPLAY Format = iota
	STAGEPLAY
	GRAPHIC_NOVEL
	MANUSCRIPT
	MANUSCRIPT_COMPACT
	DOCUMENT

	// exp
	STORYBOARD
)

type Template struct {
	kind  Format
	paper lib.Rect

	landscape         bool
	ignore_whitespace bool
	title_page_align  uint8

	line_height       float64
	margin_left       float64
	margin_right      float64
	margin_top        float64
	margin_bottom     float64
	center_line       float64
	dual_right_offset float64

	starred_margin float64
	starred_nudge  float64 // tiny offset for the height of starred rev asterisks

	header_margin float64
	footer_margin float64

	text_color      Color
	note_color      Color
	highlight_color Color

	types [TYPE_COUNT]Template_Entry
}

type Template_Entry struct {
	skip bool

	style   Leaf_Type // force style override (bitwise)
	casing  uint8     // force upper/lower case
	justify uint8     // align to left or right margin

	// 0 = ignored
	margin       float64 // bump up left or right margin
	width        float64 // max width of wrap in units
	space_above  float64 // add extra padding above
	line_height  float64 // internal wrapping override
	trail_height float64 // is this allowed to be the last thing on a page + by what margin?
	para_indent  int     // add first line indentation (by character offset)
}

func set_paper(text string) (lib.Rect, bool) {
	switch homogenise(text) {
	case "a4":
		return *lib.PageSizeA4, true
	case "uslegal", "legal":
		return *lib.PageSizeLegal, true
	case "usletter", "letter":
		return *lib.PageSizeLetter, true
	}
	return *lib.PageSizeLetter, false
}

func set_format(text string) (Format, bool) {
	switch homogenise(text) {
	case "film", "screen", "screenplay":
		return SCREENPLAY, true
	case "stage", "play", "stageplay":
		return STAGEPLAY, true
	case "graphicnovel", "comic":
		return GRAPHIC_NOVEL, true
	case "manuscript", "novel":
		return MANUSCRIPT, true
	case "manuscriptcompact", "novelcompact":
		return MANUSCRIPT_COMPACT, true
	case "document":
		return DOCUMENT, true
	case ".storyboard":
		return STORYBOARD, true
	}
	return SCREENPLAY, false
}

func build_template(config *Config, format Format) *Template {
	output := new(Template)

	output.kind = format

	output.margin_left   = MARGIN_LEFT
	output.margin_right  = MARGIN_RIGHT
	output.margin_top    = MARGIN_TOP
	output.margin_bottom = MARGIN_BOTTOM

	output.paper = config.paper_size

	switch format {
	default:
		default_screenplay(output)

	case STAGEPLAY:
		default_screenplay(output)

		output.types[ACTION]        .width  = INCH * 3.2
		output.types[ACTION]        .margin = INCH * 2.5
		output.types[CHARACTER]     .margin = INCH * 1.2
		output.types[PARENTHETICAL] .margin = INCH
		output.types[DIALOGUE]      .width  = INCH * 4.5
		output.types[LYRIC]         .width  = INCH * 4.5

	case GRAPHIC_NOVEL:
		default_screenplay(output)

		output.types[SECTION]  .skip = false
		output.types[SECTION2] .skip = false
		output.types[SECTION3] .skip = false

	case MANUSCRIPT:
		output.title_page_align  = CENTER
		output.line_height       = PICA
		output.margin_left       = INCH
		output.ignore_whitespace = true

		output.types[ACTION].para_indent = 4
		output.types[ACTION].line_height = PICA * 2

		output.types[SCENE].casing      = UPPERCASE
		output.types[SCENE].space_above = PICA
		output.types[SCENE].width       = INCH * 5

		output.types[TRANSITION].casing  = UPPERCASE
		output.types[TRANSITION].justify = RIGHT

		output.types[CENTERED].justify = CENTER
		output.types[CENTERED].width   = INCH * 5

		output.types[SYNOPSIS].skip  = true
		output.types[SYNOPSIS].style = ITALIC

		output.types[SECTION].justify      = CENTER
		output.types[SECTION].casing       = UPPERCASE
		output.types[SECTION].style        = BOLD
		output.types[SECTION].space_above  = PICA
		output.types[SECTION].trail_height = PICA * 2

		output.types[SECTION2].casing       = UPPERCASE
		output.types[SECTION2].style        = BOLD
		output.types[SECTION2].space_above  = PICA
		output.types[SECTION2].trail_height = PICA * 2

		output.types[SECTION3].style        = BOLD
		output.types[SECTION3].space_above  = PICA
		output.types[SECTION3].trail_height = PICA * 2

	case MANUSCRIPT_COMPACT:
		output.title_page_align = CENTER
		output.line_height      = PICA

		output.types[ACTION].para_indent = 4
		output.types[ACTION].line_height = PICA
		output.types[ACTION].width       = INCH * 3.5

		output.types[SCENE].casing      = UPPERCASE
		output.types[SCENE].space_above = PICA
		output.types[SCENE].width       = INCH * 5

		output.types[TRANSITION].casing  = UPPERCASE
		output.types[TRANSITION].justify = RIGHT

		output.types[CENTERED].justify = CENTER
		output.types[CENTERED].width   = INCH * 3.5

		output.types[SYNOPSIS].skip  = true
		output.types[SYNOPSIS].style = ITALIC

		output.types[SECTION].casing       = UPPERCASE
		output.types[SECTION].style        = BOLD
		output.types[SECTION].space_above  = PICA
		output.types[SECTION].trail_height = PICA * 2
		output.types[SECTION].width        = INCH * 3.5

		output.types[SECTION2].casing       = UPPERCASE
		output.types[SECTION2].style        = BOLD
		output.types[SECTION2].space_above  = PICA
		output.types[SECTION2].trail_height = PICA * 2

		output.types[SECTION3].style        = BOLD
		output.types[SECTION3].space_above  = PICA
		output.types[SECTION3].trail_height = PICA * 2

	case DOCUMENT:
		output.line_height      = PICA
		output.title_page_align = LEFT

		output.types[SCENE].casing       = UPPERCASE
		output.types[SCENE].space_above  = PICA
		output.types[SCENE].trail_height = PICA * 2

		output.types[TRANSITION].casing  = UPPERCASE
		output.types[TRANSITION].justify = RIGHT

		output.types[CENTERED].justify = CENTER
		output.types[CENTERED].width   = INCH * 5

		output.types[SYNOPSIS].style = ITALIC

		output.types[SECTION].casing       = UPPERCASE
		output.types[SECTION].style        = BOLD | UNDERLINE
		output.types[SECTION].space_above  = PICA
		output.types[SECTION].trail_height = PICA * 2

		output.types[SECTION2].casing       = UPPERCASE
		output.types[SECTION2].style        = BOLD
		output.types[SECTION2].trail_height = PICA * 2

		output.types[SECTION3].style        = BOLD
		output.types[SECTION3].trail_height = PICA * 2

	case STORYBOARD:
		output.landscape = true

		// @todo 'update paper'
		config.paper_size.W, config.paper_size.H = config.paper_size.H, config.paper_size.W
		output.margin_right = config.paper_size.W - MARGIN_RIGHT

		default_screenplay(output)

		output.types[ACTION].width              = INCH * 3.5
		output.types[CHARACTER].margin          = INCH
		output.types[DUAL_CHARACTER].margin     = INCH / 2
		output.types[PARENTHETICAL].margin      = INCH * 0.4
		output.types[DUAL_PARENTHETICAL].margin = INCH * 1.4
		output.types[DUAL_PARENTHETICAL].width  = INCH * 2
		output.types[DIALOGUE].margin           = 0
		output.types[DIALOGUE].margin           = INCH * 2.5
		output.types[DUAL_DIALOGUE].margin      = 0
		output.types[DUAL_DIALOGUE].width       = INCH * 1.5
		output.types[LYRIC].margin              = 0
		output.types[TRANSITION].margin         = INCH * 5.5
	}

	// @todo bottom margin should reflect this pre-compute standard?
	output.margin_right  = config.paper_size.W - MARGIN_RIGHT

	if output.header_margin == 0 {
		output.header_margin = PICA * 3
	}
	if output.footer_margin == 0 {
		output.footer_margin = config.paper_size.H - PICA * 3
	}
	if output.center_line == 0 {
		output.center_line = config.paper_size.W / 2
	}

	if output.types[ACTION].width == 0 {
		output.types[ACTION].width = output.margin_right - output.margin_left - PICA
	}

	output.starred_nudge  = 1.2
	output.starred_margin = output.margin_right + PICA * 2

	zero_color := Color{0, 0, 0}
	if output.note_color == zero_color {
		output.note_color = Color{128, 128, 255}
	}
	if output.highlight_color == zero_color {
		output.highlight_color = Color{255, 249, 115}
	}

	for i := range output.types {
		t := &output.types[i]

		if t.line_height == 0 {
			if output.line_height == 0 {
				t.line_height = LINE_HEIGHT
			} else {
				t.line_height = output.line_height
			}
		}
	}

	if config.include_sections {
		output.types[SECTION].skip  = false
		output.types[SECTION2].skip = false
		output.types[SECTION3].skip = false
	}
	if config.include_synopses {
		output.types[SYNOPSIS].skip = false
	}

	return output
}

// the base screenplay template is re-used / built on by
// several of the other templates, so it's here in a
// reusable form
func default_screenplay(output *Template) {
	output.kind             = SCREENPLAY
	output.line_height      = PICA
	output.title_page_align = CENTER

	output.types[SCENE].casing       = UPPERCASE
	output.types[SCENE].space_above  = PICA
	output.types[SCENE].trail_height = PICA * 2

	output.types[CHARACTER].margin       = INCH * 2
	output.types[CHARACTER].trail_height = PICA * 3

	output.types[DUAL_CHARACTER].margin       = INCH / 2
	output.types[DUAL_CHARACTER].trail_height = PICA * 3

	output.types[PARENTHETICAL].margin       = INCH * 1.4
	output.types[PARENTHETICAL].width        = INCH * 2
	output.types[PARENTHETICAL].trail_height = PICA * 2

	output.types[DUAL_PARENTHETICAL].margin       = INCH / 2 - CHAR_WIDTH * 3
	output.types[DUAL_PARENTHETICAL].width        = INCH * 2.5
	output.types[DUAL_PARENTHETICAL].trail_height = PICA * 2

	output.types[DIALOGUE].margin = INCH
	output.types[DIALOGUE].width  = INCH * 3
	output.types[DIALOGUE].trail_height = output.line_height

	output.types[DUAL_DIALOGUE].width = (output.paper.W - output.margin_right - output.margin_left) / 2 - PICA * 2

	output.types[LYRIC].margin = INCH
	output.types[LYRIC].width  = INCH * 3
	output.types[LYRIC].style  = ITALIC

	output.types[DUAL_LYRIC].margin = INCH
	output.types[DUAL_LYRIC].width  = INCH * 2.5
	output.types[DUAL_LYRIC].style  = ITALIC

	output.types[TRANSITION].casing  = UPPERCASE
	output.types[TRANSITION].justify = RIGHT

	output.types[CENTERED].justify = CENTER
	output.types[CENTERED].width   = INCH * 5

	output.types[SYNOPSIS].skip  = true
	output.types[SYNOPSIS].style = ITALIC

	output.types[SECTION].skip         = true
	output.types[SECTION].casing       = UPPERCASE
	output.types[SECTION].style        = BOLD | UNDERLINE
	output.types[SECTION].space_above  = PICA
	output.types[SECTION].trail_height = PICA * 2

	output.types[SECTION2].skip         = true
	output.types[SECTION2].casing       = UPPERCASE
	output.types[SECTION2].style        = BOLD
	output.types[SECTION2].trail_height = PICA * 2

	output.types[SECTION3].skip         = true
	output.types[SECTION3].style        = BOLD
	output.types[SECTION3].trail_height = PICA * 2

	output.dual_right_offset = output.paper.W - output.margin_right - output.types[DUAL_DIALOGUE].width - output.margin_left - PICA
}

func template_entry_parser(template *Template, current Section_Type, line string, line_count int) bool {
	line = strings.ToLower(line)
	ident, w := extract_ident(line)
	line = left_trim(line[w:])

	if r, w := get_rune(line); r == ':' {
		line = left_trim(line[w:])
	} else {
		// @error
		return false
	}

	if current < TYPE_COUNT {
		switch ident {
		case "skip":
			template.types[current].skip = true

		case "style":
			if x, success := set_style(line); success {
				template.types[current].style = x
			} else {
				eprintf("template error: invalid style %q", line)
			}

		case "casing":
			if x, success := set_casing(line); success {
				template.types[current].casing = x
			} else {
				eprintf("template error: line %-3d invalid letter case %q", line_count, line)
			}

		case "justify":
			if x, success := set_alignment(line); success {
				template.types[current].justify = x
			} else {
				eprintf("template error: invalid alignment %q", line)
			}

		case "margin":
			template.types[current].margin = do_maths(template, line)

		case "width":
			template.types[current].width = do_maths(template, line)

		case "space_above":
			template.types[current].space_above = do_maths(template, line)

		case "line_height":
			template.types[current].line_height = do_maths(template, line)

		case "trail_height":
			template.types[current].trail_height = do_maths(template, line)

		case "para_indent":
			template.types[current].para_indent = int(do_maths(template, line))

		default:
			eprintf("template error: line %-3d bad key in template %q", line_count, ident)
		}
		return true
	}

	switch ident {
	case "margin_left":
		template.margin_left = do_maths(template, line)

	case "margin_right":
		template.margin_right = do_maths(template, line)

	case "margin_top":
		template.margin_top = do_maths(template, line)

	case "margin_bottom":
		template.margin_bottom = do_maths(template, line)

	case "line_height":
		template.line_height = do_maths(template, line)

	case "center_line":
		template.center_line = do_maths(template, line)

	case "dual_right_offset":
		template.dual_right_offset = do_maths(template, line)

	case "header_margin":
		template.header_margin = do_maths(template, line)

	case "footer_margin":
		template.footer_margin = do_maths(template, line)

	case "landscape":
		if line == "false" {
			template.landscape = false
		} else {
			template.landscape = true
		}

	case "ignore_whitespace":
		if line == "false" {
			template.ignore_whitespace = false
		} else {
			template.ignore_whitespace = true
		}

	case "text_color":
		if x, success := parse_color(line); success {
			template.text_color = x
		} else {
			eprintf("template error: line %-3d invalid values in colour %q", line_count, line)
		}

	case "note_color":
		if x, success := parse_color(line); success {
			template.note_color = x
		} else {
			eprintf("template error: line %-3d invalid values in colour %q", line_count, line)
		}

	case "highlight_color":
		if x, success := parse_color(line); success {
			template.highlight_color = x
		} else {
			eprintf("template error: line %-3d invalid values in colour %q", line_count, line)
		}

	case "title_page_align":
		if x, success := set_alignment(line); success {
			template.title_page_align = x
		} else {
			eprintf("template error: invalid alignment %q", line)
		}
	}

	return true
}

func vet_template(template *Template) {
	vet_value(template.line_height,       "line_height")
	vet_value(template.margin_left,       "margin_left")
	vet_value(template.margin_right,      "margin_right")
	vet_value(template.margin_top,        "margin_top")
	vet_value(template.margin_bottom,     "margin_bottom")
	vet_value(template.center_line,       "center_line")
	vet_value(template.dual_right_offset, "dual_right_offset")
	vet_value(template.starred_margin,    "starred_margin")
	vet_value(template.starred_nudge,     "starred_nudge")
	vet_value(template.header_margin,     "header_margin")
	vet_value(template.footer_margin,     "footer_margin")

	for item_type, item := range template.types {
		item_type := Section_Type(item_type)
		vet_entry_value(item.casing,       item_type, "casing")
		vet_entry_value(item.justify,      item_type, "justify")
		vet_entry_value(item.margin,       item_type, "margin")
		vet_entry_value(item.width,        item_type, "width")
		vet_entry_value(item.space_above,  item_type, "space_above")
		vet_entry_value(item.line_height,  item_type, "line_height")
		vet_entry_value(item.trail_height, item_type, "trail_height")
		vet_entry_value(item.para_indent,  item_type, "para_indent")
	}
}

func vet_value[V uint8 | int | float64](v V, s string) {
	if v < 0 {
		eprintf("template: %s is negative value", s)
	}
}

func vet_entry_value[V uint8 | int | float64](v V, t Section_Type, s string) {
	if v < 0 {
		eprintf("template: %v.%s is negative value", t, s)
	}
}

func set_style(line string) (Leaf_Type, bool) {
	fields := strings.Fields(line)
	var x Leaf_Type

	for _, word := range fields {
		switch word {
		case "strikeout": x |= STRIKEOUT
		case "underline": x |= UNDERLINE
		case "highlight": x |= HIGHLIGHT
		case "italic":    x |= ITALIC
		case "bold":      x |= BOLD
		case "none":
			return 0, true
		default:
			return 0, false
		}
	}

	return x, true
}

func set_casing(x string) (uint8, bool) {
	switch strings.ToLower(x) {
	case "upper", "uppercase":
		return UPPERCASE, true
	case "lower", "lowercase":
		return LOWERCASE, true
	}
	return NONE, false
}

func set_alignment(x string) (uint8, bool) {
	switch strings.ToLower(x) {
	case "left":
		return LEFT, true
	case "right":
		return RIGHT, true
	case "centre", "center":
		return CENTER, true
	}
	return LEFT, false
}

func parse_color(numbers string) (Color, bool) {
	var c Color

	array := strings.Fields(numbers)

	for i, x := range array {
		n, err := strconv.Atoi(x)
		if err != nil {
			return c, false
		}

		switch i {
		case 0: c.R = uint8(n)
		case 1: c.G = uint8(n)
		case 2: c.B = uint8(n)
		}
	}

	return c, true
}

func string_to_section_type(x string) (Section_Type, bool) {
	switch strings.ToLower(x) {
	case "action":
		return ACTION, true
	case "scene":
		return SCENE, true
	case "character":
		return CHARACTER, true
	case "dual_character":
		return DUAL_CHARACTER, true
	case "parenthetical":
		return PARENTHETICAL, true
	case "dual_parenthetical":
		return DUAL_PARENTHETICAL, true
	case "dialogue":
		return DIALOGUE, true
	case "dual_dialogue":
		return DUAL_DIALOGUE, true
	case "lyric":
		return LYRIC, true
	case "dual_lyric":
		return DUAL_LYRIC, true
	case "transition":
		return TRANSITION, true
	case "synopsis":
		return SYNOPSIS, true
	case "centered":
		return CENTERED, true
	case "section":
		return SECTION, true
	case "section2":
		return SECTION2, true
	case "section3":
		return SECTION3, true
	}
	return WHITESPACE, false
}

func get_template_value(t *Template, name string) float64 {
	n := strings.IndexRune(name, '.')
	if n >= 0 {
		taxonomy, width := extract_ident(name)
		name = name[width:]
		if len(name) > 0 && name[0] != '.' {
			eprintf("bad template parse") // @todo
		}

		tax_type, success := string_to_section_type(taxonomy)
		if !success {
			eprintf("bad template parse on type") // @todo
		}

		name = name[1:]

		switch name {
		case "margin":
			return t.types[tax_type].margin
		case "width":
			return t.types[tax_type].width
		case "space_above":
			return t.types[tax_type].space_above
		case "line_height":
			return t.types[tax_type].line_height
		case "trail_height":
			return t.types[tax_type].trail_height
		case "para_indent":
			return float64(t.types[tax_type].para_indent)
		}

		eprintf("can't do maths on template field %q", name)
		return 0
	}

	switch name {
	case "title_page_align":
		return float64(t.title_page_align)
	case "line_height":
		return t.line_height
	case "margin_left":
		return t.margin_left
	case "margin_right":
		return t.margin_right
	case "margin_top":
		return t.margin_top
	case "margin_bottom":
		return t.margin_bottom
	case "center_line":
		return t.center_line
	case "dual_right_offset":
		return t.dual_right_offset
	case "starred_margin":
		return t.starred_margin
	case "starred_nudge":
		return t.starred_nudge
	case "header_margin":
		return t.header_margin
	case "footer_margin":
		return t.footer_margin

	case "pica":
		return PICA
	case "inch":
		return INCH
	case "char_width":
		return CHAR_WIDTH
	case "paper_width":
		return t.paper.W
	case "paper_height":
		return t.paper.H
	}

	eprintf("can't do maths on template field %q", name)
	return 0
}

const (
	OPERAND_VALUE uint8 = iota
	OPERATION_ADD
	OPERATION_SUB
	OPERATION_MUL
	OPERATION_DIV
	PARENS_OPEN
	PARENS_CLOSE
)

type Operation struct {
	kind     uint8
	priority uint8
	value    float64
}

func extract_dotted_ident(input string) (string, int) {
	width := 0
	for _, c := range input {
		if !(unicode.IsLetter(c) || c == '_' || c == '.') {
			return input[:width], width
		}
		width += utf8.RuneLen(c)
	}
	return input, width
}

func extract_dotted_number(input string) (string, int) {
	width := 0
	for _, c := range input {
		if !(unicode.IsNumber(c) || c == '.') {
			return input[:width], width
		}
		width += utf8.RuneLen(c)
	}
	return input, width
}

func do_maths(t *Template, text string) float64 {
	operations := make([]Operation, 0, 32)

	for {
		text = left_trim(text)
		if len(text) == 0 {
			break
		}

		char, char_width := get_rune(text)

		var op Operation

		switch char {
		case '(', '[':
			op.kind = PARENS_OPEN
			text    = text[char_width:]
		case ')', ']':
			op.kind = PARENS_CLOSE
			text    = text[char_width:]
		case '-':
			op.kind = OPERATION_SUB
			text    = text[char_width:]
		case '+':
			op.kind = OPERATION_ADD
			text    = text[char_width:]
		case '*', '×':
			op.kind     = OPERATION_MUL
			op.priority = 1
			text        = text[char_width:]
		case '/', '÷':
			op.kind     = OPERATION_DIV
			op.priority = 1
			text        = text[char_width:]
		default:
			if unicode.IsNumber(char) {
				number, number_width := extract_dotted_number(text)
				op.value = to_float(number)
				text = text[number_width:]

			} else if unicode.IsLetter(char) {
				ident, ident_width := extract_dotted_ident(text)
				op.value = get_template_value(t, ident)
				text = text[ident_width:]
			}
		}

		operations = append(operations, op)
	}

	operations = shunting_yard(operations)
	stack := make([]Operation, 0, len(operations))

	for _, token := range operations {
		if token.kind == OPERAND_VALUE {
			stack = append(stack, token)
			continue
		}

		a := stack[len(stack) - 2].value
		b := stack[len(stack) - 1].value

		stack = stack[:len(stack) - 2]

		result := float64(0)

		switch token.kind {
		case OPERATION_SUB:
			result = a - b
		case OPERATION_ADD:
			result = a + b
		case OPERATION_DIV:
			result = a / b
		case OPERATION_MUL:
			result = a * b
		}

		stack = append(stack, Operation{value: result})
	}

	return stack[0].value
}

func shunting_yard(operations []Operation) []Operation {
	final     := make([]Operation, 0, len(operations))
	operators := make([]Operation, 0, len(operations))

	for _, v := range operations {
		if v.kind == OPERAND_VALUE {
			final = append(final, v)
		} else {
			if v.kind == PARENS_OPEN {
				operators = append(operators, v)

			} else if v.kind == PARENS_CLOSE {
				found_left := false

				for len(operators) > 0 {
					o := operators[len(operators) - 1]
					operators = operators[:len(operators) - 1]

					if o.kind == PARENS_OPEN {
						found_left = true
						break
					} else {
						final = append(final, o)
					}
				}

				if !found_left {
					eprintln("mismatched parentheses in expression")
					return nil
				}

			} else {
				for len(operators) > 0 {
					top := operators[len(operators)-1]

					if top.kind == PARENS_OPEN { break }

					if v.priority <= top.priority {
						operators = operators[:len(operators) - 1]
						final = append(final, top)
					} else {
						break
					}
				}

				operators = append(operators, v)
			}
		}
	}

	for len(operators) > 0 {
		operator := operators[len(operators) - 1]
		operators = operators[:len(operators) - 1]
		final = append(final, operator)
	}

	return final
}

func to_float(s string) float64 {
	n, e := strconv.ParseFloat(s, 64)
	if e != nil {
		return 0
	}
	return n
}
