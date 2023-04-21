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
	"time"

	lib "github.com/signintech/gopdf"
)

func command_render_document(config *config) {
	text, ok := merge(config.source_file)
	if !ok {
		return
	}

	data := init_data()
	syntax_parser(config, data, text)

	return

	doc := &lib.GoPdf{}

	doc.Start(lib.Config{PageSize: *data.paper})
	doc.SetInfo(lib.PdfInfo{
		Title:        clean_string(data.Title.Title),
		Author:       clean_string(data.Title.Author),
		Creator:      MEANDER,
		CreationDate: time.Now(),
	})

	render_content(config, data, doc)

	if err := doc.WritePdf(fix_path(config.output_file)); err != nil {
		eprintln("error saving", config.output_file)
	}
}

func render_content(config *config, data *Fountain, doc *lib.GoPdf) {
	doc.AddPage()
}