package font

import _ "embed"

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