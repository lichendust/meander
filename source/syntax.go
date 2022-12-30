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

type fountain_content struct {
	title map[string]string
	nodes []*syntax_node
}

const (
	WHITESPACE uint8 = iota
	PAGE_BREAK
	HEADER
	FOOTER
	PAGE_NUMBER
	SCENE_NUMBER

	ACTION
	LIST
	SCENE
	CHARACTER
	PARENTHETICAL
	DIALOGUE
	LYRIC
	TRANSITION
	CENTERED
	SYNOPSIS

	SECTION
	SECTION2
	SECTION3
)

type syntax_node struct {
	node_type uint8
	level     uint8
	revised   bool
	template  *template_entry
	raw_text  string
	lines     []*syntax_line
}