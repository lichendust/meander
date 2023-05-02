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

import "encoding/json"

const DATA_VERSION = 1

func command_data(config *config) {
	text, ok := merge(config.source_file)
	if !ok {
		return
	}

	data := init_data()
	syntax_parser(config, data, text)

	if !config.json_keep_formatting {
		// we could do this with reflection... but we don't.
		data.Title.Title     = clean_string(data.Title.Title)
		data.Title.Credit    = clean_string(data.Title.Credit)
		data.Title.Author    = clean_string(data.Title.Author)
		data.Title.Source    = clean_string(data.Title.Source)
		data.Title.Notes     = clean_string(data.Title.Notes)
		data.Title.DraftDate = clean_string(data.Title.DraftDate)
		data.Title.Copyright = clean_string(data.Title.Copyright)
		data.Title.Revision  = clean_string(data.Title.Revision)
		data.Title.Contact   = clean_string(data.Title.Contact)
		data.Title.Info      = clean_string(data.Title.Info)

		for i := range data.Content {
			data.Content[i].Text = clean_string(data.Content[i].Text)
		}
	}

	b, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		eprintln("failed to marshal", config.output_file)
		return
	}

	ok = write_file(config.output_file, b)
	if !ok {
		eprintln("failed to write", config.output_file)
	}
}