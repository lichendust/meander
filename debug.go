//go:build helpers

package main

import (
	"fmt"
	"strings"
)

var debug_title_order = []string {
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

func debug_arguments(conf *config) {
	scene := ""

	switch conf.scenes {
	case SCENE_INPUT:    scene = "input"
	case SCENE_REMOVE:   scene = "remove"
	case SCENE_GENERATE: scene = "generate"
	}

	fmt.Printf("source: %q\noutput: %q\n\nscenes: %s\nformat: %s\npaper: %s\n",
		fix_path(conf.source_file),
		fix_path(conf.output_file),

		scene,
		conf.template,
		conf.paper_size,
	)
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

func (f *inline_format) String() string {
	could_open   := " "
	could_close  := " "
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

/*func (leaf *syntax_leaf) String() string {
	// leaf_type   uint8
	// space_width int
	// text_width  int
	// opening     bool
	// space_only  bool
	// text        string

	if leaf.leaf_type == NORMAL {
		return leaf.text
	}
	return leaf_type_convert[leaf.leaf_type]
}*/

/*func debug_formatting(input string, max_width int) {
	list, _ := syntax_leaf_parser(input, max_width, 0)

	fmt.Printf("\n%q\n\"", input)

	for _, entry := range list {
		fmt.Printf(strings.Repeat(" ", entry.space_width))

		if entry.leaf_type == NORMAL {
			fmt.Printf(entry.text)
		}
	}

	fmt.Printf("\"\n\n")

	for _, entry := range list {
		fmt.Printf("    %s\n", entry)
	}

	fmt.Println()
}*/

func debug_title_page(title map[string]string) {
	for _, c := range debug_title_order {
		if x, ok := title[c]; ok {
			fmt.Printf("%s\n%s\n\n", c, x)
		}
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

func debug_print_screenplay(content *fountain_content) {
	for _, node := range content.nodes {
		fmt.Println(node)
	}
}

func debug_fonts(input_name string, results map[string]string) {
	fmt.Printf("requested:   %s\nin path:     %s\n\n", input_name, strings.Join(system_dirs, ";"))

	for _, weight := range []string{"regular", "bold", "italic", "bolditalic"} {
		if x, ok := results[weight]; ok {
			fmt.Printf("%-12s %s\n", weight, x)
		}
	}
}