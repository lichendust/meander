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
		panic(err)
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

	document := gopdf.GoPdf{}

	paper := paper_store[config.paper_size]
	document.Start(gopdf.Config{
		PageSize: *paper.paper_data,
	})

	register_fonts(&document)
	set_font(&document, "")

	// we print in black, so make sure it's set
	document.SetFillColor(0, 0, 0)

	// detect the base width of the font for later
	char_width = text_width(&document, " ")

	// embed some info — doesn't seem to work?
	// need to check the library issues.
	{
		info := gopdf.PdfInfo{}

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
	if len(content.title) > 0 {
		render_title_page(&document, config, content)
	}
	if config.include_gender {
		render_gender_data(&document, config)
	}
	if len(content.nodes) > 0 {
		render_content(&document, config, content)
	}

	// write the file to disk and go home
	if err := document.WritePdf(fix_path(config.output_file)); err != nil {
		fmt.Fprintf(os.Stderr, "error saving %s", config.output_file)
	}
}

func render_title_page(document *gopdf.GoPdf, config *config, content *fountain_content) {
	has_any := false

	for _, c := range []string{"title", "credit", "author", "source", "notes", "copyright", "contact", "revision", "draft date"} {
		if _, ok := content.title[c]; ok {
			has_any = true
			break
		}
	}

	if !has_any {
		return
	}

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
		main_align_pos = inch
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

type renderer struct {
	pos_x float64
	pos_y float64

	safe_height float64

	page_center_line float64
	page_max_width   int

	scene_counter  int
	page_counter   int
	first_on_page  bool

	current_header string
	current_footer string

	last_char *syntax_line

	template *template
	paper    *paper
}

func format_header(manager *renderer, content *fountain_content, input string) string {
	input = strings.ReplaceAll(input, "%p", fmt.Sprintf("%d", manager.page_counter))
	input = strings.ReplaceAll(input, "%title", content.title["title"])
	return input
}

func draw_header(document *gopdf.GoPdf, manager *renderer, text string, height float64) {
	if text == "" {
		return
	}

	list := strings.SplitN(text, "|", 3)

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
		draw_text(document, syntax_leaf_parser(right, 500, 0)[0], manager.paper.width - margin_right, height, RIGHT, &line_data{})
	}
	if center != "" {
		draw_text(document, syntax_leaf_parser(center, 500, 0)[0], manager.page_center_line, height, CENTER, &line_data{})
	}
	if left != "" {
		draw_text(document, syntax_leaf_parser(left, 500, 0)[0], margin_left, height, LEFT, &line_data{})
	}
}

func new_page(document *gopdf.GoPdf, manager *renderer, header, footer string) {
	document.AddPage()

	manager.pos_y = margin_top
	manager.first_on_page = true
	manager.page_counter++

	draw_header(document, manager, header, header_height)
	draw_header(document, manager, footer, manager.paper.height - footer_height)
}

func print_more(document *gopdf.GoPdf, manager *renderer) {
	p_indent := manager.template.look_up[PARENTHETICAL].margin

	document.SetX(margin_left + p_indent)
	document.SetY(manager.pos_y)
	document.Text(more_tag)
}

func print_cont(document *gopdf.GoPdf, manager *renderer) {
	c_indent := manager.template.look_up[CHARACTER].margin

	draw_text(document, manager.last_char, margin_left + c_indent, manager.pos_y, LEFT, &line_data{})

	// this is a really sketchy method of
	// checking so as not to dupe
	// physically written (CONT'D)s when
	// page-splitting
	for _, leaf := range manager.last_char.leaves {
		if strings.Contains(leaf.text, cont_check) {
			manager.pos_y += manager.template.line_height
			return
		}
	}

	document.Text(" " + cont_tag)
	manager.pos_y += manager.template.line_height
}

