package font

import _ "embed"

// we need to register the reserved font name
// for the PDF metadata to be correct
const ReservedName = "Courier Prime"

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