package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"encoding/json"
)

type JSON_File struct {
	Meta struct {
		Source  string `json:"source"`
		Version uint16 `json:"version"`
	} `json:"meta"`
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
	Content    []*JSON_MarkupLine `json:"content"`
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

const remove_format_chars = true

func strip_on_condition(input string, is_stripped bool) string {
	if input == "" {
		return ""
	}
	if is_stripped {
		return clean_string(input)
	}
	return input
}

func command_json(config *config) {
	content, ok := syntax_parser(config)

	if !ok {
		fmt.Fprintf(os.Stderr, "failed to merge file %s\n", config.source_file)
		return
	}

	json_file := &JSON_File{}

	// meta
	{
		version_number, _ := make_version_number(title)

		json_file.Meta.Source  = title
		json_file.Meta.Version = version_number
	}

	// title page
	{
		json_file.Title.Title     = strip_on_condition(content.title["title"],      remove_format_chars)
		json_file.Title.Credit    = strip_on_condition(content.title["credit"],     remove_format_chars)
		json_file.Title.Author    = strip_on_condition(content.title["author"],     remove_format_chars)
		json_file.Title.Source    = strip_on_condition(content.title["source"],     remove_format_chars)
		json_file.Title.Notes     = strip_on_condition(content.title["notes"],      remove_format_chars)
		json_file.Title.DraftDate = strip_on_condition(content.title["draft date"], remove_format_chars)
		json_file.Title.Copyright = strip_on_condition(content.title["copyright"],  remove_format_chars)
		json_file.Title.Revision  = strip_on_condition(content.title["revision"],   remove_format_chars)
		json_file.Title.Contact   = strip_on_condition(content.title["contact"],    remove_format_chars)
	}

	// gender
	{
		_, data, _, ok := do_full_analysis(config)

		if ok {
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
				array[n].Dialogue = append(array[n].Dialogue, strip_on_condition(node.raw_text, remove_format_chars))
				continue

			case CHARACTER:
				array = append(array, &JSON_MarkupLine {
					Type:      node_type_convert[node.node_type],
					Revised:   node.revised,
					Character: strip_on_condition(node.raw_text, remove_format_chars),
				})
				continue
			}

			array = append(array, &JSON_MarkupLine {
				Type:    node_type_convert[node.node_type],
				Revised: node.revised,
				Text:    strip_on_condition(node.raw_text, remove_format_chars),
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