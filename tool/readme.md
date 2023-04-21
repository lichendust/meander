# Build Tools

All of Meander's build tools are in this directory.  There aren't many and they're mostly time-savers rather than direct necessities.

Given a full copy of the source, Meander can always be compiled using only —

	go build -ldflags "-s -w" -trimpath

`-ldflags` strips out compiler cruft and debug symbols, halving the size of the binary.  Why this is not the default expression of `go build` is utterly maddening.

`-trimpath` makes sure the Go compiler doesn't leave *information about your personal filesystem on your computer in the binary for everyone to see in the event of a crash*.

## Usage

All of the following tools must be run with the working directory set as the root of the project, therefore called in the form —

	tool/build.sh

## Build

`build.sh` does *everything.*  It cross-compiles Meander for several platforms, packages them with licenses and readmes, then outputs those packages' checksums ready for distribution.

## Embed

`embed.sh` takes each `help_*` file from the `/text/` directory, hard-wraps its content to 64 characters and embeds them as constant strings in the codebase.

This is designed to allow consistent maintenance of the program's internal help messages, helping ensure that text presentation and formatting matches going forwards, and changes are easily visible outside the context of the code itself.

### Guidelines

All files are embedded as constant, raw strings in the format `comm_<filename>`, where `filename` is *sans* extension.  `comm` here is short for "communication".

When creating a new file, it needs to be given a matching switch entry in `command_help` to be accessible to the user, and also be communicated in the main `help` command text in `help.txt`.

Each file can also use a set of shorthand formatters to provide colour output — `$1` and `$0`.  They are direct substitutions for ANSI escape codes, with `$1` being a hard-coded colour and `$0` clearing any effects.

The hard-wrap in the embed script is not aware of these characters, of course, so they count towards the line-total.  Some care must be taken at times to ensure the output isn't borked or poorly wrapped as a result, but this is usually only an issue when the accent usage is heavy-handed.