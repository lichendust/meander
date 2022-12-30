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
	"io"
	"bytes"
	"archive/zip"
	"unicode/utf8"
	"path/filepath"
)

const highland_extension = ".highland"

func convert_highland(config *config) {
	recurse_convert_highland(config.source_file)
}

func recurse_convert_highland(file string) {
	file = filepath.ToSlash(file)

	archive, err := zip.OpenReader(file)
	if err != nil {
		panic(err)
	}
	defer archive.Close()

	current_location := filepath.Dir(file)
	output_name := rewrite_ext(file, fountain_extension)

	for _, f := range archive.File {
		name := filepath.Base(f.Name)

		if name == "text.md" {
			s, err := f.Open()
			if err != nil {
				panic(err)
			}
			defer s.Close()

			/*d, err := os.OpenFile(output_name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
			if err != nil {
				panic(err)
			}
			defer d.Close()*/

			blob, err := io.ReadAll(s)
			if err != nil {
				panic(err)
			}

			{
				text := string(blob)

				buffer := bytes.Buffer{}
				buffer.Grow(len(text) + 256)

				for {
					if len(text) == 0 {
						break
					}

					if text[0] == '{' && len(text) > 1 && text[1] == '{' {
						n := rune_pair(text[2:], '}', '}')

						if n < 0 {
							buffer.WriteString(text[:2])
							text = text[2:]
							continue
						}

						test_text := text[2:n]

						if path, ok := macro(test_text, "include"); ok {
							if child_file, ok := find_file(current_location, path); ok {
								// if it's a highland file, we do all this again
								// and we change the reference in this file to the
								// new filename
								if filepath.Ext(child_file) == highland_extension {
									recurse_convert_highland(child_file)
									child_file = rewrite_ext(child_file, fountain_extension)
								}

								// rewrite the discovered path into the file
								buffer.WriteString("{{include: ")
								buffer.WriteString(child_file)
								buffer.WriteString("}}")
							}

							text = text[n+2:]
							continue
						}
					}

					the_rune, rune_width := utf8.DecodeRuneInString(text)
					text = text[rune_width:]
					buffer.WriteRune(the_rune)
				}

				blob = buffer.Bytes()
			}

			ok := write_file(output_name, blob)
			if !ok {
				eprintln("failed to write", output_name)
			}
		}
	}
}