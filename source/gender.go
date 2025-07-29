/*
	Meander
	A portable Fountain utility for production writing
	Copyright (C) 2022-2023 Harley Denham
*/

package main

import "fmt"
import "math"
import "sort"
import "strings"

const BAR_LENGTH = 20

func command_gender(config *Config) {
	text, success := merge(config.source_file)
	if !success {
		return
	}

	data := init_data(config)
	syntax_parser(config, data, text)

	println_color("\n   ", clean_string(data.Title.Title), "Gender Analysis")
	print_data(crunch_chars_by_gender(data), "Character Count by Gender")
	print_data(crunch_lines_by_gender(data), "Lines by Gender")
	print_data(crunch_chars_by_lines(data),  "Lines by Character")
	print("\n")
}

type Analytics_Set struct {
	longest_name_one int
	longest_name_two int

	total_value   int
	largest_value int

	data []Analytics_Entry
}

type Analytics_Entry struct {
	value int

	name_one string
	name_two string
}

type Analytics_Entries []Analytics_Entry
func (oc Analytics_Entries) Len() int           { return len(oc) }
func (oc Analytics_Entries) Less(i, j int) bool { return oc[i].value > oc[j].value }
func (oc Analytics_Entries) Swap(i, j int)      { oc[i], oc[j] = oc[j], oc[i] }

func crunch_chars_by_gender(data *Fountain) *Analytics_Set {
	array   := make(Analytics_Entries, 0, 64)
	counter := make(map[string]int,       64)

	total_chars    := 0
	longest_gender := 0

	for _, c := range data.Characters {
		if c.Gender == "ignore" {
			continue
		}

		total_chars += 1
		counter[c.Gender] += 1

		x := rune_count(c.Gender)
		if x > longest_gender {
			longest_gender = x
		}
	}

	largest_group := 0

	for gender_name, count := range counter {
		if count > largest_group {
			largest_group = count
		}
		array = append(array, Analytics_Entry{
			value:    count,
			name_one: gender_name,
		})
	}

	sort.Sort(array)

	return &Analytics_Set{
		longest_gender,
		0,
		total_chars,
		largest_group,
		array,
	}
}

func crunch_lines_by_gender(data *Fountain) *Analytics_Set {
	array := make(Analytics_Entries, 0, 12)

	total_lines    := 0
	longest_gender := 0

	counter := make(map[string]int, 12)

	for _, c := range data.Characters {
		if c.Gender == "ignore" {
			continue
		}

		total_lines += c.Lines
		counter[c.Gender] += c.Lines

		x := rune_count(c.Gender)
		if x > longest_gender {
			longest_gender = x
		}
	}

	largest_group := 0

	for gender_name, count := range counter {
		if count > largest_group {
			largest_group = count
		}
		array = append(array, Analytics_Entry{
			value:    count,
			name_one: gender_name,
		})
	}

	sort.Sort(array)

	return &Analytics_Set{
		longest_gender,
		0,
		total_lines,
		largest_group,
		array,
	}
}

func crunch_chars_by_lines(data *Fountain) *Analytics_Set {
	array := make(Analytics_Entries, 0, len(data.Characters))

	total_lines    := 0
	most_lines     := 0
	longest_gender := 0
	longest_char   := 0

	for _, c := range data.Characters {
		if c.Gender == "ignore" {
			continue
		}

		total_lines += c.Lines

		if c.Lines > most_lines {
			most_lines = c.Lines
		}

		x := rune_count(c.Name)
		if x > longest_char {
			longest_char = x
		}
		y := rune_count(c.Gender)
		if y > longest_gender {
			longest_gender = y
		}

		array = append(array, Analytics_Entry{
			value:    c.Lines,
			name_one: c.Name,
			name_two: c.Gender,
		})
	}

	sort.Sort(array)

	return &Analytics_Set{
		longest_char,
		longest_gender,
		total_lines,
		most_lines,
		array,
	}
}

func print_dashes(n int) {
	println(strings.Repeat("-", n))
}

func println_color(text ...string) {
	if !running_in_term {
		println(text...)
		return
	}

	print(ANSI_COLOR)
	println(text...)
	print(ANSI_RESET)
}

func print_padded(t string, x int) {
	x = x - rune_count(t) + 2
	print(t)
	if x > 0 {
		print(strings.Repeat(" ", x))
	}
}

func print_data(data *Analytics_Set, title string) {
	print("\n    ")
	println_color(title)

	offset := 39

	if data.longest_name_two > 0 {
		offset += 2
	}

	print("    ")
	print_dashes(data.longest_name_one + data.longest_name_two + offset)

	data_total   := float64(data.total_value)
	data_largest := float64(data.largest_value)

	for _, entry := range data.data {
		if entry.value == 0 {
			continue
		}

		print("    ")
		print_padded(title_case(entry.name_one), data.longest_name_one)

		if data.longest_name_two > 0 {
			print_padded(title_case(entry.name_two), data.longest_name_two)
		}

		print_padded(fmt.Sprintf("%d", entry.value), 5)

		the_value  := float64(entry.value)
		percentage := the_value / data_total * 100

		print_padded(fmt.Sprintf("%.1f%%", percentage), 8)
		println_color(strings.Repeat("|", int(math.Round(the_value / data_largest * BAR_LENGTH))))
	}
}
