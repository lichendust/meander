package main

import (
	"os"
	"fmt"
	"encoding/json"
)

type JSONType struct {
	Type    string `json:"type"`
	Revised bool   `json:"revised,omitempty"`
	Text    string `json:"text"`
}

var node_type_convert = map[uint8]string{
	WHITESPACE:    "whitespace",
	PAGE_BREAK:    "page_break",
	HEADER:        "header",
	FOOTER:        "footer",
	PAGE_NUMBER:   "page_number",
	SCENE_NUMBER:  "scene_number",
	ACTION:        "action",
	LIST:          "list",
	SCENE:         "scene",
	CHARACTER:     "character",
	PARENTHETICAL: "parenthetical",
	DIALOGUE:      "dialogue",
	LYRIC:         "lyric",
	TRANSITION:    "transition",
	CENTERED:      "centered",
	SYNOPSIS:      "synopsis",
	SECTION:       "section",
}

func convert_to_json(nodes []*syntax_node) {
	array := make([]*JSONType, 0, len(nodes))

	for _, node := range nodes {
		if node.node_type == WHITESPACE {
			continue
		}
		array = append(array, &JSONType {
			Type:    node_type_convert[node.node_type],
			Revised: node.revised,
			Text:    node.raw_text,
		})
	}

	b, err := json.Marshal(array)

	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to marshal JSON")
	}

	os.Stdout.Write(b)
}

func command_json(config *config) {
	content, ok := syntax_parser(config)

	if !ok {
		fmt.Fprintf(os.Stderr, "failed to merge file %s\n", config.source_file)
		return
	}

	convert_to_json(content.nodes)
}