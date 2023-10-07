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

import lib "github.com/signintech/gopdf"

const (
	FONT_SIZE = 12

	PICA float64 = 12
	INCH float64 = PICA * 6

	LINE_HEIGHT = PICA

	MARGIN_TOP    = INCH * 1.1
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
	DOCUMENT
)

type Template struct {
	kind       Format
	landscape  bool

	title_page_align uint8

	space_above float64
	line_height float64

	left_margin  float64
	right_margin float64
	center_line  float64

	header_height float64
	footer_height float64

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
	margin  float64   // bump up left or right margin

	// 0 = ignored
	width        float64 // max width of wrap in units
	space_above  float64 // add extra padding above
	line_height  float64 // internal wrapping override
	trail_height float64 // is this allowed to be the last thing on a page + by what margin?
	para_indent  int     // add first line indentation (by character offset)
}

func set_paper(text string) *lib.Rect {
	switch homogenise(text) {
	case "a4":
		return lib.PageSizeA4
	case "usletter", "letter":
		return lib.PageSizeLetter
	case "uslegal", "legal":
		return lib.PageSizeLegal
	}
	return nil
}

func is_valid_format(text string) (Format, bool) {
	switch homogenise(text) {
	case "film", "screen", "screenplay":
		return SCREENPLAY, true
	case "stage", "play", "stageplay":
		return STAGEPLAY, true
	case "graphicnovel", "comic":
		return GRAPHIC_NOVEL, true
	case "manuscript", "novel":
		return MANUSCRIPT, true
	case "document":
		return DOCUMENT, true
	}
	return 0, false
}

func default_screenplay(paper lib.Rect) *Template {
	base := Template{
		kind:             SCREENPLAY,
		line_height:      PICA,
		title_page_align: CENTER,

		types: [TYPE_COUNT]Template_Entry{
			ACTION: {
				width: paper.W - MARGIN_LEFT - MARGIN_RIGHT - PICA,
			},
			SCENE: {
				casing:       UPPERCASE,
				style:        UNDERLINE,
				space_above:  PICA,
				trail_height: PICA * 2,
			},
			CHARACTER: {
				margin:       INCH * 2,
				trail_height: PICA * 3,
			},
			DUAL_CHARACTER: {
				margin:       INCH / 2,
				trail_height: PICA * 3,
			},
			PARENTHETICAL: {
				margin:       INCH * 1.4,
				width:        INCH * 2,
				trail_height: PICA * 2,
			},
			DUAL_PARENTHETICAL: {
				margin:       INCH / 2 - CHAR_WIDTH * 3,
				width:        INCH * 2.5,
				trail_height: PICA * 2,
			},
			DIALOGUE: {
				margin: INCH,
				width:  INCH * 3,
			},
			DUAL_DIALOGUE: {
				width:  INCH * 2.5,
			},
			LYRIC: {
				margin: INCH,
				width:  INCH * 3,
				style:  ITALIC,
			},
			DUAL_LYRIC: {
				margin: INCH,
				width:  INCH * 2.5,
				style:  ITALIC,
			},
			TRANSITION: {
				casing:  UPPERCASE,
				justify: RIGHT,
			},
			CENTERED: {
				justify: CENTER,
				width:   INCH * 5,
			},
			SYNOPSIS: {
				skip:  true,
				style: ITALIC,
			},
			SECTION: {
				skip:         true,
				casing:       UPPERCASE,
				style:        BOLD | UNDERLINE,
				space_above:  PICA,
				trail_height: PICA * 2,
			},
			SECTION2: {
				skip:         true,
				casing:       UPPERCASE,
				style:        BOLD,
				trail_height: PICA * 2,
			},
			SECTION3: {
				skip:         true,
				style:        BOLD,
				trail_height: PICA * 2,
			},
		},
	}
	return &base
}

