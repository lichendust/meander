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

import "fmt"
import "math"
import "strings"

import lib "github.com/signintech/gopdf"

func set_color(doc *lib.GoPdf, color Color) {
	doc.SetFillColor(color.R, color.G, color.B)
}

func set_font(doc *lib.GoPdf, style Leaf_Type) {
	is_bold := style & BOLD != 0
	is_ital := style & ITALIC != 0

	format := ""

	if is_bold && is_ital {
		format = "BI"
	} else if is_bold {
		format = "B"
	} else if is_ital {
		format = "I"
	}

	doc.SetFont(RESERVED_NAME, format, FONT_SIZE)
}

func command_render(config *Config) {
	text, ok := merge(config.source_file)
	if !ok {
		return
	}

	data := init_data()

	syntax_parser(config, data, text)

	if len(config.paper_size) == 0 {
		data.paper = lib.PageSizeLetter
	} else {
		data.paper = set_paper(config.paper_size)
	}

	format := SCREENPLAY
	if len(config.template) > 0 {
		if f, ok := is_valid_format(config.template); ok {
			format = f
		}
	}
	data.template = build_template(config, format, *data.paper)

	if data.template.landscape {
		data.paper.W, data.paper.H = data.paper.H, data.paper.W
	}

	paginate(config, data)

	doc := new(lib.GoPdf)

	doc.Start(lib.Config{
		PageSize: *data.paper,
	})
	doc.SetInfo(lib.PdfInfo{
		Title:        clean_string(data.Title.Title),
		Author:       clean_string(data.Title.Author),
		Creator:      MEANDER,
		CreationDate: now(),
	})

	register_fonts(doc)
	set_font(doc, NO_TYPE)

	render_title(config, data, doc)
	render_gender(config, data, doc)
	render_toc(config, data, doc)
	render_content(config, data, doc)

	if err := doc.WritePdf(fix_path(config.output_file)); err != nil {
		eprintln("error saving", config.output_file)
	}
}

func render_title(config *Config, data *Fountain, doc *lib.GoPdf) {
	if !data.Title.has_any {
		return
	}

	const LINE_HEIGHT = LINE_HEIGHT * 1.5

	const WIDTH  = INCH * 3 // bottom corners
	title_width := INCH * 4 // main title

	start_x := INCH
	start_y := MARGIN_TOP + PICA * 15

	align := data.template.title_page_align
	if align == CENTER {
		start_x = data.template.center_line
	} else {
		title_width = WIDTH
	}

	doc.AddPage()

	// Title    Credit    Author    Source
	{
		title  := quick_section(data, data.Title.Title,  align, LINE_HEIGHT, title_width)
		credit := quick_section(data, data.Title.Credit, align, LINE_HEIGHT, title_width)
		source := quick_section(data, data.Title.Source, align, LINE_HEIGHT, title_width)
		author := quick_section(data, data.Title.Author, align, LINE_HEIGHT, title_width)

		if data.Title.Title != "" {
			title.pos_x = start_x
			title.pos_y = start_y
			draw_section(doc, data.template, title)
			start_y += title.total_height + LINE_HEIGHT * 4
		}

		if data.Title.Credit != "" {
			credit.pos_x = start_x
			credit.pos_y = start_y
			draw_section(doc, data.template, credit)
			start_y += credit.total_height + LINE_HEIGHT * 2
		}

		if data.Title.Author != "" {
			author.pos_x = start_x
			author.pos_y = start_y
			draw_section(doc, data.template, author)
			start_y += author.total_height + LINE_HEIGHT * 4
		}

		if data.Title.Source != "" {
			source.pos_x = start_x
			source.pos_y = start_y
			draw_section(doc, data.template, source)
		}
	}

	start_x = INCH

	// Notes    Copyright    Contact
	{
		note := quick_section(data, data.Title.Notes,     LEFT, LINE_HEIGHT, WIDTH)
		cont := quick_section(data, data.Title.Contact,   LEFT, LINE_HEIGHT, WIDTH)
		copy := quick_section(data, data.Title.Copyright, LEFT, LINE_HEIGHT, WIDTH)

		start_y = data.paper.H - INCH - (note.total_height + copy.total_height + cont.total_height) + LINE_HEIGHT

		if data.Title.Notes != "" {
			note.pos_x = start_x
			note.pos_y = start_y
			draw_section(doc, data.template, note)
			start_y += note.total_height
		}

		if data.Title.Contact != "" {
			cont.pos_x = start_x
			cont.pos_y = start_y
			draw_section(doc, data.template, cont)
			start_y += cont.total_height
		}

		if data.Title.Copyright != "" {
			copy.pos_x = start_x
			copy.pos_y = start_y
			draw_section(doc, data.template, copy)
		}
	}

	start_x = data.paper.W - INCH

	// Revision    DraftDate    Info
	{
		revs := quick_section(data, data.Title.Revision,  RIGHT, LINE_HEIGHT, WIDTH)
		drft := quick_section(data, data.Title.DraftDate, RIGHT, LINE_HEIGHT, WIDTH)
		info := quick_section(data, data.Title.Info,      RIGHT, LINE_HEIGHT, WIDTH)

		start_y = data.paper.H - INCH - (revs.total_height + drft.total_height + info.total_height) + LINE_HEIGHT

		if data.Title.Revision != "" {
			revs.pos_x = start_x
			revs.pos_y = start_y
			draw_section(doc, data.template, revs)
			start_y += revs.total_height + LINE_HEIGHT
		}

		if data.Title.DraftDate != "" {
			drft.pos_x = start_x
			drft.pos_y = start_y
			draw_section(doc, data.template, drft)
			start_y += drft.total_height + LINE_HEIGHT
		}

		if data.Title.Info != "" {
			info.pos_x = start_x
			info.pos_y = start_y
			draw_section(doc, data.template, info)
		}
	}
}

