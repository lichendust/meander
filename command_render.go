package main

import (
	"os"
	"fmt"
	"time"
	"strings"
	"strconv"

	"github.com/signintech/gopdf"
)

// caches the width of a single character
// from the selected font — this is why
// meander is monospace-only right now
var char_width float64 = 0

// line_data is used to store initial data for draw_text
// but the same copy is fed back into successive lines to
// colours and styles are preserved during line-breaks.
type line_data struct {
	is_bold      bool
	is_italic    bool
	active_color print_color
	base_color   print_color
}

// gopdf takes colours literally as RGB u8
// values, so this we just carry them in here
type print_color struct {
	R uint8
	G uint8
	B uint8
}

// shorthand for applying our little colour struct
func set_color(document *gopdf.GoPdf, color print_color) {
	document.SetFillColor(color.R, color.G, color.B)
}

// applies styles using the library's dumb API
func set_font(document *gopdf.GoPdf, style string) {
	err := document.SetFont(reserved_name, style, font_size)

	if err != nil {
		panic(err)
	}
}

// measures the pixel width of some text by the glyphs
// of the current font
func text_width(document *gopdf.GoPdf, t string) float64 {
	w, err := document.MeasureTextWidth(t)

	if err != nil {
		panic(err) // this is a timebomb
	}

	return w
}

// removes whitespace to balance skipped
// nodes, by either skipping the entire
// whitespace node or knocking it down by
// one level
func consume_node_whitespace(nodes []*syntax_node) []*syntax_node {
	if len(nodes) > 1 {
		if n := nodes[0]; n.node_type == WHITESPACE {
			if n.level > 1 {
				n.level--
			} else {
				nodes = nodes[1:]
			}
		}
	}
	return nodes
}

// the root function for PDF rendering
func command_render_document(config *config) {
	content, ok := syntax_parser(config)

	if !ok {
		return
	}

	// check title page for overridden values
	// the user might override the format or
	// paper size with arguments, so we deal
	// with that once we've got the title page
	{
		if config.template == "" {
			if x, ok := content.title["format"]; ok {
				config.template = strings.ToLower(x)
			} else {
				config.template = default_template
			}

			if _, ok := template_store[config.template]; !ok {
				fmt.Fprintf(os.Stderr, "title page: %q not a template\n", config.template)
				return
			}
		}

		if config.paper_size == "" {
			if x, ok := content.title["paper"]; ok {
				config.paper_size = strings.ToLower(x)
			} else {
				config.paper_size = default_paper
			}

			if _, ok := paper_store[config.paper_size]; !ok {
				fmt.Fprintf(os.Stderr, "title page: %q not a supported paper size\n", config.paper_size)
				return
			}
		}
	}

	// init the document
	document := gopdf.GoPdf {}

	// load the paper data to apply it...
	paper := paper_store[config.paper_size]

	// ...to the document
	document.Start(gopdf.Config {
		PageSize: *paper.paper_data,
	})

	// cache all our fonts ready for printing
	{
		ok := false

		if config.font_name == "" {
			ok = register_default_fonts(&document)
		} else {
			ok = register_custom_fonts(&document, config.font_name)
		}

		if !ok {
			return // we should @error here
		}
	}

	// set the base font
	set_font(&document, "")

	// we print in black, so make sure it's set
	document.SetFillColor(0, 0, 0)

	// detect the base width of the font for later
	char_width = text_width(&document, " ")

	// embed some info — doesn't seem to work?
	// need to check the library issues.
	{
		info := gopdf.PdfInfo {}

		if x, ok := content.title["title"]; ok {
			info.Title = clean_string(x)
		}
		if x, ok := content.title["author"]; ok {
			info.Author = clean_string(x)
		}

		info.CreationDate = time.Now()

		document.SetInfo(info)
	}

	// render all the parser data into the screenplay
	render_title_page(&document, config, content)

	// @todo
	/*if config.gender_embed {
		render_gender_analysis(&document, config)
	}*/

	render_content(&document, config, content)

	// write the file to disk and go home
	if err := document.WritePdf(fix_path(config.output_file)); err != nil {
		fmt.Fprintf(os.Stderr, "error saving %s", config.output_file)
	}
}

