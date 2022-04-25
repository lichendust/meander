package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// bitwise flags for setting styles. we don't
// actually use them as such in the leaves,
// (we just have an array of leaves with one
// flag a leaf for each line) but the flag
// mode is used in the style-overriding by
// the templating

// it's not the most sensible way of doing
// things, but the leaf parser came before
// the bit-shifting so that's how we do it
const (
	NORMAL int = 1 << iota
	ITALIC
	BOLD
	BOLDITALIC
	UNDERLINE
	STRIKEOUT
	HIGHLIGHT
	NOTE
	ESCAPE
)

/*
	cheat sheet
	set     value |  flag
	clear   value &^ flag
	toggle  value ^  flag
	has     value &  flag != 0
*/

// container for each line of text on the page
type syntax_line struct {
	// exact width of the final line,
	// distinct from the wrap width
	length     int

	// array of text data — only "normal"
	// leaves **ever** go to print, so
	// anything in between is a
	// defining "switch" that does not
	// contain anything
	leaves     []*syntax_leaf

	// ranges for each block-draw element, in
	// the form [start, end, start, end]
	// where each int is a character position
	underline  []int
	strikeout  []int
	highlight  []int

	// used to force a stylistic override on
	// the entire line. effectively changes
	// the "baseline" style (bold, italics)
	// that we reset to
	// gopdf uses a horrid "BI" string to define
	// which font variant to use, despite using
	// very nice preset style enums to register
	// the fonts in the first place.  weird.
	font_reset string
}

// a chunk of text _or_ formatter switch
// depending on whether its leaf_type is
// "normal" or not
type syntax_leaf struct {
	// "normal" leaf_type is the only one
	//  that contains text to be printed
	leaf_type   int

	// whether the formatter leaf is the
	// start or end of a range e.g:
	// (bold-begin or bold-end)
	// guaranteed to be one or the other
	// by the process of their creation
	opening     bool

	// the content of leaf, only ever
	// used when the leaf_type is "normal"
	text        string
}

// intermediary structure that is effectively
// a proto-syntax_leaf, used to figuring out
// what's going on in the string
type inline_format struct {
	format_type  int
	space_width  int
	text_width   int
	could_open   bool
	could_close  bool
	is_confirmed bool
	space_only   bool
	text         string
}

// valid formatters to be considered. used for
// ignoring them across all parsers like
// allowing **int. scene** to be a valid
// scene, even though it's got stars in the
// way
var format_chars = map[rune]bool {
	'*':  true,
	'+':  true,
	'~':  true,
	'\\': true,
	'_':  true,
	']':  true,
	'[':  true,
}

// forcing function that accepts a single line and overwrites
// its pre-calculated syntax to a new baseline while letting
// non-colliding syntaxes still shine through
/*
	given the string:
		"this is a line *with italics* and +highlights+."

	we could use this function, passing "BOLD" to it, to
	forcibly make the entire string bold -- critically,
	though, the italics will still be italics (just bold
	too) and the highlight will still appear.
*/
func syntax_line_override(line *syntax_line, style int) {
	// if the input style has a range formatter, we
	// just replace all ranges with one full-width one
	// the whole line is then highlighted, etc.
	if style & UNDERLINE != 0 {
		line.underline = []int{0, line.length}
	}
	if style & STRIKEOUT != 0 {
		line.strikeout = []int{0, line.length}
	}
	if style & HIGHLIGHT != 0 {
		line.highlight = []int{0, line.length}
	}

	// remember that aside about gopdf's weird string
	// up above in the syntax_line struct?
	// here you go:
	if style & BOLD != 0 {
		line.font_reset += "B"
	}
	if style & ITALIC != 0 {
		line.font_reset += "I"
	}
	if style & BOLDITALIC != 0 {
		line.font_reset += "BI"
	}
}

