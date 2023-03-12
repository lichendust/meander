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

import "github.com/signintech/gopdf"

const (
	font_size int = 12

	pica float64 = 12
	inch float64 = 6 * pica

	line_height       = pica
	line_height_title = pica * 1.2

	margin_top    = inch * 1.1
	margin_left   = inch * 1.5
	margin_right  = inch
	margin_bottom = inch

	header_height = margin_top - pica * 3
	footer_height = margin_bottom - pica * 2

	// hard-coded now because we stopped
	// supporting user font selection
	char_width float64 = 7.188
)

// justifications
const (
	LEFT uint8 = iota
	RIGHT
	CENTER
)

// character casing
const (
	NONE uint8 = iota
	UPPERCASE
	LOWERCASE
)

type paper struct {
	paper_data *gopdf.Rect
	width      float64
	height     float64
}

type template struct {
	line_height_title float64
	line_height       float64

	title_page_align    uint8
	allow_dual_dialogue bool

	look_up map[uint8]*template_entry
}

type template_entry struct {
	skip bool // ignore in render

	casing  uint8   // force upper/lower case
	style   int     // force style override (stacks)
	justify uint8   // align to left or right margin
	margin  float64 // bump up left or right margin

	width       float64 // max width of line wrap        (0 = ignore)
	space_above float64 // add extra padding above
	line_height float64 // internal wrapping override    (0 = ignore)
	para_indent int     // add first line indentation    (0 = ignore)
}



// @todo hardcoded
const (
	dual_width       = inch * 2.5
	dual_char_margin = inch / 2
	dual_para_margin = inch / 2
)

var paper_store = map[string]*paper{
	"a4": {
		paper_data: gopdf.PageSizeA4,
		width:      inch * 8.27,
		height:     inch * 11.69,
	},
	"usletter": {
		paper_data: gopdf.PageSizeLetter,
		width:      inch * 8.5,
		height:     inch * 11,
	},
}

