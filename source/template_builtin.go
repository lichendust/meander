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

// @todo
const (
	dual_width       = inch * 2.5
	dual_char_margin = inch / 2
	dual_para_margin = inch / 2
)

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
		},
	},
	"manuscript": {
		line_height_title: line_height_title,
		line_height:       pica * 1.5,

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
		},
	},
}