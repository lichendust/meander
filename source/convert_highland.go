package main

import (
	"io"
	"os"
	"fmt"
	"bytes"
	"archive/zip"
	"path/filepath"

	ns "github.com/danielpaulus/nskeyedarchiver"
)

func command_convert_highland(config *config) {
	convert(config.source_file)
}

func convert(file string) {
	file = filepath.ToSlash(file)

	println(file)

	archive, err := zip.OpenReader(file)
	if err != nil {
		panic(err)
	}
	defer archive.Close()

	dir := filepath.Dir(file)
	out := rewrite_ext(file, ".fountain")

	for _, f := range archive.File {
		name := filepath.Base(f.Name)

		if name == "includes.dat" {
			file, err := f.Open()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%q\n", f.Name)
				return
			}
			defer file.Close()

			blob, err := io.ReadAll(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%q\n", f.Name)
				return
			}

			x, err := ns.Unarchive(blob)
			if err != nil {
				panic(err)
			}

			the_map := x[0].(map[string]interface{})

			for k := range the_map {
				convert(filepath.Join(dir, k))
			}
		}
	}

	for _, f := range archive.File {
		name := filepath.Base(f.Name)

		if name == "text.md" {
			s, err := f.Open()
			if err != nil {
				panic(err)
			}
			defer s.Close()

			/*d, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
			if err != nil {
				panic(err)
			}
			defer d.Close()*/

			blob, err := io.ReadAll(s)
			if err != nil {
				panic(err)
			}

			blob = bytes.ReplaceAll(blob, []byte(".highland}}"), []byte(".fountain}}"))

			err = os.WriteFile(out, blob, 0777)
			if err != nil {
				panic(err)
			}
		}
	}
}