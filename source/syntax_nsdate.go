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
	"strconv"
	"strings"
	"time"
	"unicode"
)

const default_timestamp = "d MMM yyyy"

var ns_magic_convert = map[string]string{
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

// this conversion system obviously isn't
// perfect, but it supports many common
// formatters and fills in the gaps in Go's
// magic numbers to be tighter to the base
// Unicode spec

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
				n := count_rune(input, c)
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
					return nsdate(default_timestamp) // just chuck the default back
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