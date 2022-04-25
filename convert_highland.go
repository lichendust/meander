package main

import (
	"io"
	"os"
	"fmt"
	"strings"
	"archive/zip"
)

/*
	Highland files are TextBundle-compatible,
	which means they're a zip file with some
	junk in them.  Highland stores its core,
	plain text data in "text.fountain" at the
	root of the zip, which makes extracting it
	a simple matter of opening the zip file
	and copying the file out to an appropriate
	location

	unfortunately Highland handles its include
	directives by packing them in some
	undocumented binary format that I have had
	poor luck consistently parsing.  It's
	bizarre because all other ancillary data
	in ".highland" is just JSON files.  I do
	know the include file stores the full,
	absolute path to the included document,
	which makes sense — it's what Meander works
	out on the fly to guarantee relative
	accuracy — but inexplicable as to why it's
	stored differently to anything else.
*/

func command_convert_highland(config *config) {
	archive, err := zip.OpenReader(fix_path(config.source_file))

	if err != nil {
		panic(err)
	}

	defer archive.Close()

	for _, f := range archive.File {
		if !strings.HasSuffix(f.Name, "text.fountain") {
			continue
		}

		destination, err := os.OpenFile(fix_path(config.output_file), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)

		if err != nil {
			fmt.Fprintf(os.Stderr, "convert: error extracting %q\n", config.output_file)
			return
		}

		defer destination.Close()

		file, err := f.Open()

		if err != nil {
			fmt.Fprintf(os.Stderr, "convert: error extracting %q\n", config.output_file)
			return
		}
		
		defer file.Close()

		if _, err := io.Copy(destination, file); err != nil {
			fmt.Fprintf(os.Stderr, "convert: error extracting %q\n", config.output_file)
		}
	}
}