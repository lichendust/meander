package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

var ascii_space = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}

func consume_whitespace(input string) string {
	start := 0

	for ; start < len(input); start++ {
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
	stop  := len(input)

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
	buffer := strings.Builder {}
	buffer.Grow(len(input))

	input = strings.TrimSpace(input)

	last_rune := 'a'

	for _, c := range input {
		switch c {
		case '“','”':
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

func is_valid_scene(line string) bool {
	word := ""

	for i, c := range line {
		if c == '.' {
			word = line[:i]
			break
		}

		if c >= utf8.RuneSelf {
			if unicode.IsSpace(c) {
				word = line[:i]
				break
			}
			continue
		}

		if ascii_space[c] == 1 {
			word = line[:i]
			break
		}
	}

	if len(word) > 0 {
		return valid_scene[strings.ToLower(clean_string(word))]
	}

	return false
}

func is_valid_character(line string) bool {
	for i, c := range line {
		if !format_chars[c] {
			line = line[i:]
			break
		}
	}

	/*if line[0] == 'M' {
		if strings.HasPrefix(line, "Mc") {
			line = line[2:]
		}
		if strings.HasPrefix(line, "Mac") {
			line = line[3:]
		}
	}*/

	// characters must start with a letter
	if !unicode.IsLetter(rune(line[0])) {
		return false
	}

	has_letters := false
	first_char  := true
	copy        := line

	for len(copy) > 0 {
		c, rune_width := utf8.DecodeRuneInString(copy)

		if c == '(' && !first_char {
			for i, c := range copy {
				if c == ')' {
					copy = copy[i:]
					break
				}
			}
		}

		if unicode.IsLetter(c) {
			has_letters = true

			if !unicode.IsUpper(c) {
				return false
			}
		}

		copy = copy[rune_width:]
		first_char = false
	}

	return has_letters
}

func is_title_element(line string) (string, string, bool) {
	if !strings.Contains(line, ":") {
		return "", "", false
	}

	t := strings.SplitN(line, ":", 2)

	name, value := strings.ToLower(t[0]), t[1]

	if valid_title_page[name] {
		return name, value, true
	}

	return "", "", false
}

func has_scene_number(text string) (string, string, bool) {
	if text[len(text) - 1] == '#' {
		n := 0
		t := text[:len(text) - 1]

		for i := len(t) - 1; i > 0; i-- {
			the_rune, _ := utf8.DecodeLastRuneInString(t[:i + 1])

			if unicode.IsSpace(the_rune) {
				break
			}

			if the_rune == '#' {
				n = i
				break
			}
		}

		if n != 0 {
			return strings.TrimSpace(text[:n]), t[n + 1:], true
		}
	}

	return "", "", false
}

func is_all_uppercase(text string) bool {
	for _, c := range text {
		if unicode.IsLetter(c) && !unicode.IsUpper(c) {
			return false
		}
	}
	return true
}

// quickly discards the title page for included files
func consume_title_page(input string) string {
	if n := strings.IndexRune(input, '\n'); n > -1 {
		if _, _, ok := is_title_element(input[:n]); ok {
			if n := rune_pair(input, '\n', '\n'); n > -1 {
				return strings.TrimSpace(input[n:])
			}
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
		count++
	}
	return count
}

func count_rune(input string, r rune) int {
	count := 0
	for _, c := range input {
		if c != r {
			return count
		}
		count++
	}
	return count
}

func split_rune(input string, r rune) []string {
	for i, c := range input {
		if c == r {
			if i == len(input) {
				return []string {input}
			}
			return []string {input[:i], input[i + 1:]}
		}
	}
	return []string {input}
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
var short_words = map[string]bool {
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

		buffer := strings.Builder {}
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
	return input + strings.Repeat(" ", n + 2 - count_all_runes(input))
}