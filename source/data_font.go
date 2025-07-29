/*
	Meander
	A portable Fountain utility for production writing
	Copyright (C) 2022-2023 Harley Denham
*/

package main

import "bytes"
import "github.com/lichendust/meander/font"

import lib "github.com/signintech/gopdf"

const RESERVED_NAME = font.RESERVED_NAME
const LICENSE_TEXT  = font.LICENSE_TEXT
const CHAR_WIDTH    = font.CHAR_WIDTH
const FONT_SIZE     = font.FONT_SIZE

var export_fonts = font.ExportFonts

func register_fonts(doc *lib.GoPdf) {
	doc.AddTTFFontByReader(RESERVED_NAME, bytes.NewReader(font.Regular))
	doc.AddTTFFontByReaderWithOption(RESERVED_NAME, bytes.NewReader(font.Bold),       lib.TtfOption{Style: lib.Bold})
	doc.AddTTFFontByReaderWithOption(RESERVED_NAME, bytes.NewReader(font.Italic),     lib.TtfOption{Style: lib.Italic})
	doc.AddTTFFontByReaderWithOption(RESERVED_NAME, bytes.NewReader(font.BoldItalic), lib.TtfOption{Style: lib.Italic | lib.Bold})
}