// the monster herself.
// takes a string, a wrap width (and a first line indentation)
// and outputs an array of wrapped lines containing arrays
// of formatted text leaves
func syntax_leaf_parser(input string, max_width, para_indent int) []*syntax_line {
	the_list := make([]*inline_format, 0, 16)

	for {
		if len(input) == 0 {
			break
		}

		space_width := count_whitespace(input)
		input = consume_whitespace(input)

		if len(input) == 0 {
			break
		}

		the_word := ""
		rune_width := 0
		is_type := NORMAL

		// @todo handle escape characters!!!

		switch input[0] {
		case '\\':
			the_word, rune_width = extract_repeated_rune(input, '\\')
			is_type = ESCAPE

		case '*':
			the_word, rune_width = extract_repeated_rune(input, '*')

			switch rune_width {
			case 1: is_type = ITALIC
			case 2: is_type = BOLD
			case 3: is_type = BOLDITALIC
			}

		case '_':
			the_word, rune_width = extract_repeated_rune(input, '_')
			is_type = UNDERLINE

			if rune_width > 1 {
				is_type = NORMAL // @todo nope nope nope
			}

		case '+':
			the_word, rune_width = extract_repeated_rune(input, '+')
			is_type = HIGHLIGHT

			if rune_width > 1 {
				is_type = NORMAL // @todo nope nope nope
			}

		case '~':
			the_word, rune_width = extract_repeated_rune(input, '~')
			is_type = STRIKEOUT

			if rune_width != 2 {
				is_type = NORMAL // @todo nope nope nope
			}

		case '[':
			the_word, rune_width = extract_repeated_rune(input, '[')
			is_type = NOTE

			if rune_width < 2 {
				is_type = NORMAL
			}

		case ']':
			the_word, rune_width = extract_repeated_rune(input, ']')
			is_type = NOTE

			if rune_width < 2 {
				is_type = NORMAL
			}

		default:
			the_word, rune_width = formatter_tokens(input)
			is_type = NORMAL
		}

		word_width := count_all_runes(the_word)

		format := &inline_format {
			format_type: is_type,
			space_width: space_width,
			text_width:  word_width,
			text:        the_word,
		}

		if is_type == ESCAPE {
			format.space_only = true
		}

		the_list = append(the_list, format)

		input = input[rune_width:]
	}

	// decide whether any given token can be
	// an opener, closer or ambiguous
	{
		length := len(the_list) - 1

		for i, entry := range the_list {
			if entry.format_type == NORMAL {
				continue
			}

			if entry.format_type == NOTE {
				entry.is_confirmed = (entry.text[0] == '[')
				continue
			}

			// handle escape sequences
			if entry.format_type == ESCAPE {
				n := len(entry.text)

				/*
					there's a looming set of @todo-s at the top
					of this function for how formatters are
					counted, which are _definitely_
					manifestations of a bug, but in all the
					writing I have personally used Meander for,
					they've just never come up, so I haven't
					dealt with it yet like a total *****

					anyway, this is the example for how they
					_should_ be dealt with, splitting them up
					and having some act as text and others act
					as formatters -- not exactly the same in all
					cases, but as a general idea, this is good
					stuff:
				*/

				// if the number of backslashes is even
				// they're all escaped, so set them "normal"
				// and halve the length
				if n % 2 == 0 {
					entry.format_type = NORMAL
					entry.space_only  = false

					entry.text = entry.text[n / 2:]
					continue
				}

				// otherwise, if they're odd but > 1, do the same
				// but subtract the final odd-one-out and apply it
				// as a legit escape
				if n > 1 {
					entry.format_type = NORMAL
					entry.space_only  = false

					entry.text = entry.text[(n + 1) / 2 - 1:]
				}

				// lookahead and escape the next item as applicable
				if len(the_list[i:]) > 1 {
					target := the_list[i + 1:][0]
					if target.format_type != NORMAL {
						target.format_type = NORMAL
					}
				}
				continue
			}

			// if spaces are before _and_ after
			// we're normal by default
			has_spaces_before := false

			// if the leading spaces are zero,
			// we can definitely be a closer
			if entry.space_width == 0 {
				entry.could_close = true
			} else {
				has_spaces_before = true
			}

			// if we're the last entry we _could never_ open,
			// because it's the end of the damn string
			test := i + 1

			if test > length { // can't open
				continue
			}

			// if the next chunk has no spaces before it
			// we could open (we're smushed against it)
			if the_list[test].space_width == 0 {
				entry.could_open = true
			} else if has_spaces_before {
				// if we're "floating", with spaces before
				// and after, we're not a formatter - reset
				entry.format_type = NORMAL
			}

			/*
				it very rarely comes up, but we might need to
				also check the type of each of these
				lookaheads/lookbehinds because we're ___very
				rarely___ making bogus choices: sometimes there
				are no spaces only due to the neighbouring
				chunk also being a formatter, which should be
				ignored.

				it works out 99% of the time just by virtue of
				the way people write formatters naturally, but
				it's _there_.

				the issue is we haven't firmly settled on any
				types at this stage, so using that check kinda
				has to be a guess.  this guess just sometimes
				propagates through all the next steps

				you can see it here:

					this *is an italic section _*and then an underline_.

				"""technically""" the italics should not count
				because they violate the rules of Fountain. If
				the paired _ and * were swapped, the system
				correctly ignores the two stars, but does
				not in their above form.

				it's a side effect of the guesswork-level
				type-checking we have at this stage, which
				is then baked in hard because the balancing
				loop later is biased -- you can see it
				in the very next block comment in the next
				section, showing the leftward bias
			*/
		}
	}

	// balance everything we've found
	{
		italic_on    := false
		bold_on      := false
		highlight_on := false
		underline_on := false
		strike_on    := false

		var last_italic    *inline_format
		var last_bold      *inline_format
		var last_highlight *inline_format
		var last_underline *inline_format
		var last_strike    *inline_format

		// inline_balance will check against the
		// running score of "italic_on", etc. and
		// decide if the current format token is
		// right

		/*
			for example:

			*string *string end string*
			^ open  ^ open      close ^

			inline_balance will decide the
			second opener is invalid and reset
			it to "normal" text
		*/

		for _, entry := range the_list {
			switch entry.format_type {
			case ITALIC:
				italic_on = inline_balance(entry, italic_on)

			case BOLD:
				bold_on = inline_balance(entry, bold_on)

			case BOLDITALIC:
				x := inline_balance(entry, bold_on && italic_on)
				bold_on = x
				italic_on = x

			case HIGHLIGHT:
				highlight_on = inline_balance(entry, highlight_on)

			case UNDERLINE:
				underline_on = inline_balance(entry, underline_on)

			case STRIKEOUT:
				strike_on = inline_balance(entry, strike_on)
			}

			// if the first switch changed any
			// to "normal", we don't care now. we also
			// only care if they're "openers", so
			// is_confirmed would need to be true
			if entry.format_type == NORMAL || !entry.is_confirmed {
				continue
			}

			// remember the last entry under the above
			// conditions for future checking
			switch entry.format_type {
			case ITALIC:
				last_italic = entry

			case BOLD:
				last_bold = entry

			case BOLDITALIC:
				last_bold = entry
				last_italic = entry

			case HIGHLIGHT:
				last_highlight = entry

			case UNDERLINE:
				last_underline = entry

			case STRIKEOUT:
				last_strike = entry
			}
		}

		// if any of these are still active come the
		// end of the string, whoever started them is
		// bogus and should be reset to "normal".
		if italic_on {
			last_italic.format_type = NORMAL
		}
		if bold_on {
			last_bold.format_type = NORMAL
		}
		if highlight_on {
			last_highlight.format_type = NORMAL
		}
		if underline_on {
			last_underline.format_type = NORMAL
		}
		if strike_on {
			last_strike.format_type = NORMAL
		}
	}

	line_stack := make([]*syntax_line, 0, len(the_list) / 2)

	{
		x_string := strings.Builder {}
		x_string.Grow(max_width)

		x_length := 0

		highlight_range := make([]int, 0, 6)
		underline_range := make([]int, 0, 6)
		strikeout_range := make([]int, 0, 6)

		highlight_on := false
		underline_on := false
		strikeout_on := false

		test_width := max_width - para_indent

		leaf_stack := make([]*syntax_leaf, 0, len(the_list))

		for _, entry := range the_list {
			if x_length + entry.space_width > test_width || (entry.format_type == NORMAL && x_length + entry.text_width > test_width) {
				// @todo long word cutting

				if underline_on { underline_range = append(underline_range, x_length) }
				if highlight_on { highlight_range = append(highlight_range, x_length) }
				if strikeout_on { strikeout_range = append(strikeout_range, x_length) }

				x := x_string.String()

				leaf_stack = append(leaf_stack, &syntax_leaf {
					leaf_type:   NORMAL,
					text:        x,
				})

				x_string.Reset()
				x_string.Grow(max_width)

				line_stack = append(line_stack, &syntax_line {
					length:    x_length,
					leaves:    leaf_stack,
					highlight: highlight_range,
					underline: underline_range,
					strikeout: strikeout_range,
				})

				if len(line_stack) >= 1 {
					test_width = max_width
				}

				entry.space_width = 0

				leaf_stack = make([]*syntax_leaf, 0, len(the_list))

				highlight_range = make([]int, 0, 6)
				underline_range = make([]int, 0, 6)
				strikeout_range = make([]int, 0, 6)

				if underline_on { underline_range = append(underline_range, 0) }
				if highlight_on { highlight_range = append(highlight_range, 0) }
				if strikeout_on { strikeout_range = append(strikeout_range, 0) }

				x_length = 0
			}

			if entry.space_width > 0 {
				x_string.WriteString(strings.Repeat(" ", entry.space_width))
				x_length += entry.space_width
			}

			if entry.space_only {
				continue
			}

			if entry.format_type == NORMAL {
				x_string.WriteString(entry.text)
				x_length += entry.text_width
				continue
			}

			// hack to get notes placed correctly
			// this handles the closing side,
			// making sure their bracket
			// text is inserted as part of
			// the preceding leaf and not
			// dropped like everyone else
			if entry.format_type == NOTE && !entry.is_confirmed {
				x_string.WriteString(entry.text)
				x_length += entry.text_width
			}

			if entry.format_type != UNDERLINE && entry.format_type != HIGHLIGHT {
				x := x_string.String()

				leaf_stack = append(leaf_stack, &syntax_leaf {
					leaf_type:   NORMAL,
					text:        x,
				})

				x_string.Reset()
				x_string.Grow(max_width)
			}

			// hack to get notes placed correctly
			// this handles the opening side,
			// making sure their bracket
			// text is inserted as part of
			// the succeeding leaf and not
			// dropped like everyone else
			if entry.format_type == NOTE && entry.is_confirmed {
				x_string.WriteString(entry.text)
				x_length += entry.text_width
			}

			switch entry.format_type {
			case UNDERLINE:
				underline_on = entry.is_confirmed
				if underline_on {
					underline_range = append(underline_range, x_length)
				} else {
					underline_range = append(underline_range, x_length)
				}
				continue

			case HIGHLIGHT:
				highlight_on = entry.is_confirmed
				if highlight_on {
					highlight_range = append(highlight_range, x_length)
				} else {
					highlight_range = append(highlight_range, x_length)
				}
				continue

			case STRIKEOUT:
				strikeout_on = entry.is_confirmed
				if strikeout_on {
					strikeout_range = append(strikeout_range, x_length)
				} else {
					strikeout_range = append(strikeout_range, x_length)
				}
				continue
			}

			leaf_stack = append(leaf_stack, &syntax_leaf {
				leaf_type: entry.format_type,
				opening:   entry.is_confirmed,
			})
		}

		if x_string.Len() > 0 {
			x := x_string.String()

			leaf_stack = append(leaf_stack, &syntax_leaf {
				leaf_type:   NORMAL,
				text:        x,
			})
		}

		if len(leaf_stack) > 0 {
			line_stack = append(line_stack, &syntax_line {
				length:    x_length,
				leaves:    leaf_stack,
				highlight: highlight_range,
				underline: underline_range,
				strikeout: strikeout_range,
			})
		}
	}

	return line_stack
}

func formatter_tokens(input string) (string, int) {
	width := 0
	for _, c := range input {
		if unicode.IsSpace(c) || format_chars[c] {
			return input[:width], width
		}
		width += utf8.RuneLen(c)
	}
	return input, width
}

func inline_balance(entry *inline_format, check bool) bool {
	if entry.could_open && entry.could_close {
		if check {
			return false
		} else {
			entry.is_confirmed = true
			return true
		}
	} else if entry.could_open {
		if check {
			entry.format_type = NORMAL
			return check
		} else {
			entry.is_confirmed = true
			return true
		}
	} else if entry.could_close {
		if check {
			return false
		} else {
			entry.format_type = NORMAL
		}
	}
	return check
}

func extract_repeated_rune(input string, the_rune rune) (string, int) {
	width := 0
	for _, c := range input {
		if c != the_rune {
			return input[:width], width
		}
		width += utf8.RuneLen(c)
	}
	return input, width
}