package main

import (
	"os"
	"fmt"
	"bytes"
	"strings"
	"unicode/utf8"
	"path/filepath"

	"github.com/signintech/gopdf"

	_ "embed"
)

const reserved_name = "Courier Prime"

var (
	//go:embed font/courier_prime_regular.ttf
	Regular []byte

	//go:embed font/courier_prime_bold.ttf
	Bold []byte

	//go:embed font/courier_prime_italic.ttf
	Italic []byte

	//go:embed font/courier_prime_bolditalic.ttf
	BoldItalic []byte
)

// attach the embedded to the gopdf document
func register_default_fonts(document *gopdf.GoPdf) bool {
	document.AddTTFFontByReader(reserved_name, bytes.NewReader(Regular))
	document.AddTTFFontByReaderWithOption(reserved_name, bytes.NewReader(Bold),       gopdf.TtfOption { Style: gopdf.Bold })
	document.AddTTFFontByReaderWithOption(reserved_name, bytes.NewReader(Italic),     gopdf.TtfOption { Style: gopdf.Italic })
	document.AddTTFFontByReaderWithOption(reserved_name, bytes.NewReader(BoldItalic), gopdf.TtfOption { Style: gopdf.Italic | gopdf.Bold })

	return true
}

// try and find some custom (system/installed) fonts to attach to the document
func register_custom_fonts(document *gopdf.GoPdf, font_name string) bool {
	results := find_font(font_name)

	// debug_fonts(font_name, results)

	for _, c := range results {
		if c == "" {
			fmt.Fprintf(os.Stderr, "font: couldn't find all styles in family %q\n", font_name)
			return false
		}
	}

	{
		has_error := false

		if err := document.AddTTFFont(font_name, results["regular"]); err != nil {
			has_error = true
		}
		if err := document.AddTTFFontWithOption(font_name, results["bold"], gopdf.TtfOption { Style: gopdf.Bold }); err != nil {
			has_error = true
		}
		if err := document.AddTTFFontWithOption(font_name, results["italic"], gopdf.TtfOption { Style: gopdf.Italic }); err != nil {
			has_error = true
		}
		if err := document.AddTTFFontWithOption(font_name, results["bolditalic"], gopdf.TtfOption { Style: gopdf.Italic | gopdf.Bold }); err != nil {
			has_error = true
		}

		if has_error {
			fmt.Fprintf(os.Stderr, "font: couldn't load all styles in family %q\n", font_name)
			return false
		}
	}

	return true
}

/*
	@note @todo @warning

	everything below here is (bad) experimental code for
	platform-independent font-seeking, without having to
	call on operating system features (in the case of
	Windows, requiring cgo) to make it work.

	while the attempt isn't... terrible? it fails a high
	proportion of the time, most notably on Windows where
	default font installations have impractically modified
	names like "couri.ttf" for original Courier's Italic style.

	other than adding a tonne of complexity (and drastically
	impacting speed) for little tangible benefit, the best
	(and horribly less-than-ideal) solution right now is
	simply to recompile Meander with the user's own font
	selection.
*/

// tries to find a matching set of font files installed on a system,
// returning them in a map of weight/style strings against filepaths
func find_font(name string) map[string]string {

	// normalise the input name
	name = strings.ToLower(name)

	// list of initial ttf files
	initial_list := make([]string, 0, 128)

	// "system_dirs" is defined in the platform-specific files adjacent to this one
	// such as font_darwin, font_windows, etc.
	for _, dir := range system_dirs {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err == nil {
				// don't check literal directory paths (but still recurse into them)
				if info.IsDir() {
					return nil
				}
				// if the returned file is a ttf, bank 'em
				if strings.HasSuffix(path, ".ttf") {
					initial_list = append(initial_list, path)
				}
			}
			return nil
		})
	}

	// individual, pared-down lists
	italic     := make([]string, 0, 16)
	bold       := make([]string, 0, 16)
	bolditalic := make([]string, 0, 16)
	regular    := make([]string, 0, 16)

	everything := make([]string, 0, 16)

	/*
		levenshtein analyses two given strings and defines
		the number of "differences" between them.

		the levenshtein distance of strings is a component
		(or at least semantically related to other
		algorithms used) of "fuzzy finding".

		in this example, we define a maximum of two
		(2) differences as being a valid match (same as
		the code) below.

		given a found file            "path/InputMono-Regular.ttf"
		and a user input              "Input Mono"
		we attempt to match them.

		1. basename
			"InputMono-Regular.ttf"

		2. cutting the basename off at the length of the input
			"Input Mono"
			"InputMono-"

		3. levenshtein result is 2, which we arbitrarily consider
			a match!
	*/

	// this loop implements the explanation above
	for _, file := range initial_list {
		// basename filepath
		lower_file := filepath.Base(strings.ToLower(file))
		direct     := lower_file

		// chop the end off the basename
		if len(name) < len(lower_file) {
			direct = direct[:len(name)]
		}

		if levenshtein_distance(direct, name) < 3 {
			// if we're satisfied that we got a match...

			// register this match in the "all matches" list
			everything = append(everything, file)

			// do this... for some reason? there's a good reason
			// but I didn't document it at the time like a fool.
			lower_no_spaces := strings.ReplaceAll(lower_file, " ", "")

			// check for italics based on naming conventions
			if strings.Contains(lower_no_spaces, "italic") || strings.Contains(lower_no_spaces, "oblique") {
				// potentially escalate to bold italics
				if strings.Contains(lower_no_spaces, "bold") {
					bolditalic = append(bolditalic, file)
					continue
				}
				// or just save it as italic
				italic = append(italic, file)
				continue
			}

			// bolditalics were caught above so we just
			// save these as bold
			if strings.Contains(lower_no_spaces, "bold") {
				bold = append(bold, file)
				continue
			}

			// the literal "regular" is quite common in filenames
			// so just grab that.
			if strings.Contains(lower_no_spaces, "regular") {
				regular = append(regular, file)
				continue
			}
		}
	}

	// create the final map
	final := make(map[string]string, 4)

	/*
		anecdotally, choosing the shortest one was
		(weirdly consistently) the best thing to do given
		results from the levenshtein triage

		you can test this by fmt-printing the arrays above

		unfortunately, it's garbage-in garbage-out, so if
		levenshtein gave us bogus matches (which usually
		means the requested font isn't even installed) then
		this really badly compounds the problem

		these assign empty strings if the input array is empty
	*/
	final["italic"]     = give_shortest(italic,     "italic")
	final["bold"]       = give_shortest(bold,       "bold")
	final["bolditalic"] = give_shortest(bolditalic, "bolditalic")
	final["regular"]    = give_shortest(regular,    "regular")

	/*
		if "regular" is an empty string, try the "all matches"
		array as a last ditch

		this seems _insane_ but what it actually does is catch
		(9/10 times) cases where the base font isn't literally
		named "xxx-regular.ttf" and rather just "xxx.ttf" —
		they are almost always the shortest real match
	*/
	if x := final["regular"]; x == "" {
		final["regular"] = give_clean_shortest(everything)
	}

	return final
}

