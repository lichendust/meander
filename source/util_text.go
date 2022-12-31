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

var ascii_space = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}

func consume_whitespace(input string) string {
	start := 0

	for ; start < len(input); start += 1 {
		c := input[start]

		if c >= utf8.RuneSelf {
			return strings.TrimFunc(input[start:], unicode.IsSpace)
		}

		if ascii_space[c] == 0 {
			break
		}
	}

	return input[start:]
}

func consume_ending_whitespace(input string) string {
	start := 0
	stop := len(input)

	for ; stop > start; stop-- {
		c := input[stop-1]

		if c >= utf8.RuneSelf {
			return strings.TrimFunc(input[start:stop], unicode.IsSpace)
		}

		if ascii_space[c] == 0 {
			break
		}
	}

	return input[start:stop]
}

func consume_newlines(input string) string {
	for i, c := range input {
		if c != '\n' {
			return input[i:]
		}
	}
	return input
}

func normalise_text(input string) string {
	buffer := strings.Builder{}
	buffer.Grow(len(input))

	input = strings.TrimSpace(input)

	last_rune := 'a'

	for _, c := range input {
		switch c {
		case '“', '”':
			buffer.WriteRune('"')
			last_rune = c
			continue

		case '’':
			buffer.WriteRune('\'')
			last_rune = c
			continue

		case '`':
			buffer.WriteRune('\'')
			last_rune = c
			continue

		case '\n':
			if last_rune == '\r' {
				continue
			}
			buffer.WriteRune('\n')
			last_rune = c
			continue

		case '\r':
			if last_rune == '\n' {
				continue
			}
			buffer.WriteRune('\n')
			last_rune = c
			continue

		case '\t':
			buffer.WriteString(`    `) // 4 spaces
			last_rune = c
			continue

		// double width em dash
		case '—':
			buffer.WriteString("——")
			last_rune = c
			continue
		}

		last_rune = c
		buffer.WriteRune(c)
	}

	return buffer.String()
}

func extract_to_newline(input string) string {
	for i, c := range input {
		if c == '\n' {
			return input[:i]
		}
	}
	return input
}

func extract_number(input string) string {
	for i, c := range input {
		if !unicode.IsNumber(c) {
			return input[:i]
		}
	}
	return input
}

func count_whitespace(input string) int {
	for i, c := range input {
		if !unicode.IsSpace(c) {
			return i
		}
	}
	return len(input)
}

// utf8.RuneCountInString
func count_all_runes(input string) int {
	count := 0
	for range input {
		count += 1
	}
	return count
}

func count_rune(input string, r rune) int {
	count := 0
	for _, c := range input {
		if c != r {
			return count
		}
		count += 1
	}
	return count
}

func split_rune(input string, r rune) []string {
	for i, c := range input {
		if c == r {
			if i == len(input) {
				return []string{input}
			}
			return []string{input[:i], input[i+1:]}
		}
	}
	return []string{input}
}

func rune_pair(text string, x, y rune) int {
	last := 'a'

	for i, c := range text {
		if c == y && last == x { // swapped from above
			return i + 1
		}
		last = c
	}

	return -1
}

func clean_string(input string) string {
	buffer := strings.Builder{}
	buffer.Grow(len(input))

	for _, c := range input {
		if format_chars[c] {
			continue
		}
		if c == '\n' {
			c = ' '
		}
		buffer.WriteRune(c)
	}

	return buffer.String()
}

// this is not an exhaustive list
var short_words = map[string]bool{
	"a":   true,
	"an":  true,
	"and": true,
	"the": true,
	"on":  true,
	"to":  true,
	"in":  true,
	"for": true,
	"nor": true,
	"or":  true,
}

// a title caser that actually works!
func title_case(input string) string {
	input = normalise_text(input)

	words := strings.Split(input, " ")

	for i, word := range words {
		if i > 0 && short_words[word] {
			continue
		}

		buffer := strings.Builder{}
		buffer.Grow(len(word))

		for len(word) > 0 {
			c, width := utf8.DecodeRuneInString(word)

			if buffer.Len() == 0 {
				buffer.WriteRune(unicode.ToUpper(c))
				word = word[width:]
				continue
			}

			if c == '-' || c == '—' {
				buffer.WriteRune(unicode.ToLower(c))
				word = word[width:]

				c, width = utf8.DecodeRuneInString(word)

				buffer.WriteRune(unicode.ToUpper(c))
				word = word[width:]
				continue
			}

			buffer.WriteRune(unicode.ToLower(c))
			word = word[width:]
		}

		words[i] = buffer.String()
	}

	return strings.Join(words, " ")
}

func space_pad_string(input string, n int) string {
	return input + strings.Repeat(" ", n+2-count_all_runes(input))
}
