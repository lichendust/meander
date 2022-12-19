package main

import "fmt"

const title = "Meander 0.1.0RC"

const (
	// always all lowercase
	default_template = "screenplay"
	default_paper    = "a4"
)

func main() {
	config, ok := get_arguments()

	if !ok {
		return
	}

	switch config.command {
	case COMMAND_HELP:
		fmt.Println(title)
		command_help()

	case COMMAND_VERSION:
		fmt.Println(title)

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