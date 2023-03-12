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
	"strings"
	"unicode"
	"unicode/utf8"
)

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
	length int

	// array of text data — only "normal"
	// leaves **ever** go to print, so
	// anything in between is a
	// defining "switch" that does not
	// contain anything
	leaves []*syntax_leaf

	// ranges for each block-draw element, in
	// the form [start, end, start, end]
	underline []int
	strikeout []int
	highlight []int

	// used to force a stylistic override on
	// the entire line.
	// gopdf uses a horrid "BI" string to define
	font_reset string
}

// a chunk of text or formatter switch
type syntax_leaf struct {
	leaf_type int

	// whether the leaf is the
	// start or end of a range
	is_opening bool

	// content for normal leaves
	text string
}

// intermediary structure with extra members for
// processing the leaves
type inline_format struct {
	syntax_leaf
	space_width int
	text_width  int
	could_open  bool
	could_close bool
	space_only  bool
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

		case '\'':
			the_word   = "'"
			rune_width = 1
			is_type    = QUOTE

		case '"':
			the_word   = "\""
			rune_width = 1
			is_type    = DOUBLE_QUOTE

		case '*':
			the_word, rune_width = extract_repeated_rune(input, '*')

			switch rune_width {
			case 1:
				is_type = ITALIC
			case 2:
				is_type = BOLD
			case 3:
				is_type = BOLDITALIC
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

		format := &inline_format{}

		format.leaf_type = is_type
		format.text = the_word
		format.space_width = space_width
		format.text_width = rune_width

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
			if entry.leaf_type == NORMAL {
				continue
			}

			if entry.leaf_type == NOTE {
				entry.is_opening = (entry.text[0] == '[')
				continue
			}

			// handle escape sequences
			if entry.leaf_type == ESCAPE {
				n := len(entry.text)

				/*
					there's a looming set of @todos at the top
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
					entry.leaf_type = NORMAL
					entry.space_only = false

					entry.text = entry.text[n/2:]
					continue
				}

				// otherwise, if they're odd but > 1, do the same
				// but subtract the final odd-one-out and apply it
				// as a legit escape
				if n > 1 {
					entry.leaf_type = NORMAL
					entry.space_only = false

					entry.text = entry.text[(n + 1)/2 - 1:]
				}

				// lookahead and escape the next item as applicable
				if len(the_list[i:]) > 1 {
					target := the_list[i + 1:][0]
					if target.leaf_type != NORMAL {
						target.leaf_type = NORMAL
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
				entry.leaf_type = NORMAL
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
		i_on := false
		b_on := false
		h_on := false
		u_on := false
		s_on := false

		var i_last *inline_format
		var b_last *inline_format
		var h_last *inline_format
		var u_last *inline_format
		var s_last *inline_format

		// inline_balance will check against the
		// running score of "i_on", etc. and
		// decide if the current format token is
		// right

		/*
			for example:

			*string *string end string*
			^ open  ^ open      close ^

			inline_balance will decide the second opener is
			invalid and reset it to "normal" text

			this is the opposite of a lot of fountain parsers,
			which would typically match the "closest pair"
			via regex - in the example above, the second
			opener would be considered the most valid one
		*/

		for _, entry := range the_list {
			switch entry.leaf_type {
			case QUOTE:
				if entry.could_open && !entry.could_close {
					entry.text = "‘"
				} else {
					entry.text = "’"
				}
				entry.leaf_type = NORMAL
				continue

			case DOUBLE_QUOTE:
				if entry.could_open && !entry.could_close {
					entry.text = "“"
				} else {
					entry.text = "”"
				}
				entry.leaf_type = NORMAL
				continue

			case ITALIC:
				i_on = inline_balance(entry, i_on)
				if entry.is_opening {
					i_last = entry
					break
				}
				continue

			case BOLD:
				b_on = inline_balance(entry, b_on)
				if entry.is_opening {
					b_last = entry
					break
				}
				continue

			case BOLDITALIC:
				x := inline_balance(entry, b_on && i_on)
				b_on = x
				i_on = x

				if entry.is_opening {
					b_last = entry
					i_last = entry
					break
				}
				continue

			case HIGHLIGHT:
				h_on = inline_balance(entry, h_on)
				if entry.is_opening {
					h_last = entry
					break
				}
				continue

			case UNDERLINE:
				u_on = inline_balance(entry, u_on)
				if entry.is_opening {
					u_last = entry
					break
				}
				continue

			case STRIKEOUT:
				s_on = inline_balance(entry, s_on)
				if entry.is_opening {
					s_last = entry
					break
				}
				continue

			case NORMAL:
				continue
			}
		}

		// if any of these are still active come the
		// end of the string, whoever started them is
		// bogus and should be reset to "normal".
		if i_on { i_last.leaf_type = NORMAL }
		if b_on { b_last.leaf_type = NORMAL }
		if h_on { h_last.leaf_type = NORMAL }
		if u_on { u_last.leaf_type = NORMAL }
		if s_on { s_last.leaf_type = NORMAL }
	}

	line_stack := make([]*syntax_line, 0, len(the_list) / 2)

	{
		line_buffer := strings.Builder{}
		line_buffer.Grow(max_width)

		line_length := 0

		highlight_range := make([]int, 0, 6)
		underline_range := make([]int, 0, 6)
		strikeout_range := make([]int, 0, 6)

		highlight_on := false
		underline_on := false
		strikeout_on := false

		test_width := max_width - para_indent

		leaf_stack := make([]*syntax_leaf, 0, len(the_list))

		for _, entry := range the_list {
			if line_length + entry.space_width > test_width || (entry.leaf_type == NORMAL && line_length + entry.text_width > test_width) {
				// @todo long word cutting

				if underline_on {
					underline_range = append(underline_range, line_length)
				}
				if highlight_on {
					highlight_range = append(highlight_range, line_length)
				}
				if strikeout_on {
					strikeout_range = append(strikeout_range, line_length)
				}

				x := line_buffer.String()

				leaf_stack = append(leaf_stack, &syntax_leaf{
					leaf_type: NORMAL,
					text:      x,
				})

				line_buffer.Reset()
				line_buffer.Grow(max_width)

				line_stack = append(line_stack, &syntax_line{
					length:    line_length,
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

				if underline_on {
					underline_range = append(underline_range, 0)
				}
				if highlight_on {
					highlight_range = append(highlight_range, 0)
				}
				if strikeout_on {
					strikeout_range = append(strikeout_range, 0)
				}

				line_length = 0
			}

			if entry.space_width > 0 {
				line_buffer.WriteString(strings.Repeat(" ", entry.space_width))
				line_length += entry.space_width
			}

			if entry.space_only {
				continue
			}

			if entry.leaf_type == NORMAL {
				line_buffer.WriteString(entry.text)
				line_length += entry.text_width
				continue
			}

			// hack to get notes placed correctly
			// this handles the closing side,
			// making sure their bracket
			// text is inserted as part of
			// the preceding leaf and not
			// dropped like everyone else
			if entry.leaf_type == NOTE && !entry.is_opening {
				line_buffer.WriteString(entry.text)
				line_length += entry.text_width
			}

			if entry.leaf_type != UNDERLINE && entry.leaf_type != HIGHLIGHT {
				x := line_buffer.String()

				leaf_stack = append(leaf_stack, &syntax_leaf{
					leaf_type: NORMAL,
					text:      x,
				})

				line_buffer.Reset()
				line_buffer.Grow(max_width)
			}

			// hack to get notes placed correctly
			// this handles the is_opening side,
			// making sure their bracket
			// text is inserted as part of
			// the succeeding leaf and not
			// dropped like everyone else
			if entry.leaf_type == NOTE && entry.is_opening {
				line_buffer.WriteString(entry.text)
				line_length += entry.text_width
			}

			switch entry.leaf_type {
			case UNDERLINE:
				underline_on = entry.is_opening
				if underline_on {
					underline_range = append(underline_range, line_length)
				} else {
					underline_range = append(underline_range, line_length)
				}
				continue

			case HIGHLIGHT:
				highlight_on = entry.is_opening
				if highlight_on {
					highlight_range = append(highlight_range, line_length)
				} else {
					highlight_range = append(highlight_range, line_length)
				}
				continue

			case STRIKEOUT:
				strikeout_on = entry.is_opening
				if strikeout_on {
					strikeout_range = append(strikeout_range, line_length)
				} else {
					strikeout_range = append(strikeout_range, line_length)
				}
				continue
			}

			leaf_stack = append(leaf_stack, &syntax_leaf{
				leaf_type:  entry.leaf_type,
				is_opening: entry.is_opening,
			})
		}

		if line_buffer.Len() > 0 {
			x := line_buffer.String()

			leaf_stack = append(leaf_stack, &syntax_leaf{
				leaf_type: NORMAL,
				text:      x,
			})
		}

		if len(leaf_stack) > 0 {
			line_stack = append(line_stack, &syntax_line{
				length:    line_length,
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
			entry.is_opening = true
			return true
		}
	} else if entry.could_open {
		if check {
			entry.leaf_type = NORMAL
			return check
		} else {
			entry.is_opening = true
			return true
		}
	} else if entry.could_close {
		if check {
			return false
		} else {
			entry.leaf_type = NORMAL
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