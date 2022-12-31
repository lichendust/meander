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
	"fmt"
	"math"
	"sort"
	"bufio"
	"strings"
)

type Gender_Data struct {
	longest_gender int
	longest_char   int
	gender_list []*Gender_Group
}

type Gender_Group struct {
	gender_name  string
	longest_char int
	characters   []*Gender_Char
}

type Gender_Char struct {
	Name       string   `json:"name"`
	AllNames   []string `json:"other_names,omitempty"`
	Gender     string   `json:"gender"`
	LineCount  int      `json:"lines_spoken,omitempty"`
}

func command_gender_analysis(config *config) {
	content, data, has_updated, ok := do_full_analysis(config)

	if !ok {
		return
	}

	// update the table in the source file
	if has_updated && config.write_gender {
		if ok := gender_replace_comment(config, data); !ok {
			eprintln("failed to replace gender table!")
			return
		}
	}

	/*
		@efficiency @todo we parse the files three(?) times from
		disk in order to print gender analysis onto the PDF
		itself. The preprocessor has always been nested inside
		the target functions, including gender, but when I
		decided to print onto the PDF, all that exploded.  We
		should fix that.

		I've left this here even though it's not technically
		occurring in this function; rather it's a systemic issue.
	*/

	if running_in_term {
		newline()
	}

	{
		const gender_title     = " Gender Analysis"
		const gender_title_len = len(gender_title)

		title, ok := content.title["title"]

		if ok {
			title = clean_string(title)
		} else {
			title = config.source_file
		}

		{
			if running_in_term {
				print(ansi_color_accent)
			}

			print(title); println(gender_title)

			if running_in_term {
				print(ansi_color_reset)
			}
		}

		println(strings.Repeat("-", count_all_runes(title) + gender_title_len))
		newline()
	}

	print_data(crunch_chars_by_gender(data), "Character Count by Gender")
	newline()
	print_data(crunch_lines_by_gender(data), "Lines by Gender")
	newline()
	print_data(crunch_chars_by_lines(data), "Characters by Lines Spoken")

	if running_in_term {
		newline()
	}
}

func do_full_analysis(config *config) (*fountain_content, *Gender_Data, bool, bool) {
	// read comment from input file
	the_comment, ok := gender_table_parser(config)

	if !ok {
		return nil, nil, false, false
	}

	// swizzle the character table so the Name
	// is the lookup for the character's data,
	// sharing memory addresses with the original tree
	lookup_table := char_swizzle(the_comment)

	// parse the full screenplay
	content, ok := syntax_parser(config)

	if !ok {
		return nil, nil, false, false
	}

	has_changes := false
	unknown_group_size := 0

	{
		// if the unknown group exists, store a pointer
		// to it for updating new members from the text
		var unknown_group *Gender_Group

		for _, group := range the_comment.gender_list {
			if group.gender_name == "unknown" {
				unknown_group = group
				break
			}
		}

		// if there isn't an unknown group
		// already, create a new one
		if unknown_group == nil {
			unknown_group = &Gender_Group {
				characters:  make([]*Gender_Char, 0, 32),
				gender_name: "unknown",
			}
			the_comment.gender_list = append(the_comment.gender_list, unknown_group)
		}

		unknown_group_size = len(unknown_group.characters)

		// seek through nodes
		has_any_character := false

		for _, node := range content.nodes {
			if node.node_type == CHARACTER {
				has_any_character = true

				for i, c := range node.raw_text {
					if c == '(' {
						node.raw_text = strings.TrimSpace(node.raw_text[:i])
						break
					}
				}

				lower := strings.ToLower(node.raw_text)

				if _, ok := lookup_table[lower]; ok {
					lookup_table[lower].LineCount += 1
				} else {
					unknown_char := &Gender_Char {
						Name:   lower,
						AllNames:   []string{lower},
						Gender: "unknown",
						LineCount:  1,
					}

					lookup_table[lower] = unknown_char
					unknown_group.characters = append(unknown_group.characters, unknown_char)
				}
			}

			has_changes = len(unknown_group.characters) != unknown_group_size // check if we added anyone to "unknown"
		}

		if !has_any_character {
			eprintln("gender: no character data to display!")
			return nil, nil, false, false
		}
	}

	return content, the_comment, has_changes, true
}