func render_gender(config *Config, data *Fountain, doc *lib.GoPdf) {
	if !config.include_gender {
		return
	}

	doc.AddPage()

	start_y := MARGIN_TOP

	doc.SetXY(MARGIN_LEFT, start_y)

	{
		gender_title := fmt.Sprintf("%q %s", clean_string(data.Title.Title), GENDER_HEADING)
		length := rune_count(gender_title)
		t := Line{
			length: length,
			leaves: []Leaf{{NORMAL, false, gender_title}},
		}
		line_override(&t, UNDERLINE)
		draw_line(doc, data.template, &t, MARGIN_LEFT, start_y)
	}

	start_y += LINE_HEIGHT * 2

	render_gender_data(data, doc, crunch_chars_by_gender(data), GENDER_CHARS_BY_GENDER, &start_y)
	start_y += LINE_HEIGHT
	render_gender_data(data, doc, crunch_lines_by_gender(data), GENDER_LINES_BY_GENDER, &start_y)
	start_y += LINE_HEIGHT
	render_gender_data(data, doc, crunch_chars_by_lines(data),  GENDER_CHARS_BY_LINES,  &start_y)
}

func render_toc(config *Config, data *Fountain, doc *lib.GoPdf) {
	if !config.table_of_contents {
		return
	}

	doc.AddPage()

	running_x := MARGIN_LEFT
	running_y := MARGIN_TOP

	widest_scene_no := 0
	for i := range data.Content {
		section := data.Content[i]
		if section.Type == SCENE {
			w := rune_count(section.SceneNumber)
			if w > widest_scene_no {
				widest_scene_no = w
			}
		}
	}
	if widest_scene_no > 0 { widest_scene_no += 3 }
	scene_inset := CHAR_WIDTH * float64(widest_scene_no)

	for i := range data.Content {
		section := data.Content[i]

		if section.Type > is_section || section.Type == SCENE {
			new_section := copy_without_style(section)

			if section.Type > is_section && data.template.kind == SCREENPLAY {
				running_y += LINE_HEIGHT
				style_override(&new_section, UNDERLINE)
			}

			new_section.justify = LEFT

			new_section.pos_x = running_x
			new_section.pos_y = running_y

			if section.Type == SCENE {
				doc.SetXY(running_x, running_y)
				doc.Text(section.SceneNumber)
				new_section.pos_x += scene_inset
			}

			draw_section(doc, data.template, &new_section)

			page_number := fmt.Sprintf("%d", new_section.page)
			doc.SetX(data.paper.W - MARGIN_RIGHT - float64(rune_count(page_number)) * CHAR_WIDTH)
			doc.Text(page_number)

			set_font(doc, NO_TYPE)

			if section.Type > is_section {
				running_y += LINE_HEIGHT
			}

			running_y += LINE_HEIGHT
		}

		if running_y > data.paper.H - MARGIN_BOTTOM {
			doc.AddPage()
			running_y = MARGIN_TOP
		}
	}
}

func render_content(config *Config, data *Fountain, doc *lib.GoPdf) {
	if len(data.Content) == 0 {
		return
	}

	page_number := 0

	for i := range data.Content {
		section := &data.Content[i]

		if section.skip {
			continue
		}

		if section.page > page_number {
			page_number = section.page
			doc.AddPage()
		}

		if section.Type == SCENE && config.scenes != SCENE_REMOVE {
			text_width := float64(rune_count(section.SceneNumber)) * CHAR_WIDTH
			right_x    := data.paper.W - MARGIN_RIGHT - text_width
			left_x     := MARGIN_LEFT - INCH / 2 - text_width

			set_font(doc, NO_TYPE)

			doc.SetY(section.pos_y)
			doc.SetX(left_x)
			doc.Text(section.SceneNumber)

			draw_section(doc, data.template, section)

			doc.SetX(right_x)
			doc.Text(section.SceneNumber)

			continue
		}

		draw_section(doc, data.template, section)
	}
}

