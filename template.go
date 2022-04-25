package main

import "github.com/signintech/gopdf"

const (
	more_tag   = "(more)"
	cont_tag   = "(CONT'D)"
	cont_check = "(CONT" // hrgh
)

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

	header_height = margin_top    - pica * 3
	footer_height = margin_bottom - pica * 2
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

// this method should allow for easy language
// expansion later

// the parser sets userdata to lowercase for
// comparison, so all declarations in here
// should be lowercase.
var valid_scene = map[string]bool {
	"int":     true,
	"ext":     true,
	"int/ext": true,
	"ext/int": true,
	"i/e":     true,
	"e/i":     true,
	"est":     true,
	"scene":   true,
}

var valid_title_page = map[string]bool {
	// fountain
	"title":      true,
	"credit":     true,
	"author":     true,
	"source":     true,
	"contact":    true,
	"revision":   true,
	"copyright":  true,
	"draft date": true,
	"notes":      true,

	// meander
	"format": true,
	"paper":  true,
}