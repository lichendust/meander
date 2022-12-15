package main

import (
	"bytes"

	"github.com/signintech/gopdf"

	_ "embed"
)

const reserved_name = "Courier Prime"

var (
	//go:embed font/courier_prime_regular.ttf
	font_regular []byte

	//go:embed font/courier_prime_bold.ttf
	font_bold []byte

	//go:embed font/courier_prime_italic.ttf
	font_italic []byte

	//go:embed font/courier_prime_bolditalic.ttf
	font_bold_italic []byte
)

// attach the embedded to the gopdf document
func register_fonts(document *gopdf.GoPdf) {
	document.AddTTFFontByReader(reserved_name, bytes.NewReader(font_regular))
	document.AddTTFFontByReaderWithOption(reserved_name, bytes.NewReader(font_bold),        gopdf.TtfOption { Style: gopdf.Bold })
	document.AddTTFFontByReaderWithOption(reserved_name, bytes.NewReader(font_italic),      gopdf.TtfOption { Style: gopdf.Italic })
	document.AddTTFFontByReaderWithOption(reserved_name, bytes.NewReader(font_bold_italic), gopdf.TtfOption { Style: gopdf.Italic | gopdf.Bold })
}