func draw_section(doc *lib.GoPdf, template *Template, section *Section) {
	if section == nil {
		return
	}
	if section.Text == "" {
		return
	}

	set_color(doc, template.text_color)

	if section.is_raw {
		set_font(doc, NO_TYPE)

		pos_x := section.pos_x

		switch section.justify {
		case CENTER:
			pos_x -= CHAR_WIDTH * float64(section.longest_line / 2)
		case RIGHT:
			pos_x -= CHAR_WIDTH * float64(section.longest_line)
		}

		doc.SetXY(pos_x, section.pos_y)
		doc.Text(section.Text)
		return
	}

	pos_x := section.pos_x + section.para_indent
	pos_y := section.pos_y

	for i, line := range section.lines {
		if len(line.leaves) == 0 {
			pos_y += section.line_height
			continue
		}

		if i > 0 {
			pos_x = section.pos_x
		}

		switch section.justify {
		case CENTER:
			pos_x -= CHAR_WIDTH * float64(line.length / 2)
		case RIGHT:
			pos_x -= CHAR_WIDTH * float64(line.length)
		}

		draw_line(doc, template, &line, pos_x, pos_y)
		pos_y += section.line_height
	}
}

func draw_line(doc *lib.GoPdf, template *Template, line *Line, pos_x, pos_y float64) {
	doc.SetXY(pos_x, pos_y)

	if len(line.highlight) > 0 {
		set_color(doc, template.highlight_color)

		draw_range_item(line.highlight, func(a, b float64) {
			y := pos_y - PICA + 2
			doc.Rectangle(pos_x + a - 2, y, pos_x + b + 2, y + FONT_SIZE + 2, "F", 0, 0)
		})

		set_color(doc, template.text_color)
	}

	do_bold_line := false
	for _, leaf := range line.leaves {
		if leaf.leaf_type & BOLD != 0 {
			do_bold_line = true
			break
		}
	}

	if do_bold_line {
		doc.SetLineWidth(2)
	} else {
		doc.SetLineWidth(1)
	}

	draw_range_item(line.underline, func(a, b float64) {
		y := pos_y + 2.5
		doc.Line(pos_x + a, y, pos_x + b, y)
	})
	draw_range_item(line.strikeout, func(a, b float64) {
		y := pos_y - PICA / 4
		doc.Line(pos_x + a, y, pos_x + b, y)
	})

	for _, leaf := range line.leaves {
		if leaf.leaf_type & NOTE != 0 {
			set_color(doc, template.note_color)
		} else {
			set_color(doc, template.text_color)
		}

		set_font(doc, line.style_reset | leaf.leaf_type)
		doc.Text(leaf.text)
	}
}

func draw_range_item(r []int, f func(a, b float64)) {
	for i := 0; i < len(r) - 1; i += 2 {
		a := r[i]
		b := r[i + 1]

		pos_a := float64(a) * CHAR_WIDTH
		pos_b := float64(b - a) * CHAR_WIDTH + pos_a

		f(pos_a, pos_b)
	}
}

func render_gender_data(data *Fountain, doc *lib.GoPdf, data_set *Analytics_Set, title string, start_y *float64) {
	data_total   := float64(data_set.total_value)
	data_largest := float64(data_set.largest_value)

	{
		t := Line{leaves:[]Leaf{{ITALIC, false, title}}}
		draw_line(doc, data.template, &t, MARGIN_LEFT, *start_y)
	}

	set_font(doc, NO_TYPE)

	*start_y += LINE_HEIGHT * 2

	for _, entry := range data_set.data {
		if entry.value == 0 {
			continue
		}

		running_x := MARGIN_LEFT

		doc.SetXY(running_x, *start_y)

		running_x += float64(data_set.longest_name_one + 2) * CHAR_WIDTH
		doc.Text(title_case(entry.name_one))
		doc.SetX(running_x)

		if data_set.longest_name_two > 0 {
			running_x += float64(data_set.longest_name_two + 2) * CHAR_WIDTH
			doc.Text(title_case(entry.name_two))
			doc.SetX(running_x)
		}

		text := fmt.Sprintf("%d", entry.value)
		running_x += CHAR_WIDTH * 7

		doc.Text(text)
		doc.SetX(running_x)

		the_value  := float64(entry.value)
		percentage := the_value / data_total * 100

		text = fmt.Sprintf("%.1f%%", percentage)
		running_x += CHAR_WIDTH * 7

		doc.Text(text)
		doc.SetX(running_x)
		doc.Text(strings.Repeat("|", int(math.Round(the_value / data_largest * BAR_LENGTH))))

		*start_y += LINE_HEIGHT
	}
}

func quick_section(data *Fountain, text string, justify uint8, line_height, width float64) *Section {
	section := new(Section)
	if len(text) == 0 {
		return section
	}

	section.Text        = text
	section.justify     = justify
	section.line_height = line_height

	section.lines = break_section(data, section, width, 0, false)

	if x := len(section.lines); x > 0 {
		section.total_height = float64(x) * line_height
	}

	return section
}