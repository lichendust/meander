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
	"fmt"
	"bytes"
	"strings"
	"github.com/ledongthuc/pdf"
)

func convert_pdf(config *config) {
	if config.raw_convert {

	}

	content, ok := melt_pdf(config.source_file, config.raw_convert)
	if !ok {
		return
	}

	fmt.Println(content)
}

func melt_pdf(path string, do_raw bool) (string, bool) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", false
	}
	defer f.Close()

	if do_raw {
		buffer := bytes.Buffer{}
		buffer.Grow(1024 * 1024) // 1MB

		b, err := r.GetPlainText()
		if err != nil {
		    return "", false
		}

		buffer.ReadFrom(b)
		return buffer.String(), true
	}

	total_pages := r.NumPage()

	for page_index := 1; page_index <= total_pages; page_index += 1 {
		p := r.Page(page_index)
		if p.V.IsNull() {
			continue
		}

		texts := p.Content().Text

		var last pdf.Text
		var last_indent float64

		buffer := strings.Builder{}
		buffer.Grow(256)

		for _, text := range texts {
			/*
				@note @todo

				meander is leaving these three-byte sequences at
				the end of each block of text. i know it's
				not my code doing it, so it must be gopdf

				is this standard or formalised?  or do i need to
				somehow stop it from doing that?

				external testing pdfs seem not to do that
			*/

			/*if len(text.S) == 3 {
				a := []byte(text.S)
				b := []byte{239,191,189}

				matches := true

				for i, v := range a {
					if v != b[i] {
						matches = false
					}
				}

				if matches {
					continue main_loop
				}
			}*/

			if text.S == "’" {
				text.S = "'"
			}

			if last.Y == text.Y {
				buffer.WriteString(text.S)
				continue
			}

			if last.Y - text.Y <= pica * 1.5 {
				if last_indent == text.X {
					buffer.WriteString(text.S)
					last = text
					continue
				}

				fmt.Print(buffer.String())
				fmt.Print("\n")
			} else {
				fmt.Print(buffer.String())
				fmt.Print("\n\n")
			}

			/*
				@note

				when we capture a word or a line, if the
				line-break point has no trailing space,
				we're not adding any ourselves. this means
				we need to remember if the last single
				char/block was a space so we recombine correctly
			*/

			// fmt.Printf("font: %s\nsize: %f\nxxxx: %f\nyyyy: %f\ncont: %s\n", text.Font, text.FontSize, text.X, text.Y, buffer.String())
			// fmt.Println(last.Y - text.Y, "\n")

			last = text
			last_indent = text.X

			buffer.Reset()
			buffer.Grow(256)

			buffer.WriteString(text.S)
		}
	}

	return "", true
}