func render_title_page(document *gopdf.GoPdf, config *config, content *fountain_content) {
	// exit if no work is necessary
	if len(content.title) == 0 {
		return
	}

	// add the page we're writing to
	document.AddPage()

	// get a bunch of the data we need in advance
	title       := content.title

	template    := template_store[config.template]
	paper       := paper_store[config.paper_size]

	main_title_align := template.title_page_align

	line_height := template.line_height_title

	// we curate all the title page
	// positions, because clearly we know
	// best.
	pos_y := margin_top + pica * 15

	main_align_pos := float64(0)

	switch main_title_align {
	case CENTER:
		main_align_pos = paper.width / 2
	case LEFT:
		main_align_pos = margin_left
	case RIGHT:
		main_align_pos = margin_right
	}

	/*
		loop through the central group and print them

		because the title page is the only place newlines can be found
		we render them in this hacky split-then-parse-then-render
		way that's really weird and bad.  we should fix that.
	*/
	for i, c := range []string{"title", "credit", "author", "source"} {
		if t, ok := title[c]; ok {
			lines := strings.Split(t, "\n")

			for _, line := range lines {
				if line == "" {
					pos_y += line_height
					continue
				}
				draw_text(document, syntax_leaf_parser(line, 500, 0)[0], main_align_pos, pos_y, main_title_align, &line_data{})
				pos_y += line_height
			}

			if i == 0 || i == 2 {
				pos_y += line_height * 4
			} else {
				pos_y += line_height * 2
			}
		}
	}

	// jump to bottom left of page
	pos_y = paper.height - margin_bottom

	/*
		we do the same dumb line-split as before
	*/
	for _, c := range []string{"notes", "copyright", "contact"} {
		if t, ok := title[c]; ok {
			lines := strings.Split(t, "\n")

			pos_y -= line_height * float64(len(lines) - 1)

			local_y := pos_y

			for _, t := range lines {
				if t == "" {
					pos_y += line_height
					continue
				}

				draw_text(document, syntax_leaf_parser(t, 500, 0)[0], inch, local_y, LEFT, &line_data{})
				local_y += line_height
			}

			pos_y -= line_height
		}
	}

	// bottom right of page
	pos_y = paper.height - margin_bottom

	for _, c := range []string{"revision", "draft date"} {
		if t, ok := title[c]; ok {
			lines := strings.Split(t, "\n")

			pos_y -= line_height * float64(len(lines) - 1)

			local_y := pos_y

			for _, t := range lines {
				if t == "" {
					pos_y += line_height
					continue
				}

				draw_text(document, syntax_leaf_parser(t, 500, 0)[0], paper.width - inch, local_y, RIGHT, &line_data{})
				local_y += line_height
			}

			pos_y -= line_height
		}
	}
}

