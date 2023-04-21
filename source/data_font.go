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
	"bytes"
	"github.com/signintech/gopdf"
	"github.com/qxoko/meander/font"
)

const RESERVED_NAME = font.RESERVED_NAME
const CHAR_WIDTH    = font.CHAR_WIDTH

func register_fonts(document *gopdf.GoPdf) {
	document.AddTTFFontByReader(RESERVED_NAME, bytes.NewReader(font.Regular))
	document.AddTTFFontByReaderWithOption(RESERVED_NAME, bytes.NewReader(font.Bold), gopdf.TtfOption{Style: gopdf.Bold})
	document.AddTTFFontByReaderWithOption(RESERVED_NAME, bytes.NewReader(font.Italic), gopdf.TtfOption{Style: gopdf.Italic})
	document.AddTTFFontByReaderWithOption(RESERVED_NAME, bytes.NewReader(font.BoldItalic), gopdf.TtfOption{Style: gopdf.Italic | gopdf.Bold})
}