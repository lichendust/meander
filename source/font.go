package main

import "bytes"
import "github.com/signintech/gopdf"

import "meander/font"

const reserved_name = "Courier Prime"

// attach the embedded to the gopdf document
func register_fonts(document *gopdf.GoPdf) {
	document.AddTTFFontByReader(reserved_name, bytes.NewReader(font.Regular))
	document.AddTTFFontByReaderWithOption(reserved_name, bytes.NewReader(font.Bold),       gopdf.TtfOption { Style: gopdf.Bold })
	document.AddTTFFontByReaderWithOption(reserved_name, bytes.NewReader(font.Italic),     gopdf.TtfOption { Style: gopdf.Italic })
	document.AddTTFFontByReaderWithOption(reserved_name, bytes.NewReader(font.BoldItalic), gopdf.TtfOption { Style: gopdf.Italic | gopdf.Bold })
}