func render_content(document *gopdf.GoPdf, config *config, content *fountain_content) {
	if len(content.nodes) == 0 {
		return
	}

	nodes := content.nodes

	template := template_store[config.template]
	paper    := paper_store[config.paper_size]

	if config.include_synopses {
		template.look_up[SYNOPSIS].skip = false
	}

	safe_height := paper.height - margin_bottom

	pos_x := margin_left
	pos_y := margin_top

	page_max_width   := int((paper.width - margin_right - margin_left) / char_width)
	page_center_line := paper.width / 2 + margin_left / 2 - margin_right / 2

	scene_counter  := 0
	page_counter   := 1
	current_header := "%p."
	current_footer := ""
	first_on_page  := true

	var last_char *syntax_line

	pindent := template.look_up[PARENTHETICAL].margin
	cindent := template.look_up[CHARACTER].margin

	/*sections_enabled := !template.look_up[SECTION].skip
	section_text := make(map[string]string, 4)*/

	// we're cheating with some inline functions
	// god is dead and there are no rules anymore
	format_header := func(input string) string {
		input = strings.ReplaceAll(input, "%p", fmt.Sprintf("%d", page_counter))
		input = strings.ReplaceAll(input, "%title", content.title["title"])
		return input
	}

	draw_header := func(document *gopdf.GoPdf, text string, height float64) {
		if text == "" {
			return
		}

		list := strings.SplitN(format_header(text), "|", 3)

		for i, text := range list {
			list[i] = strings.TrimSpace(text)
		}

		left   := ""
		center := ""
		right  := ""

		switch len(list) {
		case 1:
			right  = list[0]
		case 2:
			left   = list[0]
			right  = list[1]
		case 3:
			left   = list[0]
			center = list[1]
			right  = list[2]
		}

		if right != "" {
			draw_text(document, syntax_leaf_parser(right, 500, 0)[0], paper.width - margin_right, height, RIGHT, &line_data{})
		}
		if center != "" {
			draw_text(document, syntax_leaf_parser(center, 500, 0)[0], page_center_line, height, CENTER, &line_data{})
		}
		if left != "" {
			draw_text(document, syntax_leaf_parser(left, 500, 0)[0], margin_left, height, LEFT, &line_data{})
		}
	}

	new_page := func(document *gopdf.GoPdf, current_header, current_footer string) {
		document.AddPage()
		pos_y = margin_top
		first_on_page = true
		page_counter++

		draw_header(document, current_header, header_height)
		draw_header(document, current_footer, paper.height - footer_height)
	}

	print_more := func() {
		document.SetX(margin_left + pindent)
		document.SetY(pos_y)
		document.Text(more_tag)
	}

	print_cont := func() {
		draw_text(document, last_char, margin_left + cindent, pos_y, LEFT, &line_data{})

		// this is a really sketchy method of
		// checking so as not to dupe
		// physically written (CONT'D)s when
		// page-splitting
		for _, leaf := range last_char.leaves {
			if strings.Contains(leaf.text, cont_check) {
				pos_y += template.line_height
				return
			}
		}

		document.Text(" " + cont_tag)
		pos_y += template.line_height
	}

	{
		inside_dual_dialogue := false

		// pre-assign template data to all content nodes
		for _, the_node := range nodes {
			the_width := page_max_width
			the_style := NORMAL

			t, has_template := template.look_up[the_node.node_type]

			if has_template {
				the_node.template = t

				the_style = t.style

				switch t.casing {
				case UPPERCASE: the_node.raw_text = strings.ToUpper(the_node.raw_text)
				case LOWERCASE: the_node.raw_text = strings.ToLower(the_node.raw_text)
				}

				if t.para_indent > 0 {
					the_node.raw_text = strings.Repeat(" ", t.para_indent) + the_node.raw_text
				}

				if t.width > 0 {
					the_width = int(t.width / char_width)
				}
			}

			if template.allow_dual_dialogue && is_character_train(the_node.node_type) {
				if the_node.node_type == CHARACTER && the_node.level > 0 {
					inside_dual_dialogue = true
				}
			} else {
				inside_dual_dialogue = false
			}

			if inside_dual_dialogue {
				switch the_node.node_type {
				case PARENTHETICAL:   the_width = int((dual_width - dual_para_margin) / char_width)
				case DIALOGUE, LYRIC: the_width = int(dual_width / char_width)
				}
			}

			// apply the line-wrap and the_style overrides
			the_node.lines = syntax_leaf_parser(the_node.raw_text, the_width, 0)

			if the_style != NORMAL {
				for _, line := range the_node.lines {
					syntax_line_override(line, the_style)
				}
			}
		}
	}

	// let's begin!
	document.AddPage()

	for {
		if len(nodes) == 0 {
			break
		}

		node := nodes[0]

		pos_x = margin_left

		switch node.node_type {
		case WHITESPACE:
			if !first_on_page {
				pos_y += template.line_height * float64(node.level)
			}
			nodes = nodes[1:]
			continue

		/*case SECTION:
			if sections_enabled {
				section_text[strings.Repeat("#", int(node.level))] = node.raw_text
			}*/

		case HEADER:
			current_header = node.raw_text

			if current_header == "%none" {
				current_header = ""
			}

			draw_header(document, current_header, header_height)

			nodes = consume_node_whitespace(nodes[1:])
			continue

		case FOOTER:
			current_footer = node.raw_text

			if current_footer == "%none" {
				current_footer = ""
			}

			draw_header(document, current_footer, paper.height - footer_height)

			nodes = consume_node_whitespace(nodes[1:])
			continue

		case PAGENUMBER:
			i, err := strconv.Atoi(node.raw_text)

			if err != nil {
				i = 1 // when in rome
			}

			page_counter = i

			if i > 0 {
				page_counter -= 1
			}

			nodes = consume_node_whitespace(nodes[1:])
			continue
		}

		// dual dialogue
		if template.allow_dual_dialogue && node.node_type == CHARACTER && node.level > 0 {
			// get left-hand side
			left := grab_character_block(nodes)
			nodes = nodes[len(left):]

			left_height := float64(0)

			for _, node := range left {
				line_height := template.line_height

				if node.template.line_height > 0 {
					line_height = node.template.line_height
				}

				left_height += line_height * float64(len(node.lines))
				left_height += node.template.space_above
			}

			// this is guaranteed by the parser
			if nodes[0].node_type == WHITESPACE {
				nodes = nodes[1:]
			}

			// get right-hand side
			right := grab_character_block(nodes)
			nodes = nodes[len(right):]

			right_height := float64(0)

			for _, node := range right {
				line_height := template.line_height

				if node.template.line_height > 0 {
					line_height = node.template.line_height
				}

				right_height += line_height * float64(len(node.lines))
				right_height += node.template.space_above
			}

			// store the starting position of
			// either so we know where we're at
			starting_pos_y := pos_y

			// which is tallest
			highest := left_height
			if right_height > highest {
				highest = right_height
			}

			// make a new page if the tallest is longer than the
			// remaining page. we just don't split dual
			// dialogue because we can't walk back through the
			// PDF buffer
			if pos_y + highest > safe_height {
				new_page(document, current_header, current_footer)
			}

			{
				draw_character_block(document, right, margin_left, pos_y, template.line_height)
				pos_y = starting_pos_y

				draw_character_block(document, right, margin_left + inch * 3, pos_y, template.line_height)
				pos_y = starting_pos_y + highest
			}
			continue
		}

		// we get the type just so we can
		// check if it's a section
		// the_type := node.node_type

		// if it is a section, we offset it
		// by the level to get SECTION2/3 etc.
		/*if the_type == SECTION {
			the_type += node.level
		}*/

		// do all that templating stuff!
		inside_char := is_character_train(node.node_type)
		justify     := LEFT
		style       := NORMAL
		line_height := template.line_height

		if node.template != nil {
			t := node.template

			// in screenplays, sections are skipped
			// so we do that from here and remove
			// the relevant whitespace
			if t.skip {
				nodes = consume_node_whitespace(nodes[1:])
				continue
			}

			style   = t.style
			justify = t.justify

			switch justify {
			case LEFT:   pos_x = margin_left + t.margin
			case RIGHT:  pos_x = paper.width - margin_right - t.margin + char_width
			case CENTER: pos_x = page_center_line
			}

			// some entries can have an extra pad near them
			if !first_on_page && t.space_above != 0 {
				pos_y += t.space_above
			}

			if t.line_height > 0 {
				line_height = t.line_height
			} else {
				line_height = template.line_height
			}
		}

		// page overflow edge cases
		{
			do_new_page := false

			switch node.node_type {
			case PARENTHETICAL:
				if pos_y + line_height > safe_height {
					do_new_page = true
				}

			case SCENE, SECTION, CHARACTER:
				if pos_y + line_height * 3 > safe_height {
					do_new_page = true
				}
			}

			if do_new_page {
				new_page(document, current_header, current_footer)
			}
		}

		// handle scene numbers
		if node.node_type == SCENE {
			document.SetY(pos_y)

			scene_number := ""

			// always remove SCENE_NUMBER node if it exists
			if len(nodes) >= 2 && nodes[1:][0].node_type == SCENE_NUMBER {
				nodes = nodes[1:]
				scene_number = nodes[0].raw_text
			}

			// if generating, iterate the scene_counter and format it
			if config.scenes == SCENE_GENERATE {
				scene_counter++
				scene_number = fmt.Sprintf("%d", scene_counter)
			}

			// if not set to remove, print the scene numbers
			if config.scenes != SCENE_REMOVE {
				document.SetX(margin_left / 2)
				document.Text(scene_number)

				document.SetX(paper.width - margin_right)
				document.Text(scene_number)
			}
		}

		{
			lines := node.lines

			data := line_data {
				is_bold:      style & BOLD != 0,
				is_italic:    style & ITALIC != 0,
				active_color: print_color {0, 0, 0},
				base_color:   print_color {0, 0, 0},
			}

			// immediate page overflow before writing
			// based on wrapped block height, if necessary
			if len(lines) >= 2 {
				if pos_y + line_height > safe_height {
					new_page(document, current_header, current_footer)
				}
			}

			switch node.node_type {
				case SYNOPSIS:
					data.base_color   = print_color {100, 100, 100} // @color hardcoded
					data.active_color = print_color {100, 100, 100}

				case CHARACTER:
					last_char = lines[0]
			}

			for i, line := range lines {
				if style != NORMAL {
					syntax_line_override(line, style) // apply any template base styles
				}

				draw_text(document, line, pos_x, pos_y, justify, &data)

				if node.revised {
					document.SetX(paper.width - inch / 2)
					document.Text("*")
				}

				if i != len(lines) - 1 {
					pos_y += line_height

					// internal wrapped block overflow
					if pos_y > safe_height {
						if inside_char { print_more() }
						new_page(document, current_header, current_footer)
						if inside_char { print_cont() }
					}
				}
			}

			// lookahead in case we're about to exit
			if len(nodes[1:]) > 0 && !is_character_train(nodes[1].node_type) {
				inside_char = false
			}
		}

		// regular page overflow at the end of writing
		if pos_y > safe_height || node.node_type == PAGE_BREAK {
			new_page(document, current_header, current_footer)

			if node.node_type == PAGE_BREAK {
				nodes = consume_node_whitespace(nodes[1:])
				continue
			}

			nodes = nodes[1:]
			continue
		}

		// edge case for char + dialogue with
		// many elements that were getting split
		if pos_y + line_height > safe_height {
			if inside_char {
				print_more()
				new_page(document, current_header, current_footer)
				print_cont()
			}
		}

		if line_height != template.line_height {
			pos_y += line_height
		} else {
			pos_y += template.line_height
		}
		nodes = nodes[1:]
		first_on_page = false
	}
}

