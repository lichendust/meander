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

import "os"
import "fmt"
import "time"
import "strings"
import "unicode"
import "unicode/utf8"
import "path/filepath"

import "github.com/mattn/go-isatty"

var now = time.Now
var ascii_space = [256]uint8{'\t':1,'\n':1,'\v':1,'\f':1,'\r':1,' ':1}

var running_in_term bool

func init() {
	running_in_term = isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

func left_trim(input string) string {
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

func left_trim_ignore_newlines(input string) string {
	ascii_space['\n'] = 0
	defer func() { ascii_space['\n'] = 1 }()
	return left_trim(input)
}

func right_trim(input string) string {
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

func is_format_char(x rune) bool {
	switch x {
	case '*':  return true
	case '+':  return true
	case '~':  return true
	case '_':  return true
	case ']':  return true
	case '[':  return true
	case '$': return true
	case '#': return true
	case '\\': return true
	case '\n': return true
	}
	return false
}

func non_token_word(input string) (string, int) {
	width := 0
	for _, c := range input {
		if unicode.IsSpace(c) || is_format_char(c) {
			return input[:width], width
		}
		width += utf8.RuneLen(c)
	}
	return input, width
}

func extract_ident(input string) (string, int) {
	width := 0
	for _, c := range input {
		if !(unicode.IsLetter(c) || c == '_') {
			return input[:width], width
		}
		width += utf8.RuneLen(c)
	}
	return input, width
}

func extract_letters_or_numbers(input string) (string, int) {
	width := 0
	for _, c := range input {
		if !(unicode.IsLetter(c) || unicode.IsNumber(c)) {
			return input[:width], width
		}
		width += utf8.RuneLen(c)
	}
	return input, width
}

func is_all_numbers(input string) bool {
	for _, c := range input {
		if !unicode.IsNumber(c) {
			return false
		}
	}
	return true
}

func is_all_letters(input string) bool {
	for _, c := range input {
		if !unicode.IsLetter(c) {
			return false
		}
	}
	return true
}

// github.com/arturwwl/gointtoletters
func alphabetical_increment(number int) string {
	number  -= 1
	letters := ""

	if first_letter := number / 26; first_letter > 0 {
		letters += alphabetical_increment(first_letter)
		letters += string('A' + number % 26)
	} else {
		letters += string('A' + number)
	}

	return letters
}

// this is a much nicer implementation what i
// wrote but it gives a reversed string and
// i'm too stupid to figure out how to get
// it to not do that.
/*func alphabetical_increment(number int) string {
	buffer := new(strings.Builder)
	buffer.Grow(3)

	number -= 1

	for {
		letter := number / 26

		if letter > 0 {
			buffer.WriteRune(rune('A' + number % 26))
			number = letter - 1
			continue
		}

		buffer.WriteRune(rune('A' + number))
		break
	}

	return buffer.String()
}*/

func alphabet_to_int(input string) int {
	number := 0
	length := len(input) - 1

	for _, c := range input {
		x := int(c - 64)

		number += x * int_power(26, length)
		length -= 1
	}

	return number
}

func int_power(n, m int) int {
	if m == 0 {
		return 1
	}

	result := n
	for i := 2; i <= m; i += 1 {
		result *= n
	}

	return result
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

		case '`', '‘':
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
		}

		last_rune = c
		buffer.WriteRune(c)
	}

	return buffer.String()
}

// homogenise "Draft Date" or "draft_date" into "draftdate"
// this helps us simplify any multi-matches in the title page
func homogenise(input string) string {
	buffer := strings.Builder{}
	buffer.Grow(len(input))

	for _, c := range input {
		if c >= utf8.RuneSelf {
			continue
		}
		if ascii_space[c] == 1 {
			continue
		}
		if c == '_' || c == '-' {
			continue
		}
		buffer.WriteRune(unicode.ToLower(c))
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

func count_whitespace(input string) int {
	for i, c := range input {
		if !unicode.IsSpace(c) {
			return i
		}
	}
	return len(input)
}

// utf8.RuneCountInString
func rune_count(input string) int {
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

func rune_on_line(input string, x rune) int {
	for i, c := range input {
		if c == x {
			return i + 1
		}
		if c == '\n' {
			break
		}
	}

	return -1
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
	if input == "" {
		return ""
	}

	buffer := strings.Builder{}
	buffer.Grow(len(input))

	for _, c := range input {
		if c == '\n' {
			buffer.WriteRune(' ')
			continue
		}
		if is_format_char(c) {
			continue
		}
		buffer.WriteRune(c)
	}

	return buffer.String()
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

func short_words(t string) bool {
	switch t {
	case "a":   return true
	case "an":  return true
	case "and": return true
	case "the": return true
	case "on":  return true
	case "to":  return true
	case "in":  return true
	case "for": return true
	case "nor": return true
	case "or":  return true
	}
	return false
}

func title_case(input string) string {
	input = normalise_text(input)

	words := strings.Split(input, " ")

	for i, word := range words {
		if i > 0 && short_words(word) {
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

func word_count(text string) int {
	total_count := 0
	each_word   := 0

	for {
		if len(text) == 0 {
			break
		}

		r, w := utf8.DecodeRuneInString(text)
		text = text[w:]

		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			each_word += 1
			continue
		}

		if each_word > 0 {
			total_count += 1
			each_word = 0
		}
	}

	if each_word > 0 {
		total_count += 1
	}

	return total_count
}

func load_file(source_file string) (string, bool) {
	bytes, err := os.ReadFile(source_file)
	if err != nil {
		return "", false
	}
	return string(bytes), true
}

func load_file_normalise(source_file string) (string, bool) {
	text, ok := load_file(source_file)
	if !ok {
		return "", false
	}
	return normalise_text(text), ok
}

func write_file(path string, content []byte) bool {
	return os.WriteFile(path, content, os.ModePerm) == nil
}

func fix_path(input string) string {
	if path, err := filepath.Abs(input); err == nil {
		return path
	} else {
		panic(err)
	}
}

func rewrite_ext(path, new_ext string) string {
	ext := filepath.Ext(path)
	raw := path[:len(path)-len(ext)]

	return raw + new_ext
}

func include_path(parent, input string) string {
	if !filepath.IsAbs(input) {
		return filepath.Join(filepath.Dir(parent), input)
	}
	return input
}

func print(words ...string) {
	l := len(words) - 1
	for i, w := range words {
		os.Stdout.WriteString(w)
		if i < l {
			os.Stdout.WriteString(" ")
		}
	}
}

func println(words ...string) {
	l := len(words) - 1
	for i, w := range words {
		os.Stdout.WriteString(w)
		if i < l {
			os.Stdout.WriteString(" ")
		}
	}
	os.Stdout.WriteString("\n")
}

func eprintln(words ...string) {
	l := len(words) - 1
	for i, w := range words {
		os.Stderr.WriteString(w)
		if i < l {
			os.Stderr.WriteString(" ")
		}
	}
	os.Stderr.WriteString("\n")
}

func eprintf(format string, guff ...any) {
	fmt.Fprintf(os.Stderr, format, guff...)
	os.Stderr.WriteString("\n")
}

const ANSI_RESET = "\033[0m"
const ANSI_COLOR = "\033[91m"

func apply_color(input string) string {

	buffer := strings.Builder{}
	buffer.Grow(len(input) + 128)

	last_rune := 'x'

	for {
		if len(input) == 0 {
			break
		}

		r, w := utf8.DecodeRuneInString(input)
		input = input[w:]

		if r == '$' {
			last_rune = r
			continue
		}

		if last_rune == '$' {
			last_rune = 'x'

			if r == '0' || r == '1' {
				if !running_in_term {
					continue
				} else if r == '0' {
					buffer.WriteString(ANSI_RESET)
				} else {
					buffer.WriteString(ANSI_COLOR)
				}
			} else {
				buffer.WriteRune('$')
				buffer.WriteRune(r)
			}

			continue
		}

		last_rune = r
		buffer.WriteRune(r)
	}

	return buffer.String()
}