var template_store = map[string]*template{
	"screenplay": {
		line_height_title: line_height_title,
		line_height:       line_height,

		title_page_align: CENTER,

		allow_dual_dialogue: true,

		look_up: map[uint8]*template_entry{
			ACTION: {},
			SCENE: {
				casing:      UPPERCASE,
				style:       UNDERLINE,
				space_above: pica,
				width:       inch * 5,
			},
			CHARACTER: {
				margin: inch * 2,
			},
			PARENTHETICAL: {
				margin: inch * 1.4,
				width:  inch * 2,
			},
			DIALOGUE: {
				margin: inch * 1,
				width:  inch * 3,
			},
			LYRIC: {
				margin: inch * 1,
				width:  inch * 3,
				style:  ITALIC,
			},
			TRANSITION: {
				casing:  UPPERCASE,
				justify: RIGHT,
			},
			CENTERED: {
				justify: CENTER,
				width:   inch * 5,
			},
			SYNOPSIS: {
				skip:  true,
				style: ITALIC,
			},
			SECTION: {
				skip:        true,
				casing:      UPPERCASE,
				style:       BOLD | UNDERLINE,
				space_above: pica,
			},
			SECTION2: {
				skip:   true,
				casing: UPPERCASE,
				style:  BOLD,
			},
			SECTION3: {
				skip:  true,
				style: BOLD,
			},
			LIST: {
				margin: char_width * 2,
			},
		},
	},
	"stageplay": {
		line_height_title: line_height_title,
		line_height:       line_height,

		title_page_align: CENTER,

		allow_dual_dialogue: true,

		look_up: map[uint8]*template_entry{
			ACTION: {
				margin: inch * 2.5,
				width:  inch * 3.2,
			},
			SCENE: {
				margin:      -inch / 4,
				casing:      UPPERCASE,
				style:       BOLD,
				space_above: pica,
				width:       inch * 5,
			},
			CHARACTER: {
				margin: inch * 1.2,
			},
			PARENTHETICAL: {
				margin: inch,
				width:  inch * 2,
			},
			DIALOGUE: {
				width: inch * 4.5,
			},
			LYRIC: {
				width: inch * 4.5,
				style: ITALIC,
			},
			TRANSITION: {
				casing:  UPPERCASE,
				justify: RIGHT,
			},
			CENTERED: {
				justify: CENTER,
				width:   inch * 5,
			},
			SYNOPSIS: {
				skip:  true,
				style: ITALIC,
			},
			SECTION: {
				skip:        true,
				casing:      UPPERCASE,
				style:       BOLD | UNDERLINE,
				space_above: pica,
			},
			SECTION2: {
				skip:   true,
				casing: UPPERCASE,
				style:  BOLD,
			},
			SECTION3: {
				skip:  true,
				style: BOLD,
			},
			LIST: {
				margin: char_width * 2,
			},
		},
	},
	"graphicnovel": {
		line_height_title: line_height_title,
		line_height:       line_height,

		title_page_align: CENTER,

		allow_dual_dialogue: true,

		look_up: map[uint8]*template_entry{
			ACTION: {},
			SCENE: {
				casing:      UPPERCASE,
				style:       UNDERLINE,
				space_above: pica,
				width:       inch * 5,
			},
			CHARACTER: {
				margin: inch * 2,
			},
			PARENTHETICAL: {
				margin: inch * 1.4,
				width:  inch * 2,
			},
			DIALOGUE: {
				margin: inch * 1,
				width:  inch * 3,
			},
			LYRIC: {
				margin: inch * 1,
				width:  inch * 3,
				style:  ITALIC,
			},
			TRANSITION: {
				casing:  UPPERCASE,
				justify: RIGHT,
			},
			CENTERED: {
				justify: CENTER,
				width:   inch * 5,
			},
			SYNOPSIS: {
				skip:  true,
				style: ITALIC,
			},
			SECTION: {
				casing:      UPPERCASE,
				style:       BOLD | UNDERLINE,
				space_above: pica,
			},
			SECTION2: {
				casing: UPPERCASE,
				style:  BOLD,
			},
			SECTION3: {
				style: BOLD,
			},
			LIST: {
				margin: char_width * 2,
			},
		},
	},
	"manuscript": {
		line_height_title: line_height_title,
		line_height:       pica * 2,

		title_page_align: CENTER,

		allow_dual_dialogue: false,

		look_up: map[uint8]*template_entry{
			ACTION: {
				para_indent: 4,
			},
			SCENE: {
				casing:      UPPERCASE,
				space_above: pica,
				width:       inch * 5,
			},
			CHARACTER: {
				margin: inch * 2,
			},
			PARENTHETICAL: {
				margin: inch * 1.4,
				width:  inch * 2,
			},
			DIALOGUE: {
				margin: inch * 1,
				width:  inch * 3,
			},
			LYRIC: {
				margin: inch * 1,
				width:  inch * 3,
				style:  ITALIC,
			},
			TRANSITION: {
				casing:  UPPERCASE,
				justify: RIGHT,
			},
			CENTERED: {
				justify: CENTER,
				width:   inch * 5,
			},
			SYNOPSIS: {
				skip:  true,
				style: ITALIC,
			},
			SECTION: {
				justify:     CENTER,
				casing:      UPPERCASE,
				style:       BOLD,
				space_above: pica,
			},
			SECTION2: {
				casing:      UPPERCASE,
				style:       BOLD,
				space_above: pica,
			},
			SECTION3: {
				style:       BOLD,
				space_above: pica,
			},
			LIST: {
				margin: char_width * 2,
			},
		},
	},
	"document": {
		line_height_title: line_height_title,
		line_height:       line_height,

		allow_dual_dialogue: true,

		look_up: map[uint8]*template_entry{
			ACTION: {},
			SCENE: {
				casing:      UPPERCASE,
				space_above: pica,
				width:       inch * 5,
			},
			CHARACTER: {
				margin: inch * 2,
			},
			PARENTHETICAL: {
				margin: inch * 1.4,
				width:  inch * 2,
			},
			DIALOGUE: {
				margin: inch * 1,
				width:  inch * 3,
			},
			LYRIC: {
				margin: inch * 1,
				width:  inch * 3,
				style:  ITALIC,
			},
			TRANSITION: {
				casing:  UPPERCASE,
				justify: RIGHT,
			},
			CENTERED: {
				justify: CENTER,
				width:   inch * 5,
			},
			SYNOPSIS: {
				skip:  true,
				style: ITALIC,
			},
			SECTION: {
				style:       BOLD | UNDERLINE,
				space_above: pica,
			},
			SECTION2: {
				style:       UNDERLINE,
				space_above: pica,
			},
			SECTION3: {
				style:       ITALIC | UNDERLINE,
				space_above: pica,
			},
			LIST: {
				margin: char_width * 2,
			},
		},
	},
}