// give the literal shortest match in the array, with no other
// checking
func give_clean_shortest(input []string) string {
	if len(input) == 0 {
		return ""
	}

	x := 0
	shortest := len(input[0])

	for i, entry := range input {
		lower_file := filepath.Base(strings.ToLower(entry))

		n := len(lower_file)

		if n < shortest {
			x = i
			shortest = n
		}
	}

	return input[x]
}

// give shortest string in array given a keyword that weights
// the check
func give_shortest(input []string, keyword string) string {
	if len(input) == 0 {
		return ""
	}

	x := 0
	shortest := len(input[0])

	is_regular := (keyword == "regular")

	for i, entry := range input {
		// delete the keyword from the filepath to find the
		// shortest, which helps us remove random edge cases
		// where a font has an expectedly long name despite
		// being the "base" form,
		// like "italic-regular-with-ligatures" or some other
		// arbitrary nonsense

		lower_file := filepath.Base(strings.ToLower(entry))
		direct     := strings.ReplaceAll(lower_file, keyword, "")

		n := len(direct)

		if n < shortest {
			x = i
			shortest = n
		}

		// this second step also specifically handles cases
		// involving the term "regular", where "regular-italic"
		// is actually the match we want, but it's longer than,
		// say, "100-italic"

		// this also pares out files called "italic
		// regular" from being accidentally returned
		// as "regular" instead of "italic".
		if !is_regular && strings.Contains(lower_file, "regular") {
			return entry
		}
	}

	return input[x]
}

// levenshtein implementation taken from
// https://github.com/agnivade/levenshtein [MIT]
const alloc_threshold = 32

// Works on runes (Unicode code points) but does not normalize
// the input strings. See https://blog.golang.org/normalization
// and the golang.org/x/text/unicode/norm package.
func levenshtein_distance(a, b string) int {
	if len(a) == 0 {
		return utf8.RuneCountInString(b)
	}
	if len(b) == 0 {
		return utf8.RuneCountInString(a)
	}
	if a == b {
		return 0
	}

	// We need to convert to []rune if the strings are non-ASCII.
	// This could be avoided by using utf8.RuneCountInString
	// and then doing some juggling with rune indices,
	// but leads to far more bounds checks. It is a reasonable trade-off.
	string_one := []rune(a)
	string_two := []rune(b)

	// swap to save some memory O(min(a,b)) instead of O(a)
	if len(string_one) > len(string_two) {
		string_one, string_two = string_two, string_one
	}

	len_one := len(string_one)
	len_two := len(string_two)

	// Init the row.
	var x []uint16
	if len_one + 1 > alloc_threshold {
		x = make([]uint16, len_one + 1)
	} else {
		// We make a small optimization here for small strings.
		// Because a slice of constant length is effectively an array,
		// it does not allocate. So we can re-slice it to the right length
		// as long as it is below a desired threshold.
		x = make([]uint16, alloc_threshold)
		x = x[:len_one + 1]
	}

	// we start from 1 because index 0 is already 0.
	for i := 1; i < len(x); i++ {
		x[i] = uint16(i)
	}

	// make a dummy bounds check to prevent the 2 bounds check down below.
	// The one inside the loop is particularly costly.
	_ = x[len_one]

	// fill in the rest
	for i := 1; i <= len_two; i++ {
		prev := uint16(i)
		for j := 1; j <= len_one; j++ {
			current := x[j - 1] // match
			if string_two[i - 1] != string_one[j - 1] {
				current = min(min(x[j - 1] + 1, prev + 1), x[j] + 1)
			}
			x[j - 1] = prev
			prev = current
		}
		x[len_one] = prev
	}
	return int(x[len_one])
}

func min(a, b uint16) uint16 {
	if a < b {
		return a
	}
	return b
}