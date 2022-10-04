package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"encoding/json"
)

type JSON_File struct {
	Title struct {
		Title     string `json:"title,omitempty"`
		Credit    string `json:"credit,omitempty"`
		Author    string `json:"author,omitempty"`
		Source    string `json:"source,omitempty"`
		Notes     string `json:"notes,omitempty"`
		DraftDate string `json:"draft_date,omitempty"`
		Copyright string `json:"copyright,omitempty"`
		Revision  string `json:"revision,omitempty"`
		Contact   string `json:"contact,omitempty"`
	} `json:"title"`
	Characters []*Gender_Char `json:"characters,omitempty"`
	Content    []*JSON_MarkupLine
}

type JSON_MarkupLine struct {
	Revised     bool     `json:"revised,omitempty"`
	Type        string   `json:"type"`
	Text        string   `json:"text,omitempty"`
	SceneNumber string   `json:"scene_number,omitempty"`
	Character   string   `json:"name,omitempty"`
	Dialogue    []string `json:"dialogue,omitempty"`
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

func command_json(config *config) {
	content, ok := syntax_parser(config)

	if !ok {
		fmt.Fprintf(os.Stderr, "failed to merge file %s\n", config.source_file)
		return
	}

	json_file := &JSON_File{}

	// title page
	{
		json_file.Title.Title     = content.title["title"]
		json_file.Title.Credit    = content.title["credit"]
		json_file.Title.Author    = content.title["author"]
		json_file.Title.Source    = content.title["source"]
		json_file.Title.Notes     = content.title["notes"]
		json_file.Title.DraftDate = content.title["draft_date"]
		json_file.Title.Copyright = content.title["copyright"]
		json_file.Title.Revision  = content.title["revision"]
		json_file.Title.Contact   = content.title["contact"]
	}

	// gender
	{
		_, data, _, ok := do_full_analysis(config)

		if !ok {
			return
		}

		length := 0

		for _, group := range data.gender_list {
			length += len(group.characters)
		}

		array := make([]*Gender_Char, 0, length)

		for _, group := range data.gender_list {
			for _, char := range group.characters {
				char.AllNames = char.AllNames[1:] // drop the duplicate of the main name
				array = append(array, char)
			}
		}

		json_file.Characters = array
	}

	// content
	{
		array := make([]*JSON_MarkupLine, 0, len(content.nodes))

		for _, node := range content.nodes {
			switch node.node_type {
			case WHITESPACE:
				continue

			case SCENE_NUMBER:
				array[len(array) - 1].SceneNumber = node.raw_text
				continue

			case DIALOGUE, PARENTHETICAL, LYRIC:
				n := len(array) - 1
				array[n].Dialogue = append(array[n].Dialogue, node.raw_text)
				continue

			case CHARACTER:
				array = append(array, &JSON_MarkupLine {
					Type:      node_type_convert[node.node_type],
					Revised:   node.revised,
					Character: node.raw_text,
				})
				continue
			}

			array = append(array, &JSON_MarkupLine {
				Type:    node_type_convert[node.node_type],
				Revised: node.revised,
				Text:    node.raw_text,
			})
		}

		json_file.Content = array
	}

	// write the file
	b, err := json.MarshalIndent(json_file, "", "\t")

	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to marshal JSON")
		return
	}

	err = ioutil.WriteFile(config.output_file, b, 0777)

	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to write out JSON")
		return
	}
}