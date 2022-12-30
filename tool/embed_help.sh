#!/bin/bash

# hard-wraps the "help" command text into the
# codebase to ensure consistency

set -e

target=source/command_help_text.go

printf "$(cat tool/header_license.txt)\n\n" > $target

printf "package main" >> $target

for f in help/*.txt; do
	name=$(basename ${f%%.txt})
	data=$(fold -w 64 -s $f)

	printf "\n\nconst comm_$name = \`\n$data\`" >> $target
done