/*
	Meander
	A portable Fountain utility for production writing
	Copyright (C) 2022-2023 Harley Denham
*/

package main

import "fmt"
import "math"
import "strings"

import lib "github.com/signintech/gopdf"

func set_color(doc *lib.GoPdf, color Color) {
	doc.SetFillColor(color.R, color.G, color.B)
	doc.SetTextColor(color.R, color.G, color.B)
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
	text, success := merge(config.source_file)
	if !success {
		return
	}

	data := init_data(config)

	syntax_parser(config, data, text)
	vet_template(data.template)
	paginate(config, data)

	if config.starred_show && len(data.config.starred_target) == 0 && len(data.Title.Revision) > 0 {
		data.config.starred_target = strings.ToLower(data.Title.Revision)
	}

	if config.starred_only {
		page_count  := data.Content[len(data.Content) - 1].page
		valid_pages := make([]bool, page_count + 1)

		has_any := false

		for i := range data.Content {
			section := &data.Content[i]
			if len(section.Revision) > 0 && strings.Contains(data.config.starred_target, section.Revision) {
				valid_pages[section.page] = true
				has_any = true
			}
		}

		if !has_any {
			eprintln("--revision-only: there are no revision tags in this document that match the input — PDF would be blank")
			return
		}

		for i := range data.Content {
			section := &data.Content[i]
			section.skip = !valid_pages[section.page]
		}
	}

	doc := new(lib.GoPdf)

	doc.Start(lib.Config{
		PageSize: config.paper_size,
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
	if !data.Title.has_any || data.config.starred_only {
		return
	}

	const LINE_HEIGHT = LINE_HEIGHT * 1.5

	const WIDTH  = INCH * 3 // bottom corners
	title_width := INCH * 4 // main title

	start_x := INCH
	start_y := INCH * 3.5

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
			draw_section(doc, data, title)
			start_y += title.total_height + LINE_HEIGHT * 4
		}

		if data.Title.Credit != "" {
			credit.pos_x = start_x
			credit.pos_y = start_y
			draw_section(doc, data, credit)
			start_y += credit.total_height + LINE_HEIGHT * 2
		}

		if data.Title.Author != "" {
			author.pos_x = start_x
			author.pos_y = start_y
			draw_section(doc, data, author)
			start_y += author.total_height + LINE_HEIGHT * 4
		}

		if data.Title.Source != "" {
			source.pos_x = start_x
			source.pos_y = start_y
			draw_section(doc, data, source)
		}
	}

	start_x = INCH

	// Notes    Copyright    Contact
	{
		note := quick_section(data, data.Title.Notes,     LEFT, LINE_HEIGHT, WIDTH)
		cont := quick_section(data, data.Title.Contact,   LEFT, LINE_HEIGHT, WIDTH)
		copy := quick_section(data, data.Title.Copyright, LEFT, LINE_HEIGHT, WIDTH)

		start_y = data.template.paper.H - INCH - (note.total_height + copy.total_height + cont.total_height) + LINE_HEIGHT

		if data.Title.Notes != "" {
			note.pos_x = start_x
			note.pos_y = start_y
			draw_section(doc, data, note)
			start_y += note.total_height
		}

		if data.Title.Contact != "" {
			cont.pos_x = start_x
			cont.pos_y = start_y
			draw_section(doc, data, cont)
			start_y += cont.total_height
		}

		if data.Title.Copyright != "" {
			copy.pos_x = start_x
			copy.pos_y = start_y
			draw_section(doc, data, copy)
		}
	}

	start_x = data.template.paper.W - INCH

	// Revision    DraftDate    Info
	{
		revs := quick_section(data, data.Title.Revision,  RIGHT, LINE_HEIGHT, WIDTH)
		drft := quick_section(data, data.Title.DraftDate, RIGHT, LINE_HEIGHT, WIDTH)
		info := quick_section(data, data.Title.Info,      RIGHT, LINE_HEIGHT, WIDTH)

		start_y = data.template.paper.H - INCH - (revs.total_height + drft.total_height + info.total_height) + LINE_HEIGHT

		if data.Title.Revision != "" {
			revs.pos_x = start_x
			revs.pos_y = start_y
			draw_section(doc, data, revs)
			start_y += revs.total_height + LINE_HEIGHT
		}

		if data.Title.DraftDate != "" {
			drft.pos_x = start_x
			drft.pos_y = start_y
			draw_section(doc, data, drft)
			start_y += drft.total_height + LINE_HEIGHT
		}

		if data.Title.Info != "" {
			info.pos_x = start_x
			info.pos_y = start_y
			draw_section(doc, data, info)
		}
	}
}