func build_template(config *Config, format Format, paper lib.Rect) (output *Template) {
	switch format {
	default:
		output = default_screenplay(paper)

	case STAGEPLAY:
		output = default_screenplay(paper)

		output.types[ACTION]        .width  = INCH * 3.2
		output.types[ACTION]        .margin = INCH * 2.5
		output.types[CHARACTER]     .margin = INCH * 1.2
		output.types[PARENTHETICAL] .margin = INCH
		output.types[DIALOGUE]      .width  = INCH * 4.5
		output.types[LYRIC]         .width  = INCH * 4.5

	case GRAPHIC_NOVEL:
		output = default_screenplay(paper)

		output.types[SECTION] .skip = false
		output.types[SECTION2].skip = false
		output.types[SECTION3].skip = false

	case MANUSCRIPT:
		manuscript := Template{
			title_page_align: CENTER,
			line_height: PICA,

			types: [TYPE_COUNT]Template_Entry{
				ACTION: {
					para_indent: 4,
					line_height: PICA * 2,
					space_above: PICA,
					width:       paper.W - MARGIN_LEFT - MARGIN_RIGHT - PICA,
				},
				SCENE: {
					casing:      UPPERCASE,
					space_above: PICA,
					width:       INCH * 5,
				},
				TRANSITION: {
					casing:  UPPERCASE,
					justify: RIGHT,
				},
				CENTERED: {
					justify: CENTER,
					width:   INCH * 5,
				},
				SYNOPSIS: {
					skip:  true,
					style: ITALIC,
				},
				SECTION: {
					justify:      CENTER,
					casing:       UPPERCASE,
					style:        BOLD,
					space_above:  PICA,
					trail_height: PICA * 2,
				},
				SECTION2: {
					casing:       UPPERCASE,
					style:        BOLD,
					space_above:  PICA,
					trail_height: PICA * 2,
				},
				SECTION3: {
					style:        BOLD,
					space_above:  PICA,
					trail_height: PICA * 2,
				},
			},
		}

		output = &manuscript

	case DOCUMENT:
		document := Template{
			line_height:      PICA,
			title_page_align: LEFT,

			types: [TYPE_COUNT]Template_Entry{
				ACTION: {
					width: paper.W - MARGIN_LEFT - MARGIN_RIGHT - PICA,
				},
				SCENE: {
					casing:       UPPERCASE,
					style:        UNDERLINE,
					space_above:  PICA,
					trail_height: PICA * 2,
				},
				TRANSITION: {
					casing:  UPPERCASE,
					justify: RIGHT,
				},
				CENTERED: {
					justify: CENTER,
					width:   INCH * 5,
				},
				SYNOPSIS: {
					style: ITALIC,
				},
				SECTION: {
					casing:       UPPERCASE,
					style:        BOLD | UNDERLINE,
					space_above:  PICA,
					trail_height: PICA * 2,
				},
				SECTION2: {
					casing: UPPERCASE,
					style:  BOLD,
					trail_height: PICA * 2,
				},
				SECTION3: {
					style: BOLD,
					trail_height: PICA * 2,
				},
			},
		}

		output = &document
	}

	zero_color := Color{0, 0, 0}
	if output.note_color == zero_color {
		output.note_color = Color{128, 128, 255}
	}
	if output.highlight_color == zero_color {
		output.highlight_color = Color{255, 249, 115}
	}

	for i := range output.types {
		t := &output.types[i]

		if t.space_above == 0 {
			t.space_above = output.space_above
		}

		if t.line_height == 0 {
			if output.line_height == 0 {
				t.line_height = LINE_HEIGHT
			} else {
				t.line_height = output.line_height
			}
		}
	}

	if output.header_height == 0 {
		output.header_height = MARGIN_TOP - PICA * 3
	}
	if output.footer_height == 0 {
		output.footer_height = paper.H - MARGIN_BOTTOM + PICA * 2
	}

	if output.left_margin == 0 {
		output.left_margin = MARGIN_LEFT
	}
	if output.right_margin == 0 {
		output.right_margin = paper.W - MARGIN_RIGHT
	}
	if output.center_line == 0 {
		output.center_line = paper.W / 2
	}

	if config.include_sections {
		output.types[SECTION].skip = false
		output.types[SECTION2].skip = false
		output.types[SECTION3].skip = false
	}
	if config.include_synopses {
		output.types[SYNOPSIS].skip = false
	}

	output.kind = format

	return output
}