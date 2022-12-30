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

const title = "Meander 0.1.0"

func main() {
	config, ok := get_arguments()
	if !ok {
		return
	}

	switch config.command {
	case COMMAND_VERSION:
		println(title)

	case COMMAND_HELP:
		command_help()

	case COMMAND_RENDER:
		command_render_document(config)

	case COMMAND_MERGE:
		command_merge_document(config)

	case COMMAND_JSON:
		command_json(config)

	case COMMAND_GENDER:
		command_gender_analysis(config)

	case COMMAND_CONVERT:
		command_convert(config)
	}
}