func render_content(document *gopdf.GoPdf, config *config, content *fountain_content) {
	nodes := content.nodes

	template := template_store[config.template]
	paper    := paper_store[config.paper_size]

	if config.include_synopses {
		template.look_up[SYNOPSIS].skip = false
	}
	if config.include_sections {
		template.look_up[SECTION].skip  = false
		template.look_up[SECTION2].skip = false
		template.look_up[SECTION3].skip = false
	}

	manager := &renderer{
		safe_height:      paper.height - margin_bottom,

		page_center_line: paper.width / 2 + margin_left / 2 - margin_right / 2,
		page_max_width:   int((paper.width - margin_right - margin_left) / char_width),

		page_counter:     1,
		current_header:   "%p.",
		first_on_page:    true,

		template:         template,
		paper:            paper,
	}

	{
		inside_dual_dialogue := false

		// pre-assign template data to all content nodes
		for _, the_node := range nodes {
			the_width := manager.page_max_width
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
	manager.first_on_page = true
	manager.pos_y = margin_top
	manager.pos_x = margin_left

	for {
		if len(nodes) == 0 {
			break
		}

		node := nodes[0]

		manager.pos_x = margin_left

		switch node.node_type {
		case WHITESPACE:
			if !manager.first_on_page {
				manager.pos_y += template.line_height * float64(node.level)
			}
			nodes = nodes[1:]
			continue

		case HEADER:
			t := node.raw_text

			if t == "%none" {
				t = ""
			}

			manager.current_header = t
			t = format_header(manager, content, t)
			draw_header(document, manager, t, header_height)

			nodes = consume_node_whitespace(nodes[1:])
			continue

		case FOOTER:
			t := node.raw_text

			if t == "%none" {
				t = ""
			}

			manager.current_footer = t
			t = format_header(manager, content, t)
			draw_header(document, manager, t, paper.height - footer_height)

			nodes = consume_node_whitespace(nodes[1:])
			continue

		case PAGE_NUMBER:
			i, err := strconv.Atoi(node.raw_text)

			if err != nil {
				i = 1 // when in rome
			}

			manager.page_counter = i

			if i > 0 {
				manager.page_counter -= 1
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
			starting_pos_y := manager.pos_y

			// which is tallest
			highest := left_height + pica // this is a magic pica
			if right_height > highest {
				highest = right_height
			}

			// make a new page if the tallest is longer than the
			// remaining page. we just don't split dual
			// dialogue because we can't walk back through the
			// PDF buffer
			if manager.pos_y + highest > manager.safe_height {
				header := format_header(manager, content, manager.current_header)
				footer := format_header(manager, content, manager.current_footer)

				new_page(document, manager, header, footer)
			}

			{
				draw_character_block(document, left, margin_left, manager.pos_y, template.line_height)
				manager.pos_y = starting_pos_y

				draw_character_block(document, right, margin_left + inch * 3, manager.pos_y, template.line_height)
				manager.pos_y = starting_pos_y + highest
			}
			continue
		}

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
			case LEFT:   manager.pos_x = margin_left + t.margin
			case RIGHT:  manager.pos_x = paper.width - margin_right - t.margin + char_width
			case CENTER: manager.pos_x = manager.page_center_line
			}

			// some entries can have an extra pad near them
			if !manager.first_on_page && t.space_above != 0 {
				manager.pos_y += t.space_above
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
				if manager.pos_y + line_height > manager.safe_height {
					do_new_page = true
				}

			case SCENE, SECTION, CHARACTER:
				if manager.pos_y + line_height * 3 > manager.safe_height {
					do_new_page = true
				}
			}

			if do_new_page {
				header := format_header(manager, content, manager.current_header)
				footer := format_header(manager, content, manager.current_footer)

				new_page(document, manager, header, footer)
			}
		}

		// handle scene numbers
		if node.node_type == SCENE {
			document.SetY(manager.pos_y)

			scene_number := ""

			// always remove SCENE_NUMBER node if it exists
			if len(nodes) >= 2 && nodes[1:][0].node_type == SCENE_NUMBER {
				nodes = nodes[1:]
				scene_number = nodes[0].raw_text
			}

			// if generating, iterate the scene_counter and format it
			if config.scenes == SCENE_GENERATE {
				manager.scene_counter++
				scene_number = fmt.Sprintf("%d", manager.scene_counter)
			}

			// if not set to remove, print the scene numbers
			if config.scenes != SCENE_REMOVE {
				font_setter := ""

				plate := template.look_up[SCENE]

				if plate.style & BOLD != 0 {
					font_setter += "B"
				}
				if plate.style & ITALIC != 0 {
					font_setter += "I"
				}

				set_font(document, font_setter)

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
				if manager.pos_y + line_height * 2 > manager.safe_height {
					header := format_header(manager, content, manager.current_header)
					footer := format_header(manager, content, manager.current_footer)

					new_page(document, manager, header, footer)
				}
			}

			switch node.node_type {
				case SYNOPSIS:
					data.base_color   = print_color {100, 100, 100} // @color hardcoded
					data.active_color = print_color {100, 100, 100}

				case CHARACTER:
					manager.last_char = lines[0]
			}

			for i, line := range lines {
				if style != NORMAL {
					syntax_line_override(line, style) // apply any template base styles
				}

				draw_text(document, line, manager.pos_x, manager.pos_y, justify, &data)

				if node.revised {
					document.SetX(paper.width - inch / 2)
					document.Text("*")
				}

				if i != len(lines) - 1 {
					manager.pos_y += line_height

					// internal wrapped block overflow
					if manager.pos_y > manager.safe_height {
						if inside_char { print_more(document, manager) }

						header := format_header(manager, content, manager.current_header)
						footer := format_header(manager, content, manager.current_footer)

						new_page(document, manager, header, footer)
						if inside_char { print_cont(document, manager) }
					}
				}
			}

			// lookahead in case we're about to exit
			if len(nodes[1:]) > 0 && !is_character_train(nodes[1].node_type) {
				inside_char = false
			}
		}

		// regular page overflow at the end of writing
		if manager.pos_y + pica > manager.safe_height || node.node_type == PAGE_BREAK {
			header := format_header(manager, content, manager.current_header)
			footer := format_header(manager, content, manager.current_footer)

			new_page(document, manager, header, footer)

			if node.node_type == PAGE_BREAK {
				nodes = consume_node_whitespace(nodes[1:])
				continue
			}

			nodes = nodes[1:]
			continue
		}

		// edge case for char + dialogue with
		// many elements that were getting split
		if manager.pos_y + line_height > manager.safe_height {
			if inside_char {
				print_more(document, manager)

				header := format_header(manager, content, manager.current_header)
				footer := format_header(manager, content, manager.current_footer)

				new_page(document, manager, header, footer)

				print_cont(document, manager)
			}
		}

		if line_height != template.line_height {
			manager.pos_y += line_height
		} else {
			manager.pos_y += template.line_height
		}

		nodes = nodes[1:]
		manager.first_on_page = false
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
			data.is_bold   = leaf.is_opening
		}
		if leaf.leaf_type == ITALIC {
			data.is_italic = leaf.is_opening
		}
		if leaf.leaf_type == BOLDITALIC {
			data.is_bold   = leaf.is_opening
			data.is_italic = leaf.is_opening
		}
		if leaf.leaf_type == NOTE {
			if leaf.is_opening {
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

func render_gender_data(document *gopdf.GoPdf, config *config) {
	_, data, has_updated, ok := do_full_analysis(config)

	if !ok {
		return // no gender data
	}

	// update the table in the source file
	if has_updated && config.write_gender {
		if ok := gender_replace_comment(config, data); !ok {
			fmt.Fprintf(os.Stderr, "failed to replace gender table!")
			return
		}
	}

	document.AddPage()

	start_x := margin_left
	start_y := margin_top

	{
		title_line := syntax_leaf_parser("GENDER ANALYSIS", 500, 0)[0]
		syntax_line_override(title_line, UNDERLINE | BOLD)
		draw_text(document, title_line, start_x, start_y, LEFT, &line_data{})

		start_y += pica * 2
	}

	start_y = print_data_block(document, config, crunch_chars_by_gender(data), "Character Count by Gender", start_x, start_y, true)
	start_y += pica * 2
	start_y = print_data_block(document, config, crunch_lines_by_gender(data), "Lines by Gender", start_x, start_y, true)
	start_y += pica * 2
	start_y = print_data_block(document, config, crunch_chars_by_lines(data), "Characters by Lines Spoken", start_x, start_y, false)
}

func print_data_block(document *gopdf.GoPdf, config *config, data *data_container, title string, pos_x, pos_y float64, do_bar_graph bool) float64 {
	template := template_store[config.template]
	line_height := template.line_height_title

	document.SetX(pos_x)
	document.SetY(pos_y)

	set_font(document, "B")

	document.Text(title)

	set_font(document, "")

	pos_y += pica * 2

	for _, entry := range data.ordered_data {
		if entry.value == 0 {
			continue
		}

		buffer := strings.Builder{}
		buffer.Grow(128)

		buffer.WriteString(space_pad_string(title_case(entry.name_one), data.longest_name_one))

		if data.longest_name_two > 0 {
			buffer.WriteString(space_pad_string(title_case(entry.name_two), data.longest_name_two))
		}

		buffer.WriteString(space_pad_string(fmt.Sprintf("%d", entry.value), 5))

		{
			percentage := float64(entry.value) / float64(data.total_value) * 100
			buffer.WriteString(space_pad_string(fmt.Sprintf("%.1f%%", percentage), 8))
		}

		if do_bar_graph {
			bar_graph := normalise(entry.value, data.largest_value, 20)
			buffer.WriteString(strings.Repeat("|", bar_graph))
		}

		document.SetX(pos_x)
		document.SetY(pos_y)
		document.Text(buffer.String())

		pos_y += line_height
	}

	return pos_y
}