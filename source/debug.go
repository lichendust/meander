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

//go:build dev

package main

import "fmt"

func (f *inline_format) String() string {
	could_open := " "
	could_close := " "
	is_confirmed := " "

	if f.leaf_type > NORMAL {
		is_confirmed = "C"
	}

	if f.could_close {
		could_close = "C"
	}

	if f.could_open {
		could_open = "O"
	}

	if f.is_opening {
		is_confirmed = "O"
	}

	return fmt.Sprintf(
		"%d %-12s %s %s %s %q",
		f.space_width,
		leaf_type_convert[f.leaf_type],
		could_open,
		could_close,
		is_confirmed,
		f.text,
	)
}

func (line *syntax_line) String() string {
	// leaves    []*syntax_leaf
	// highlight []int
	// underline []int
	// return fmt.Sprintf("%d\n", line.length)

	the_string := "\n"

	the_string += fmt.Sprintln(line.length)
	the_string += fmt.Sprintln(line.strikeout)

	for _, leaf := range line.leaves {
		if leaf.leaf_type == NORMAL {
			the_string += leaf.text
		} else {
			the_string += leaf_type_convert[leaf.leaf_type]
		}
	}

	return the_string
}

var leaf_type_convert = map[int]string{
	NORMAL:     "<normal>",
	BOLD:       "<bold>",
	ITALIC:     "<italic>",
	BOLDITALIC: "<bold_italic>",
	UNDERLINE:  "<underline>",
	STRIKEOUT:  "<strikeout>",
	HIGHLIGHT:  "<highlight>",
	NOTE:       "<note>",
	ESCAPE:     "<escape>",
	// NEWLINE:    "<newline>",
}

func debug_title_page(title map[string]string) {
	for _, c := range debug_title_order {
		if x, ok := title[c]; ok {
			fmt.Printf("%s\n%s\n\n", c, x)
		}
	}
}

var debug_title_order = []string{
	"format",
	"title",
	"credit",
	"author",
	"source",
	"contact",
	"revision",
	"draft date",
	"notes",
	"copyright",
}

func debug_print_screenplay(content *fountain_content) {
	for _, node := range content.nodes {
		fmt.Println(node)
	}
}

func (node *syntax_node) String() string {
	kind := node_type_convert[node.node_type]

	if node.node_type == WHITESPACE {
		if node.level > 1 {
			return fmt.Sprintf("\n--- %d\n", node.level)
		}
		return ""
	}

	if node.raw_text == "" {
		return fmt.Sprintf("%-15s%-4d", kind, node.level)
	}

	return fmt.Sprintf("%-15s%-4d|%s", kind, node.level, node.raw_text)
}