func gender_table_parser(config *config) (*Gender_Data, bool) {
	text, ok := load_file(fix_path(config.source_file))

	if !ok {
		eprintln("gender: failed to load", config.source_file)
		return nil, false
	}

	for len(text) > 0 {
		if text[0] == '/' && len(text) > 1 && text[1] == '*' {
			text = text[2:]

			n := rune_pair(text, '*', '/')

			if n < 0 {
				continue
			}

			if strings.HasPrefix(consume_whitespace(text), "[gender") {
				text = strings.TrimSpace(text[:n - 2])
				break
			}
		}
		text = text[1:]
	}

	the_comment := &Gender_Data {
		gender_list: make([]*Gender_Group, 0, 4),
	}
	the_group := &Gender_Group {}

	buffer := bufio.NewReader(strings.NewReader(text))

	for {
		line, err := buffer.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				if line == "" {
					break
				}
			} else {
				eprintln("gender: error reading comment string")
				return nil, false
			}
		}

		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if line[0] == '[' {
			if line[len(line) - 1] != ']' {
				eprintln("malformed gender heading", line)
				return nil, false
			}

			inner_line := strings.ToLower(strings.TrimSpace(line[1:len(line) - 1]))

			if !strings.HasPrefix(inner_line, "gender.") {
				eprintln("expected \"[gender.<term>]\" instead of", line)
				return nil, false
			}

			gender_name := inner_line[7:]

			the_group = &Gender_Group {
				characters:  make([]*Gender_Char, 0, 32),
				gender_name: gender_name,
			}

			if n := count_all_runes(gender_name); n > the_comment.longest_gender {
				the_comment.longest_gender = n
			}

			the_comment.gender_list = append(the_comment.gender_list, the_group)
			continue
		}

		name  := ""
		names := strings.Split(line, "|")

		for i, entry := range names {
			names[i] = strings.TrimSpace(entry)
		}

		name = names[0]

		char := &Gender_Char {
			Name:   name,
			AllNames:   names,
			Gender: the_group.gender_name,
		}

		if n := count_all_runes(name); n > the_group.longest_char {
			the_group.longest_char = n
		}

		the_group.characters = append(the_group.characters, char)
	}

	for _, group := range the_comment.gender_list {
		if group.gender_name == "ignore" {
			continue
		}
		if group.longest_char > the_comment.longest_char {
			the_comment.longest_char = group.longest_char
		}
	}

	if the_comment.longest_gender < 7 {
		the_comment.longest_gender = 7
	}

	if text == "" {
		// @todo add write confirmation
		eprintln("gender: no table found in", config.source_file)
		return nil, false
	}

	return the_comment, true
}

// remap the character data into a map,
// ensuring every character is addressable
// by every variant of their name
func char_swizzle(the_comment *Gender_Data) map[string]*Gender_Char {
	new_map := make(map[string]*Gender_Char, len(the_comment.gender_list) * 32)

	for _, gender := range the_comment.gender_list {
		for _, char := range gender.characters {
			for _, name := range char.AllNames {
				new_map[strings.ToLower(name)] = char
			}
		}
	}

	return new_map
}

// storage for the container
type data_container struct {
	longest_name_one int
	longest_name_two int

	total_value   int  // total value of all entries in ordered_data
	largest_value int  // largest value entry in ordered_data

	ordered_data  []*data_entry
}

type data_entry struct {
	value        int         // sortable pivot

	name_one     string
	name_two     string
}

type data_order []*data_entry

func (oc data_order) Len() int {
	return len(oc)
}
func (oc data_order) Less(i, j int) bool {
	return oc[i].value > oc[j].value
}
func (oc data_order) Swap(i, j int) {
	oc[i], oc[j] = oc[j], oc[i]
}

func print_data(data *data_container, title string) {
	if running_in_term {
		print(ansi_color_accent)
	}

	println(title)

	if running_in_term {
		print(ansi_color_reset)
	}

	offset := 39

	if data.longest_name_two > 0 {
		offset += 2
	}

	println(strings.Repeat("-", data.longest_name_one + data.longest_name_two + offset))

	for _, entry := range data.ordered_data {
		if entry.value == 0 {
			continue
		}

		print(space_pad_string(title_case(entry.name_one), data.longest_name_one))

		if data.longest_name_two > 0 {
			print(space_pad_string(title_case(entry.name_two), data.longest_name_two))
		}

		print(space_pad_string(fmt.Sprintf("%d", entry.value), 5))

		{
			percentage := float64(entry.value) / float64(data.total_value) * 100
			print(space_pad_string(fmt.Sprintf("%.1f%%", percentage), 8))
		}

		if running_in_term {
			print(ansi_color_accent)
		}

		{
			bar_graph := normalise(entry.value, data.largest_value, 20)
			print(strings.Repeat("▪", bar_graph))
		}

		if running_in_term {
			print(ansi_color_reset)
		}

		newline()
	}
}