func render_gender(config *Config, data *Fountain, doc *lib.GoPdf) {
	if !config.include_gender || data.config.starred_only {
		return
	}

	doc.AddPage()

	start_y := data.template.margin_top

	doc.SetXY(data.template.margin_left, start_y)

	{
		gender_title := fmt.Sprintf("%q %s", clean_string(data.Title.Title), GENDER_HEADING)
		length := rune_count(gender_title)
		t := Line{
			length: length,
			leaves: []Leaf{{NORMAL, false, gender_title}},
		}
		line_override(&t, UNDERLINE)
		draw_line(doc, data.template, &t, data.template.margin_left, start_y)
	}

	start_y += LINE_HEIGHT * 2

	render_gender_data(data, doc, crunch_chars_by_gender(data), GENDER_CHARS_BY_GENDER, &start_y)
	start_y += LINE_HEIGHT
	render_gender_data(data, doc, crunch_lines_by_gender(data), GENDER_LINES_BY_GENDER, &start_y)
	start_y += LINE_HEIGHT
	render_gender_data(data, doc, crunch_chars_by_lines(data),  GENDER_CHARS_BY_LINES,  &start_y)
}

func render_toc(config *Config, data *Fountain, doc *lib.GoPdf) {
	if !config.table_of_contents || data.config.starred_only {
		return
	}

	has_any := false
	widest_scene_no := 0
	for i := range data.Content {
		section := data.Content[i]

		if section.Type > is_section || section.Type == SCENE {
			has_any = true
		}

		if section.Type == SCENE {
			w := rune_count(section.SceneNumber)
			if w > widest_scene_no {
				widest_scene_no = w
			}
		}
	}
	if widest_scene_no > 0 { widest_scene_no += 3 }
	scene_inset := CHAR_WIDTH * float64(widest_scene_no)

	if !has_any {
		return
	}

	doc.AddPage()

	running_x := data.template.margin_left
	running_y := data.template.margin_top

	for i := range data.Content {
		section := data.Content[i]

		if section.Type > is_section || section.Type == SCENE {
			new_section := copy_without_style(section)

			if section.Type > is_section && data.template.kind == SCREENPLAY {
				// @todo why is there extra spacing being added post-pagination???
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

			draw_section(doc, data, &new_section)

			page_number := fmt.Sprintf("%d", new_section.page)
			doc.SetX(data.template.margin_right - float64(rune_count(page_number)) * CHAR_WIDTH)
			doc.Text(page_number)

			set_font(doc, NO_TYPE)

			if section.Type > is_section {
				running_y += LINE_HEIGHT
			}

			running_y += LINE_HEIGHT
		}

		if running_y > data.template.paper.H - data.template.margin_bottom {
			doc.AddPage()
			running_y = data.template.margin_top
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

			if config.template == STORYBOARD {
				draw_board(data, doc)
			}
		}

		if section.Type == SCENE && config.scenes != SCENE_REMOVE {
			text_width := float64(rune_count(section.SceneNumber)) * CHAR_WIDTH
			right_x    := data.template.margin_right - text_width
			left_x     := data.template.margin_left - INCH / 2 - text_width

			set_font(doc, NO_TYPE)

			doc.SetY(section.pos_y)
			doc.SetX(left_x)
			doc.Text(section.SceneNumber)

			draw_section(doc, data, section)

			doc.SetX(right_x)
			doc.Text(section.SceneNumber)
			continue
		}

		draw_section(doc, data, section)
	}
}

func draw_section(doc *lib.GoPdf, data *Fountain, section *Section) {
	if section == nil {
		return
	}
	if section.Text == "" {
		return
	}

	set_color(doc, data.template.text_color)

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
		draw_star(doc, data, section, section.pos_y)
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

		draw_line(doc, data.template, &line, pos_x, pos_y)
		draw_star(doc, data, section, pos_y)
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
		draw_line(doc, data.template, &t, data.template.margin_left, *start_y)
	}

	set_font(doc, NO_TYPE)

	*start_y += LINE_HEIGHT * 2

	for _, entry := range data_set.data {
		if entry.value == 0 {
			continue
		}

		running_x := data.template.margin_left

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

func draw_star(doc *lib.GoPdf, data *Fountain, section *Section, pos_y float64) {
	if !data.config.starred_show {
		return
	}

	if len(section.Revision) > 0 && strings.Contains(data.config.starred_target, section.Revision) {
		// @todo might do left/right dual stars at some point
		/*if is_character_train(section.Type) && section.Level == 1 {
			doc.SetX(data.template.starred_left)
		} else {
			doc.SetX(data.template.starred_right)
		}
		doc.SetY(pos_y + data.template.starred_nudge)*/

		doc.SetXY(data.template.starred_margin, pos_y + data.template.starred_nudge)
		doc.Text("*")
	}
}

func draw_board(data *Fountain, doc *lib.GoPdf) {
	x := MARGIN_LEFT + data.template.types[ACTION].width + INCH / 2
	y := MARGIN_TOP - PICA * 0.7

	h := (data.template.paper.H - MARGIN_TOP - MARGIN_BOTTOM - PICA * 1.5) / 3
	w := h * 2.35

	for i := float64(0); i < 3; i += 1 {
		y := y + (h + PICA) * i
		doc.Rectangle(x, y, x + w, y + h, "", 0, 0)
	}
}
