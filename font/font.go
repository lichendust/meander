package font

import "os"
import _ "embed"

// we need to register the reserved font name
// for the PDF metadata to be correct
const RESERVED_NAME = "Courier Prime"
const LICENSE_TEXT  = `Copyright (C) 2013, Quote-Unquote Apps
(quoteunquoteapps.com) with Reserved Font Name Courier Prime.`

const CHAR_WIDTH float64 = 7.188

var (
	//go:embed courier_prime_regular.ttf
	Regular []byte

	//go:embed courier_prime_bold.ttf
	Bold []byte

	//go:embed courier_prime_italic.ttf
	Italic []byte

	//go:embed courier_prime_bolditalic.ttf
	BoldItalic []byte

	//go:embed OFL.txt
	OFL []byte

	//go:embed OFL-FAQ.txt
	OFL_FAQ []byte
)

func ExportFonts() {
	os.WriteFile(RESERVED_NAME + " Regular.ttf",    Regular,    os.ModePerm)
	os.WriteFile(RESERVED_NAME + " Bold.ttf",       Bold,       os.ModePerm)
	os.WriteFile(RESERVED_NAME + " Italic.ttf",     Italic,     os.ModePerm)
	os.WriteFile(RESERVED_NAME + " BoldItalic.ttf", BoldItalic, os.ModePerm)

	os.WriteFile("OFL.txt",     OFL,     os.ModePerm)
	os.WriteFile("OFL-FAQ.txt", OFL_FAQ, os.ModePerm)
}