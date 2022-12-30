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
import "github.com/signintech/gopdf"

import "meander/font"

const reserved_name = font.ReservedName

// attach the embedded fonts to the gopdf document
func register_fonts(document *gopdf.GoPdf) {
	document.AddTTFFontByReader(reserved_name, bytes.NewReader(font.Regular))
	document.AddTTFFontByReaderWithOption(reserved_name, bytes.NewReader(font.Bold), gopdf.TtfOption{Style: gopdf.Bold})
	document.AddTTFFontByReaderWithOption(reserved_name, bytes.NewReader(font.Italic), gopdf.TtfOption{Style: gopdf.Italic})
	document.AddTTFFontByReaderWithOption(reserved_name, bytes.NewReader(font.BoldItalic), gopdf.TtfOption{Style: gopdf.Italic | gopdf.Bold})
}