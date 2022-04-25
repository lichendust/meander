# Build Tools

All of Meander's build tools are in this directory.  There aren't many and they're mostly time-savers rather than direct necessities.

Given a full copy of the source, Meander can always be compiled using only —

	go build -ldflags "-s -w" -trimpath

`-ldflags` strips out compiler cruft and debug symbols, halving the size of the binary.  Why this is not the default expression of `go build` is utterly maddening.

`-trimpath` makes sure the Go compiler doesn't leave *information about your personal filesystem on your computer in the binary for everyone to see in the event of a crash*.  Apparently this is some sort of feature and not an absolutely garbage oversight.

---

All of the following tools must be run with the working directory set as the root of the project, therefore called in the form —

	tool/build.sh

## Build

The `build.sh` and `build.bat` tools are 1:1 copies of each other, providing easy cross-compilation for many platforms simulaneously.

Only the cross-compiler is available in two forms, due to Go taking some time and effort to get running in emulated platforms like Cygwin or WSL.  The other tools can easily be used with any default installation of either.

## Help

`embed_help.sh` takes each plain-text file from the `/help/` directory, hard-wraps its content to 64 characters and embeds them as constant strings in the codebase.

This is designed to allow consistent maintenance of the program's internal help messages, helping ensure that text presentation and formatting matches going forwards, and changes are easily visible outside the context of the code itself.

### Guidelines

All files are embedded as constant, raw strings in the format `comm_<filename>`, where `filename` is *sans* extension.  `comm` here is short for "communication".

When creating a new file, it needs to be given a matching switch entry in `command_help` to be accessible to the user, and also be communicated in the main `help` command text in `help.txt`.

## Package Release

The `package_release.sh` script compiles the various license files and the final binaries for each platform into tidy zip files ready for public release.