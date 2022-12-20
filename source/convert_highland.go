package main

import (
	"io"
	"os"
	"strings"
	"unicode/utf8"
	"archive/zip"
	"path/filepath"
)

func command_convert_highland(config *config) {
	highland_convert(config.source_file)
}

func highland_convert(file string) {
	file = filepath.ToSlash(file)

	archive, err := zip.OpenReader(file)
	if err != nil {
		panic(err)
	}
	defer archive.Close()

	current_location := filepath.Dir(file)
	output_name      := rewrite_ext(file, ".fountain")

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

			bblob, err := io.ReadAll(s)
			if err != nil {
				panic(err)
			}

			blob := highland_includes(current_location, string(bblob))

			err = os.WriteFile(output_name, []byte(blob), 0777)
			if err != nil {
				panic(err)
			}
		}
	}
}

func highland_includes(current_location, text string) string {
	buffer := strings.Builder {}
	buffer.Grow(len(text))

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
					if filepath.Ext(child_file) == ".highland" {
						highland_convert(child_file)
						child_file = rewrite_ext(child_file, ".fountain")
					}

					buffer.WriteString("{{include: ")
					buffer.WriteString(child_file)
					buffer.WriteString("}}")
				}

				text = text[n + 2:]
				continue
			}
		}

		the_rune, rune_width := utf8.DecodeRuneInString(text)
		text = text[rune_width:]
		buffer.WriteRune(the_rune)
	}

	return buffer.String()
}