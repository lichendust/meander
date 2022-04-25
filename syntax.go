package main

type fountain_content struct {
	title map[string]string
	nodes []*syntax_node
}

// node type
const (
	WHITESPACE uint8 = iota
	PAGE_BREAK
	HEADER
	FOOTER
	PAGENUMBER
	SCENE_NUMBER

	ACTION
	LIST
	SCENE
	CHARACTER
	PARENTHETICAL
	DIALOGUE
	LYRIC
	TRANSITION
	CENTERED
	SYNOPSIS

	SECTION
	SECTION2
	SECTION3
)

type syntax_node struct {
	node_type  uint8
	level      uint8
	revised    bool
	template   *template_entry

	raw_text   string

	lines      []*syntax_line
}