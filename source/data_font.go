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

import "bytes"
import "github.com/qxoko/meander/font"

import lib "github.com/signintech/gopdf"

const RESERVED_NAME = font.RESERVED_NAME
const LICENSE_TEXT  = font.LICENSE_TEXT
const CHAR_WIDTH    = font.CHAR_WIDTH

var export_fonts = font.ExportFonts

func register_fonts(doc *lib.GoPdf) {
	doc.AddTTFFontByReader(RESERVED_NAME, bytes.NewReader(font.Regular))
	doc.AddTTFFontByReaderWithOption(RESERVED_NAME, bytes.NewReader(font.Bold),       lib.TtfOption{Style: lib.Bold})
	doc.AddTTFFontByReaderWithOption(RESERVED_NAME, bytes.NewReader(font.Italic),     lib.TtfOption{Style: lib.Italic})
	doc.AddTTFFontByReaderWithOption(RESERVED_NAME, bytes.NewReader(font.BoldItalic), lib.TtfOption{Style: lib.Italic | lib.Bold})
}