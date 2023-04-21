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

import "path/filepath"

func command_convert(config *config) {
	ext := filepath.Ext(config.source_file)

	switch ext {
	case highland_extension:
		convert_highland(config)

	case final_draft_extension:
		convert_final_draft(config)

	/*case ".pdf":
		convert_pdf(config)*/

	case fountain_extension, fountain_short_ext:
		eprintf("convert: %q is already a Fountain file", config.source_file)

	default:
		eprintf("convert: no handler for filetype %q", ext)
	}
}