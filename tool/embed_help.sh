#!/bin/bash

# hard-wraps the "help" command text into the
# codebase to ensure consistency

set -e

printf "package main" > command_help_text.go

for f in help/*.txt; do
	name=$(basename ${f%%.txt})
	data=$(fold -w 64 -s $f)

	printf "\n\nconst comm_$name = \`\n$data\`" >> command_help_text.go
done