func draw_text(document *gopdf.GoPdf, line *syntax_line, start_x, start_y float64, justify uint8, data *line_data) {
	set_color(document, data.active_color)

	pos_x := start_x
	pos_y := start_y

	switch justify {
	case RIGHT:
		pos_x = start_x - float64(line.length) * char_width
	case CENTER:
		pos_x = start_x - float64(line.length) * char_width / 2
	case LEFT:
		pos_x = start_x
	}

	document.SetX(pos_x)
	document.SetY(pos_y)

	// draw highlights for the entire line
	if len(line.highlight) > 0 {
		set_color(document, print_color {255, 249, 115}) // @color hardcoded

		for i := 0; i < len(line.highlight) - 1; i += 2 {
			a := line.highlight[i]
			b := line.highlight[i + 1]

			pos_a := float64(a) * char_width
			pos_b := float64(b - a) * char_width + pos_a

			y := pos_y - pica + 2

			document.Rectangle(pos_x + pos_a - 2, y, pos_x + pos_b + 2, y + float64(font_size) + 2, "F", 0, 0)
		}

		set_color(document, data.active_color)
	}

	// if there's any bold on the line, draw
	// the underlines and strikeouts in bold
	do_bold_line := false

	if data.is_bold {
		// if we're in a bold patch,
		// we're already good to go
		do_bold_line = true
	} else {
		// otherwise just check real quick
		for _, leaf := range line.leaves {
			if leaf.leaf_type == BOLD {
				do_bold_line = true
				break
			}
		}
	}

	// set the line thickness
	if do_bold_line {
		document.SetLineWidth(2) // @todo hardcoded
	} else {
		document.SetLineWidth(1)
	}

	// draw the underlines for the entire line
	if len(line.underline) > 0 {
		for i := 0; i < len(line.underline) - 1; i += 2 {
			a := line.underline[i]
			b := line.underline[i + 1]

			pos_a := float64(a) * char_width
			pos_b := float64(b - a) * char_width + pos_a

			y := pos_y + 2.5

			document.Line(pos_x + pos_a, y, pos_x + pos_b, y)
		}
	}

	// draw strikeouts for the entire line
	if len(line.strikeout) > 0 {
		for i := 0; i < len(line.strikeout) - 1; i += 2 {
			a := line.strikeout[i]
			b := line.strikeout[i + 1]

			pos_a := float64(a) * char_width
			pos_b := float64(b - a) * char_width + pos_a

			y := pos_y - pica / 4

			document.Line(pos_x + pos_a, y, pos_x + pos_b, y)
		}
	}

	// draw all the leaves
	for _, leaf := range line.leaves {
		if leaf.leaf_type == BOLD {
			data.is_bold   = leaf.opening
		}
		if leaf.leaf_type == ITALIC {
			data.is_italic = leaf.opening
		}
		if leaf.leaf_type == BOLDITALIC {
			data.is_bold   = leaf.opening
			data.is_italic = leaf.opening
		}
		if leaf.leaf_type == NOTE {
			if leaf.opening {
				data.active_color = print_color {100, 100, 100} // @color hardcoded
			} else {
				data.active_color = data.base_color
			}
		}

		set_color(document, data.active_color)

		font_setter := line.font_reset

		if data.is_bold   { font_setter += "B" }
		if data.is_italic { font_setter += "I" }

		set_font(document, font_setter)

		if leaf.leaf_type == NORMAL {
			document.Text(leaf.text)
		}
	}
}

