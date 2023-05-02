package main

const (
	default_more_tag = "(more)"
	default_cont_tag = "(CONT'D)"
)

// all declarations in here should be lowercase
var valid_scene = map[string]bool{
	"int":     true,
	"ext":     true,
	"int/ext": true,
	"ext/int": true,
	"i/e":     true,
	"e/i":     true,
	"est":     true,
	"scene":   true,
}

var valid_transition = map[string]bool{
	"to:": true,
}