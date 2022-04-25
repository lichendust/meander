# Meander

Meander is a tiny, single-binary, portable renderer for the production writing markup language [Fountain](https://fountain.io).

In addition to the core Fountain specification, Meander also extends the syntax with clever and worthwhile features from other screenwriting tools, where possible or idiomatic to do so.

Meander has a focus on beautiful formatting on the page, as well as being available and fully functional on as large a number of platforms as possible — most of the highly-regarded industry standard tools are prohibitively expensive simply by virtue of only being available on Apple devices.

Instead, Meander lets you write wherever you like, on whatever platform you like, with any plain-text editor you like.  Or, like some of us, on a bunch of them at once.  You can install it anywhere, run it anywhere and take it anywhere, by design.  It's licensed under the GPL 3.0, to ensure it remains available and modifiable.

The binaries are available from [itch.io](https://qxoko.itch.io/meander) under a "Pay What You Want" model — which includes free!

In spite of this quite scary table of contents, Meander is *extremely* simple to use.  There's just a lot of cool things to cover!

## Table of Contents

<!-- MarkdownTOC autolink="true" -->

- [Usage](#usage)
- [Basic Commands](#basic-commands)
	- [Render](#render)
	- [Merge](#merge)
	- [Gender](#gender)
	- [Convert](#convert)
		- [A Note on Highland](#a-note-on-highland)
		- [A Note on Final Draft](#a-note-on-final-draft)
- [Additional Flags](#additional-flags)
	- [Scenes](#scenes)
	- [Formats](#formats)
	- [Paper Sizes](#paper-sizes)
	- [Notes / Synopses](#notes--synopses)
- [Syntax Extensions](#syntax-extensions)
	- [Text Styling](#text-styling)
	- [Directives](#directives)
		- [Timestamps](#timestamps)
		- [Headers / Footers](#headers--footers)
		- [Counters](#counters)
- [Compilation](#compilation)
- [Attribution](#attribution)
- [Planned Features](#planned-features)
	- [Starred Revisions](#starred-revisions)
	- [Language Support](#language-support)
	- [Full Template Overriding](#full-template-overriding)
	- [Additional Paper Sizes](#additional-paper-sizes)

<!-- /MarkdownTOC -->

## Usage

Meander is very simple to use.  Render your first screenplay with —

	$ meander

If there's only one Fountain file in the working directory, Meander will just choose that one.  Otherwise, it will give priority to files named `main` or `manifest` or `root` — a monolithic file where you might use [include](#includes) to combine individual scene files into a full screenplay.

If your screenplay *isn't* named `main`, which it likely isn't, you can specify the target file with an argument —

	$ meander myfilm.fountain

This will create a file `myfilm.pdf` alongside the original file, regardless of the current working directory.

You can then get *really* adventurous by naming the PDF file yourself —

	$ meander myfilm.fountain "My Magnum Opus.pdf"

— though you'll need to be explicit about the full path of the output, otherwise it will put it wherever you happen to be.

## Basic Commands

The base Meander commands, which should always be the first argument, are —

+ `render`
+ `merge`
+ `gender`

There is also a detailed `help` command which teaches you how to use itself.

### Render

	$ meander render

The default, implied option.  Creates a PDF of your input file.

### Merge

	$ meander merge

Meander supports a multi-file workflow using a special `{{include}}` syntax.  Merging collapses your multi-file screenplay into a single text file.  Rendering does this automatically, but that doesn't mean you don't need to do it in plain text too.

Using the directive —

	{{include: scenes/some_file.fountain}}

— somewhere in your Fountain file will cause it to import the contents of the path at that location.  The include paths used are *relative to the file they're written in*, not where Meander is being run from.

⚠ Includes in Meander are infinitely recursive, meaning you can turtle all the way down with child files as many levels as makes you happy, but **they're not cycle-safe**.  You will get a stack overflow or an infinite loop where your computer runs out of memory if you circularly include files.

### Gender

Meander comes with the ability to analyse the genders of your characters, giving you a detailed print-out of how they break down across the whole screenplay and whose voices are heard the most.

Calling —

	$ meander gender some_film.fountain

— will output a terminal-friendly version of the stats for that file (and its included files, if applicable).

![Screenshot of a computer terminal window displaying a breakdown of the lines spoken by characters in the film "Big Fish", with specific focus on their genders](images/meander_gender.webp)

The information backing this analysis comes from a custom [boneyard](https://fountain.io/syntax#section-bone) comment[^1] in the root file of your screenplay.

You can put it anywhere, so if you want to shove it way down at the end, Meander doesn't mind.  If you write in more than one table, it will always use the first one in the text, and tables in included, child files will be ignored at any depth; in English, that means the table must be in the file you target with the `gender` command.

```c
title:  Some Great Film
credit: by
author: Some Dude

/*
	[gender.male]
	John

	[gender.female]
	Jess | Young Jess
	Jane

	[gender.<custom>]
	Jesse
	Dylan

	[gender.ignore]
	Some Dude
*/
```

Characters will be assigned the gender from the heading they reside under.  Any number of genders can be specified, useful for non-binary and queer characters, as well as for atypical instances like non-humans in science fiction.

In fact, `male` and `female` aren't programmed in literally — every gender is "custom" in Meander, so you can use whatever terms you'd like instead.

The only reserved word is `ignore`.  Characters assigned to `ignore` will be left out of consideration in the breakdown, useful for single-appearance characters or special cases like "the crowd" shouting back at a main player that probably shouldn't count toward any totals.

Any characters found in the screenplay but _not_ in the gender table will be reported as "unknown" and classified in the statistics under that additional group.

Characters can also have multiple names — `Jess` and `Young Jess`, for example.  By writing each name in with a pipe separating them (see the table example above), all instances of the character's appearances under different names will be combined and handled as if they are one.  The report will use the *first* name provided as **the** name.

Only include the actual gender data in the boneyard, with at least one `[gender.x]` header as the first non-whitespace text inside.  Whitespace, indentation and letter casings are not considered.

### Convert

Meander can also convert certain formats from other screenwriting tools into plain text —

+ `.highland` files from [Highland 2](https://highland2.app).
+ `.fdx` files from [Final Draft](https://www.finaldraft.com).

```
$ meander convert input.fdx
```

Meander will detect the input format (and report back if it doesn't know what to do with it), then output a Fountain file alongside the original with a matching file name.  You can also override the output path with another argument, as with other commands.

⚠ Note that, as with all inter-app conversions, Meander *does its best*.  There are some specific considerations for the different formats below.

#### A Note on Highland

For Highland — which is TextBundle based — it simply extracts the plain text exactly as it is displayed within the original editor.  This, of course, works without fail.

However, Highland's native `{{include}}` system, while syntactically compatible with Meander, is not guaranteed to work correctly due to the file references being stored in a bizarre undocumented binary format (the only part of the Highland format which is not plain-text and which I have had little luck reverse engineering).

This means Meander cannot recursively convert all files involved in a screenplay because it cannot find them.  It also cannot automatically correct the include paths because those paths are "locked away".

In short, the basic extraction handler has been left in Meander for now because it *is* faster than manually unzipping the Highland file, but it lacks the quality of life feature that is automagically converting all connected/included files.

#### A Note on Final Draft

For Final Draft — which is XML based — Meander parses the XML and attempts to write out a decent approximation in Fountain.  It also adds force-characters to text that it knows Fountain would not recognise as its Final Draft designation.

As far as testing goes, it seems to work very accurately on basic screenplays (and seems to convert FDX files better than some other tools tested).  However, only a limited number of files have been tested, none of which have contained more advanced Final Draft features like page-locking, colours and versioning, which will likely cause Meander to stumble.

## Additional Flags

### Scenes

One necessity when formatting screenplays is the numbering of scenes.  In Fountain, this is done by tacking `#12#` (for example) to the end of a scene heading to denote it as the twelfth.

However, Meander offers more options during rendering —

	$ meander -s input
	$ meander --scene input

+ `input`, the default, simply takes the original `#12#` markers exactly as they are in the input files.
+ `remove` ignores all scene numbers and doesn't print them.  It's as if they never existed.
+ `generate` creates new scene numbers, ignoring existing ones, starting from `1`.  They also increment correctly across multi-file screenplays.

If you choose to use `input`, you're not limited to numbers either — you can go mad with stuff like `#1.3-A#`, provided you write them all in yourself.

### Formats

Meander also offers different formatting options.  Right now, it comes with —

- `screenplay`
- `stageplay`
- `manuscript`
- `graphicnovel`

These formats can be specified as part of the title page, in the form `format: screenplay`, but the command line flag will take priority.

	$ meander -f screenplay
	$ meander --format screenplay

(Although, `screenplay` is the default — you don't need to explicitly specify it anywhere).

### Paper Sizes

Meander also supports different paper sizes:

- `A4`
- `USLetter`

Again, the paper size may be included as part of the title page, in the same form `paper: a4`.

	$ meander -p A4
	$ meander --paper A4

Controversially, `A4` is the default.

### Notes / Synopses

+ `[[notes]]`
+ `= synopses`

— are hidden Fountain elements, used during of the writing process for reminders, alternative versions, outlining and other such things.  They're normally stripped out very neatly when finishing a document for print.

During the creative process though, printing a draft to take away and read and mull over is incredibly valuable — and so are your notes.

Running Meander with the relevant flags —

	$ meander --notes --synopses

— will ensure either (or both) get printed in distinguished colours, designed to make them obvious when reading.

## Syntax Extensions

As mentioned at the outset of this ridiculously long document, Meander incorporates some Neat™ features of other Fountain tools that it deems interesting.

### Text Styling

The core Fountain spec includes —

+ `*italics*`
+ `**bold**`
+ `***bold italics***`
+ `_underlines_`.

Meander also includes —

+ `~~strikethroughs~~`
+ `+highlights+`

### Directives

You've already seen the `{{include}}` directive above in the [Merge](#merge) command, but Meander includes a few other directives.

#### Timestamps

	{{timestamp: dd MM yyyy}}

Timestamps embed the date, per the supplied template (or the sensible default) at the time the file is rendered.  You can use them anywhere in the text.

#### Headers / Footers

	{{header: Some Header}}
	{{footer: Some Footer}}

Headers and footers add their contents to the top and bottom of all subsequent pages starting from the page on which their declaration would appear.  In Meander, they can be stopped by leaving them empty — `{{header}}` — or using the Highland-compatible syntax `{{header: %none}}`.

They can also include the page number using `%p` as a placeholder, or the document title using `%title`.  The latter includes any formatting specified in the title page.

#### Counters

Sometimes, numerical counters are useful for tracking values across a screenplay, independently of say, the scene numbers or the page count.

Meander has four such directives, compatible with Highland's —

	{{series}}
	{{panel}}
	{{figure}}
	{{chapter}}

The counters can be used anywhere in text and will be replaced with an incrementing number.  You can reset the counter to an arbitrary value by using the syntax —

	{{series: 10}}

In a similar vein, the current page number can also be reset by using —

	{{pagenumber: 1}}

## Compilation

Building Meander is super easy.  Install [Go](https://golang.org) — check the `go.mod` file for the most up-to-date information on versions, then clone this repository and run:

```sh
go mod tidy                    # to fetch libraries
go build -ldflags "-s -w"      # strips garbage from the binary
```

You're done.  There should be a shiny executable in the build directory, all ready to run.

Great care has been taken to minimise the use of libraries in Meander for future-proofedness and maintainability.  We currently only rely on —

+ `gopdf` — [source](https://github.com/signintech/gopdf)
+ `isatty` — [source](https://github.com/mattn/go-isatty)
+ `levenshtein` — [source](https://github.com/agnivade/levenshtein)

## Attribution

The `{{include}}` syntax feature was originally from the tiny Python utility [Mountain](https://github.com/mjrusso/mountain), where it used the note syntax `[[include]]`.

Highland would then borrow this idea, using curly braces instead.  Meander has adopted the latter for compatibility, but it still felt important to thank Mountain where they did not.

## Planned Features

### Starred Revisions

Using version control diffs and tags, Meander can provide starred revision features displaying changes since an arbitrary historical point, allowing screenwriters to automatically mark changes.

Using tags as the historical anchor allows any number of Git/Mercurial revisions between the writer-defined screenplay revisions.

This feature is partially implemented in that all of the necessary components exist within the source, but the feature is temporarily disabled due to frequent bugs that are difficult to test.

### Language Support

Adding support for multiple languages is a priority for Meander.  No work is necessary to extend Meander's Fountain parser to support internationalised terms in language-driven matches (such as `INT.` in scene headings).

Automatic tags like `(MORE)` and `(CONT'D)` are slightly more complex, with automatic substitution and recognition being used to ensure Meander doesn't double up on any that have been manually placed, but a working solution is certainly possible with a little effort.  (Unlike the parser, this is currently hard-coded.)

### Full Template Overriding

Template overriding would be provided by a TOML-adjacent file format with custom values injected to provide standard measurements — "inch", "pica", "point", etc.

Built-in templates would also be exportable to be modified by the user, or brand new ones able to be specified for any desired layout.

```
[scene]
casing = upper

[section]
skip = true

[character]
margin = inch * 2.5
```

### Additional Paper Sizes

Available upon request!  It's not obvious which additional paper sizes may be most useful to anyone, with A4 and US Letter being the geographical standards for production documents worldwide, so please submit an issue if this seems a critical oversight.

[^1]: "Magic comments" are generally to be avoided, but this was intentionally designed to play nicely with other Fountain tools while ensuring the gender table can still travel with the screenplay, instead of being fed in by a separate file.