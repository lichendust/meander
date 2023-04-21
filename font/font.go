package font

import _ "embed"

// we need to register the reserved font name
// for the PDF metadata to be correct
const RESERVED_NAME = "Courier Prime"
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
)