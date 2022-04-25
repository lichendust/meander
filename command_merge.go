package main

import (
	"os"
	"fmt"
	"strings"
	"io/ioutil"
	"unicode/utf8"
)

func command_merge_document(config *config) {
	merged_file, ok := merge(config.source_file)

	if !ok {
		fmt.Fprintf(os.Stderr, "failed to merge file %s\n", config.source_file)
		return
	}

	// @todo replace me with standard file writer
	err := ioutil.WriteFile(fix_path(config.output_file), []byte(merged_file), 0777)

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write %s\n", config.output_file)
	}
}

func merge(source_file string) (string, bool) {
	text, ok := load_file(fix_path(source_file))

	if !ok {
		fmt.Fprintf(os.Stderr, "%q not found\n", source_file)
		return "", false
	}

	content := strings.Builder {}
	content.Grow(len(text))

	for {
		if len(text) == 0 {
			break
		}

		if text[0] == '{' && len(text) > 1 && text[1] == '{' {
			n := rune_pair(text[2:], '}', '}')

			if n < 0 {
				content.WriteString(text[:2])
				text = text[2:]
				continue
			}

			test_text := text[2:n]

			if path, ok := macro(test_text, "include"); ok {
				path = include_path(source_file, path)

				if child_text, ok := merge(path); ok {
					content.Grow(len(child_text))
					content.WriteString(consume_title_page(child_text))
				} else {
					content.WriteString(text[:n + 2])
				}

				text = text[n + 2:]
				continue
			}
		}

		the_rune, rune_width := utf8.DecodeRuneInString(text)

		text = text[rune_width:]

		content.WriteRune(the_rune)
	}

	return content.String(), true
}