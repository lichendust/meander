/*
	Meander
	A portable Fountain utility for production writing
	Copyright (C) 2022-2023 Harley Denham

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"os"
	"strings"
	"path/filepath"
)

func load_file(source_file string) (string, bool) {
	bytes, err := os.ReadFile(source_file)
	if err != nil {
		return "", false
	}

	if err != nil {
		return "", false
	}

	return string(bytes), true
}

func load_file_normalise(source_file string) (string, bool) {
	text, ok := load_file(source_file)

	if ok {
		return strings.TrimSpace(normalise_text(text)), ok
	}

	return "", false
}

// fixes paths from command input
func fix_path(input string) string {
	if path, err := filepath.Abs(input); err == nil {
		return path
	} else {
		panic(err)
	}
}

// fixes paths from included files
func include_path(parent, input string) string {
	if !filepath.IsAbs(input) {
		return filepath.Join(filepath.Dir(parent), input)
	}
	return input
}

func find_file(start_location, target string) (string, bool) {
	final_path := ""

	err := filepath.Walk(start_location, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == start_location {
			return nil
		}

		if filepath.Base(path) == target {
			final_path = path
		}

		return nil
	})
	if err != nil {
		return "", false
	}

	return filepath.ToSlash(final_path), final_path != ""
}

// makes working_dir output from input
/*func make_output_from_input(input, target_ext string) string {
	return input[:len(input) - len(filepath.Ext(input))] + target_ext
}*/

// relativises output path based on working dir
/*func get_relative_output(target_path string) string {
	cwd, err := os.Getwd()

	if err != nil {
		return target_path
	}

	path, err := filepath.Rel(cwd, target_path)

	if err != nil {
		return target_path
	}

	return path
}*/

// test for "default" files in working directory
func find_default_file() string {
	files := make([]string, 0, 12)

	// find all the fountain/ftn files in the working directory
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && path != "." {
			return filepath.SkipDir
		}
		if path[0] == '.' {
			return nil
		}

		ext := filepath.Ext(path)

		if ext == "" {
			return nil
		}

		if ext[1] != 'f' && ext != fountain_short_ext && ext != fountain_extension {
			return nil
		}

		files = append(files, path)
		return nil
	})
	if err != nil {
		return ""
	}

	if len(files) == 0 {
		return ""
	}

	// if there's only one fountain file, just choose that one!
	if len(files) == 1 {
		return files[0]
	}

	// otherwise, get more specific and make
	// a sensible selection based on these names
	default_names := [...]string{
		"root",
		"main",
		"master",
		"manifest",
	}

	// if there's more than one, priority is
	// given by this array's order
	for _, name := range files {
		base_name := split_rune(name, '.')[0]
		for _, default_name := range default_names {
			if default_name == base_name {
				return name
			}
		}
	}

	return ""
}

func rewrite_ext(path, new_ext string) string {
	ext := filepath.Ext(path)
	raw := path[:len(path)-len(ext)]

	return raw + new_ext
}

func write_file(path string, content []byte) bool {
	return os.WriteFile(path, content, 0777) == nil
}