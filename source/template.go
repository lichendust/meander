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
	FONT_SIZE int = 12

	PICA float64 = 12
	INCH float64 = 6 * PICA

	LINE_HEIGHT = PICA

	MARGIN_TOP    = INCH * 1.1
	MARGIN_LEFT   = INCH * 1.5
	MARGIN_RIGHT  = INCH
	MARGIN_BOTTOM = INCH

	MARGIN_HEADER = MARGIN_TOP    - PICA * 3
	MARGIN_FOOTER = MARGIN_BOTTOM - PICA * 2
)

const (
	LEFT uint8 = iota
	RIGHT
	CENTER
)

const (
	NONE uint8 = iota
	UPPERCASE
	LOWERCASE
)

type Template struct {
	SpaceAbove float64
	SpaceBelow float64
	LineHeight float64

	TitlePageAlign uint8

	AllowDual      bool
	DualWidth      float64
	DualCharMargin float64
	DualParaMargin float64

	Types [TYPE_COUNT]Type_Template
}

type Type_Template struct {
	Skip bool

	Style   int     // force style override (bitwise)
	Casing  uint8   // force upper/lower case
	Justify uint8   // align to left or right margin
	Margin  float64 // bump up left or right margin

	// 0 = ignored
	Width           float64 // max width of line wrap
	SpaceAbove      float64 // add extra padding above
	SpaceBelow      float64 // add extra padding below
	LineHeight      float64 // internal wrapping override
	ParagraphIndent int     // add first line indentation
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
	return lib.PageSizeA4 // default
}

func set_template(text string) *Template {
	/*switch homogenise(text) {
	case "film", "screen", "screenplay":
		return &SCREENPLAY
	case "stage", "play", "stageplay":
		return &STAGEPLAY
	}*/
	return &SCREENPLAY // default
}

var SCREENPLAY = Template{
	SpaceAbove: 0,
	SpaceBelow: PICA,
	LineHeight: PICA,

	TitlePageAlign: CENTER,

	AllowDual:      true,
	DualWidth:      2.5 * INCH,
	DualCharMargin: INCH / 2,
	DualParaMargin: INCH / 2,

	Types: [TYPE_COUNT]Type_Template{
		SCENE: {
			Casing:     UPPERCASE,
			Style:      UNDERLINE,
			SpaceAbove: PICA,
			Width:      INCH * 5,
		},
		CHARACTER: {
			Margin: INCH * 2,
		},
		PARENTHETICAL: {
			Margin: INCH * 1.4,
			Width:  INCH * 2,
		},
		DIALOGUE: {
			Margin: INCH * 1,
			Width:  INCH * 3,
		},
		LYRIC: {
			Margin: INCH * 1,
			Width:  INCH * 3,
			Style:  ITALIC,
		},
		TRANSITION: {
			Casing:  UPPERCASE,
			Justify: RIGHT,
		},
		JUSTIFY_CENTER: {
			Justify: CENTER,
			Width:   INCH * 5,
		},
		SYNOPSIS: {
			Skip:  true,
			Style: ITALIC,
		},
		SECTION: {
			Skip:       true,
			Casing:     UPPERCASE,
			Style:      BOLD | UNDERLINE,
			SpaceAbove: PICA,
		},
		SECTION2: {
			Skip:   true,
			Casing: UPPERCASE,
			Style:  BOLD,
		},
		SECTION3: {
			Skip:  true,
			Style: BOLD,
		},
		LIST: {
			Margin: CHAR_WIDTH * 2,
		},
	},
}