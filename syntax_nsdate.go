package main

import (
	"time"
	"strconv"
	"strings"
	"unicode"
)

var ns_magic_convert = map[string]string {
	"M":    "1",
	"MM":   "01",
	"MMM":  "Jan",
	"MMMM": "January",

	"d":    "2",
	"dd":   "02",

	"E":    "Mon",
	"EEEE": "Monday",

	"h":    "3",
	"hh":   "03",
	"HH":   "15",
	"a":    "PM",

	"m":    "4",
	"mm":   "04",

	"s":    "5",
	"ss":   "05",
	"SSS":  ".000",
}

const default_timestamp = "d MMM yyyy"

// this conversion system obviously isn't
// perfect, but it supports many common
// formatters and fills in the gaps in Go's
// magic numbers to be tighter to the base
// Unicode spec

// unsupported formatters are warned to the user
// and ommitted from render.

func nsdate(input string) string {
	final := strings.Builder{}

	t := time.Now()

	input = strings.TrimSpace(input)

	for {
		if len(input) == 0 {
			break
		}

		for _, c := range input {
			if unicode.IsLetter(c) {
				n      := count_rune(input, c)
				repeat := input[:n]

				// years
				if c == 'y' {
					switch n {
					case 1:
						final.WriteString(strconv.Itoa(t.Year()))
					case 2:
						final.WriteString(t.Format("06"))
					default:
						y := strconv.Itoa(t.Year())
						final.WriteString(strings.Repeat("0", clamp(n-len(y))))
						final.WriteString(y)
					}
					input = input[n:]
					break
				}

				// H - unpadded hour
				if c == 'H' && n == 1 {
					final.WriteString(strconv.Itoa(t.Hour()))
				}
				// MMMMM - single letter month
				if c == 'M' && n == 5 {
					final.WriteString(t.Month().String()[:1])
				}
				// EEEEE - single letter week
				if c == 'E' && n == 5 {
					final.WriteString(t.Weekday().String()[:1])
				}
				// EEEEEE - two letter week
				if c == 'E' && n == 6 {
					final.WriteString(t.Weekday().String()[:2])
				}

				if x, ok := ns_magic_convert[repeat]; ok {
					final.WriteString(t.Format(x))
				} else {
					return nsdate(default_timestamp)
				}

				input = input[n:]
				break
			}

			final.WriteRune(c)
			input = input[1:]
		}
	}

	return final.String()
}

func clamp(m int) int {
	if m < 0 {
		return 0
	}
	return m
}