func crunch_chars_by_lines(the_comment *Gender_Data) *data_container {
	data := make(data_order, 0, len(the_comment.gender_list) * 32)

	total_lines := 0
	most_lines  := 0

	for _, group := range the_comment.gender_list {
		if len(group.characters) == 0 {
			continue
		}

		if group.gender_name == "ignore" {
			continue
		}

		for _, char := range group.characters {
			if char.LineCount == 0 {
				continue
			}

			if char.LineCount > most_lines {
				most_lines = char.LineCount
			}

			total_lines += char.LineCount

			data = append(data, &data_entry {
				value:    char.LineCount,
				name_one: char.Name,
				name_two: char.Gender,
			})
		}
	}

	sort.Sort(data)

	return &data_container {
		the_comment.longest_char,
		the_comment.longest_gender,
		total_lines,
		most_lines,
		data,
	}
}

func crunch_lines_by_gender(the_comment *Gender_Data) *data_container {
	data := make(data_order, 0, len(the_comment.gender_list))

	total_lines := 0
	most_lines  := 0

	for _, group := range the_comment.gender_list {
		group_lines := 0

		if len(group.characters) == 0 {
			continue
		}

		if group.gender_name == "ignore" {
			continue
		}

		for _, char := range group.characters {
			if char.LineCount == 0 {
				continue
			}

			total_lines += char.LineCount
			group_lines += char.LineCount

			if group_lines > most_lines {
				most_lines = group_lines
			}
		}

		data = append(data, &data_entry {
			value:    group_lines,
			name_one: group.gender_name,
		})
	}

	sort.Sort(data)

	return &data_container {
		the_comment.longest_gender,
		0,
		total_lines,
		most_lines,
		data,
	}
}

func crunch_chars_by_gender(the_comment *Gender_Data) *data_container {
	data := make(data_order, 0, len(the_comment.gender_list))

	total_chars   := 0
	largest_group := 0

	for _, group := range the_comment.gender_list {
		if len(group.characters) == 0 {
			continue
		}

		if group.gender_name == "ignore" {
			continue
		}

		group_chars := 0

		for _, char := range group.characters {
			if char.LineCount > 0 {
				group_chars += 1
			}
		}

		total_chars += group_chars

		if group_chars > largest_group {
			largest_group = group_chars
		}

		data = append(data, &data_entry {
			value:    group_chars,
			name_one: group.gender_name,
		})
	}

	sort.Sort(data)

	return &data_container {
		the_comment.longest_gender,
		0,
		total_chars,
		largest_group,
		data,
	}
}

// this function rewrites the raw comment in a sensible layout,
// specifically mirroring the user's layout (order of genders
// and chars, as well as casing of the names) for re-insertion
// into the original script text. it does not do the parsing,
// just provides the actual comment, with any additional
// discoveries written in
func (the_comment *Gender_Data) String() string {
	buffer := strings.Builder {}
	buffer.Grow(len(the_comment.gender_list) * 32 * 32)

	buffer.WriteString("/*")

	for _, group := range the_comment.gender_list {
		if len(group.characters) == 0 {
			continue
		}

		do_title := group.gender_name == "unknown"

		buffer.WriteString(fmt.Sprintf("\n\t[gender.%s]\n", group.gender_name))

		for _, char := range group.characters {
			if do_title {
				for i, name := range char.AllNames {
					char.AllNames[i] = title_case(name)
				}
			}

			buffer.WriteString(fmt.Sprintf("\t%s\n", strings.Join(char.AllNames, " | ")))
		}
	}

	buffer.WriteString("*/")

	return buffer.String()
}

// takes the input file, strips the original comment out, then
// rewrites the data as close to the original as possible while
// updating with new
func gender_replace_comment(config *config, the_comment *Gender_Data) bool {
	filepath := fix_path(config.source_file)
	text, ok := load_file(filepath)

	if !ok {
		return false
	}

	input := the_comment.String()

	buffer := strings.Builder {}
	buffer.Grow(len(text) + len(input))

	copy := text

	starting_byte := 0
	ending_byte   := 0

	for len(text) > 0 {
		if text[0] == '/' && len(text) > 1 && text[1] == '*' {
			text = text[2:]

			n := rune_pair(text, '*', '/')

			if n < 0 {
				continue
			}

			if strings.HasPrefix(consume_whitespace(text), "[gender") {
				ending_byte = starting_byte + n + 2
				break
			}
		}

		starting_byte+= 1
		text = text[1:]
	}

	buffer.WriteString(copy[:starting_byte])
	buffer.WriteString(input)
	buffer.WriteString(copy[ending_byte:])

	return write_file(filepath, []byte(buffer.String()))
}

func normalise(value, largest, max int) int {
	return int(math.Round(float64(value) / float64(largest) * float64(max)))
}