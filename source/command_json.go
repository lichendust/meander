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

import "encoding/json"

type JSON_File struct {
	Meta struct {
		Source  string `json:"source"`
		Version uint8  `json:"version"`
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
	CHARACTER:     "dialogue", // CHARACTER is the starting point for a "dialogue" json entry
	LYRIC:         "lyric",
	TRANSITION:    "transition",
	CENTERED:      "centered",
	SYNOPSIS:      "synopsis",
	SECTION:       "section",
}

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
		eprintln("failed to merge file", config.source_file)
		return
	}

	remove_format_chars := !config.json_keep_formatting

	json_file := &JSON_File{}

	// meta
	{
		json_file.Meta.Source  = title
		json_file.Meta.Version = 1
	}

	// title page
	{
		/*
			this is nuts, obviously, but we do it to make the
			output of the title page idempotent. Go has no
			way of guaranteeing map order, and even if
			we manually sorted them before copying, the JSON
			package will go do whatever it wants after that.
		*/
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
		eprintln("failed to marshal JSON")
		return
	}

	ok = write_file(config.output_file, b)
	if !ok {
		eprintln("failed to write", config.output_file)
	}
}