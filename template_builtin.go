package main

import "github.com/signintech/gopdf"

var paper_store = map[string]*paper {
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
	dual_width = inch * 2.5

	dual_char_margin = inch / 2
	dual_para_margin = inch / 2
)

var template_store = map[string]*template {
	"screenplay": {
		line_height_title: line_height_title,
		line_height:       line_height,

		title_page_align: CENTER,

		allow_dual_dialogue: true,

		look_up: map[uint8]*template_entry {
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
			SECTION:  {skip: true},
			SECTION2: {skip: true},
			SECTION3: {skip: true},
		},
	},
	"stageplay": {
		line_height_title: line_height_title,
		line_height:       line_height,

		title_page_align: CENTER,

		allow_dual_dialogue: true,

		look_up: map[uint8]*template_entry {
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
			SECTION:  {skip: true},
			SECTION2: {skip: true},
			SECTION3: {skip: true},
		},
	},
	"graphicnovel": {
		line_height_title: line_height_title,
		line_height:       line_height,

		title_page_align: CENTER,

		allow_dual_dialogue: true,

		look_up: map[uint8]*template_entry {
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
		line_height:       line_height,

		title_page_align: LEFT,

		allow_dual_dialogue: false,

		look_up: map[uint8]*template_entry {
			ACTION: {
				para_indent: 4,
				line_height: pica * 2,
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

		look_up: map[uint8]*template_entry {
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
				style: BOLD | UNDERLINE,
				space_above: pica,
			},
			SECTION2: {
				style: UNDERLINE,
				space_above: pica,
			},
			SECTION3: {
				style: ITALIC | UNDERLINE,
				space_above: pica,
			},
		},
	},
}