// @todo temp functions before rewrite
func grab_character_block(nodes []*syntax_node) ([]*syntax_node) {
	for i, node := range nodes {
		if !is_character_train(node.node_type) {
			return nodes[:i]
		}
	}
	return nil
}

// @todo temp functions before rewrite
func draw_character_block(document *gopdf.GoPdf, nodes []*syntax_node, start_pos_x, pos_y, doc_line_height float64) {
	line_height := doc_line_height

	for _, node := range nodes {
		style := NORMAL
		pos_x := start_pos_x

		if node.template != nil {
			t := node.template

			style = t.style

			switch t.casing {
			case UPPERCASE: node.raw_text = strings.ToUpper(node.raw_text)
			case LOWERCASE: node.raw_text = strings.ToLower(node.raw_text)
			}

			if t.line_height > 0 {
				line_height = t.line_height
			}
		}

		switch node.node_type {
		case CHARACTER:     pos_x += dual_char_margin
		case PARENTHETICAL: pos_x += dual_para_margin
		}

		data := line_data {
			is_bold:      style & BOLD != 0,
			is_italic:    style & ITALIC != 0,
			active_color: print_color {0, 0, 0},
			base_color:   print_color {0, 0, 0},
		}

		for i, line := range node.lines {
			draw_text(document, line, pos_x, pos_y, LEFT, &data)

			if i != len(node.lines) - 1 {
				pos_y += line_height
			}
		}

		if line_height != doc_line_height {
			pos_y += line_height
		} else {
			pos_y += doc_line_height
		}
	}
}