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
)

func command_help() {
	println(title)

	args := os.Args[2:]

	if len(args) == 0 {
		println(apply_color(comm_help))
		newline()
		return
	}

	switch strings.ToLower(args[0]) {
	case "fountain":
		println(apply_color(comm_fountain))
		newline()

	case "render":
		println(apply_color(comm_render))
		newline()

	case "merge":
		println(apply_color(comm_merge))
		newline()

	case "gender":
		println(apply_color(comm_gender))
		newline()

	case "data":
		println(apply_color(comm_data))
		newline()

	case "convert":
		println(apply_color(comm_convert))
		newline()
	}
}