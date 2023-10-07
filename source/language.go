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

// language.go is an extension of parser.go that reveals
// the underlying syntax that's matched for; basically we can
// easily extend language support in this file

const GENDER_HEADING         = "Gender Analysis"
const GENDER_CHARS_BY_GENDER = "Character Count by Gender"
const GENDER_LINES_BY_GENDER = "Lines by Gender"
const GENDER_CHARS_BY_LINES  = "Lines by Character"

const DEFAULT_MORE_TAG = "(more)"
const DEFAULT_CONT_TAG = "(CONT'D)"

// text is lowercased
func lang_scene(text string) bool {
	switch text {
	case "int":     return true
	case "ext":     return true
	case "int/ext": return true
	case "ext/int": return true
	case "i/e":     return true
	case "e/i":     return true
	case "est":     return true
	case "scene":   return true
	}
	return false
}

// text is lowercased
func lang_transition(text string) bool {
	return text == "to:"
}

// text is homogenised
func lang_title_page(text string) bool {
	switch text {
	// fountain
	case "title":     return true
	case "credit":    return true
	case "author":    return true
	case "source":    return true
	case "contact":   return true
	case "revision":  return true
	case "copyright": return true
	case "draftdate": return true
	case "notes":     return true

	// meander
	case "paper":   return true
	case "format":  return true
	case "conttag": return true
	case "moretag": return true
	case "header":  return true
	case "footer":  return true
	}
	return false
}