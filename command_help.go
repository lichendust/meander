package main

import (
	"os"
	"fmt"
	"strings"
	"github.com/mattn/go-isatty"
)

var running_in_term = false

func init() {
	running_in_term = isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

func command_help() {
	defer fmt.Println() // trailing newline

	args := os.Args[2:]

	if len(args) == 0 {
		fmt.Println(apply_color(comm_help))
		return
	}

	switch strings.ToLower(args[0]) {
	case "fountain":
		fmt.Println(apply_color(comm_fountain))

	case "render":
		fmt.Println(apply_color(comm_render))

	case "merge":
		fmt.Println(apply_color(comm_merge))

	case "gender":
		fmt.Println(apply_color(comm_gender))

	case "convert":
		fmt.Println(apply_color(comm_convert))
	}
}

const ansi_color_reset  = "\033[0m"
const ansi_color_accent = "\033[92m"

func apply_color(input string) string {
	if running_in_term {
		input = strings.ReplaceAll(input, "$0", ansi_color_reset)
		input = strings.ReplaceAll(input, "$1", ansi_color_accent)
		return input
	}
	return strip_color(input)
}

func strip_color(input string) string {
	input = strings.ReplaceAll(input, "$0", "")
	input = strings.ReplaceAll(input, "$1", "")
	return input
}