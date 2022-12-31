# Meander

Meander is a tiny, single-binary, portable renderer for the production writing markup language [Fountain](https://fountain.io).

Meander has a focus on beautiful formatting on the page, as well as being available and fully functional on as large a number of platforms as possible — most of the highly-regarded industry standard tools are prohibitively expensive simply by virtue of only being available on Apple devices.

Instead, Meander lets you write wherever you like, on whatever platform you like, with any plain-text editor you like.  Or, like some of us, on a bunch of them at once.  You can install it anywhere, run it anywhere and take it anywhere, by design.  It's licensed under the GPL 3.0, to ensure it remains available and modifiable.

In addition to the core Fountain specification, Meander also extends the syntax with clever and worthwhile features from other screenwriting tools, where possible or idiomatic to do so.

The binaries are available from [itch.io](https://qxoko.itch.io/meander) under a "Pay What You Want" model — which includes free!

In spite of this quite scary table of contents, Meander is *extremely* simple to use.  There's just a lot of cool things to cover!

## Table of Contents

<!-- MarkdownTOC autolink="true" -->

- [Usage](#usage)
- [Basic Commands](#basic-commands)
	- [Render](#render)
	- [Merge](#merge)
	- [Gender](#gender)
	- [Data](#data)
	- [Convert](#convert)
		- [⚠️ PDF Files](#%E2%9A%A0%EF%B8%8F-pdf-files)
		- [⚠️ Final Draft Files](#%E2%9A%A0%EF%B8%8F-final-draft-files)
		- [Highland Files](#highland-files)
- [Render Flags](#render-flags)
	- [Scenes](#scenes)
	- [Formats](#formats)
	- [Paper Sizes](#paper-sizes)
	- [Hidden Syntaxes](#hidden-syntaxes)
- [Syntax Extensions](#syntax-extensions)
	- [Text Styling](#text-styling)
	- [Directives](#directives)
		- [Timestamps](#timestamps)
		- [Headers / Footers](#headers--footers)
		- [Counters](#counters)
- [Compilation](#compilation)
- [Experimental Features](#experimental-features)
	- [Starred Revisions](#starred-revisions)
	- [Language Support](#language-support)
- [Attribution](#attribution)

<!-- /MarkdownTOC -->

## Usage

Meander is very simple to use.  Render your first screenplay with —

    $ meander

If there's only one Fountain file in the working directory, Meander will just choose that one.

If you're dealing with multiple files, you can specify the target file with an argument —

    $ meander myfilm.fountain

It will also, regardless of where the command was run from, place the output `myfilm.pdf` alongside the original.

You can then get *really* adventurous by naming the PDF file yourself —

    $ meander myfilm.fountain "My Magnum Opus.pdf"

— though now you'll have to be explicit about where you want that PDF to go.

## Basic Commands

The base Meander commands, which should always be the first argument, are —

+ `render`
+ `merge`
+ `gender`
+ `data`
+ `convert`

There's also the usual self-explanatory stuff —

+ `help`
+ `version`

Full documentation for both Fountain and Meander is available from within the binary using the `help` command.

### Render

    $ meander render

The default, implied option.  Creates a PDF of your input file.  See [Render Flags](#render-flags) for the myriad additional options.

### Merge

    $ meander merge

Meander supports a multi-file workflow using a special `{{include}}` syntax.  Merging collapses your multi-file screenplay into a single text file.  The render command does this automatically, but merging allows you to output the combined plain-text.

Using the directive —

    {{include: scenes/some_file.fountain}}

— somewhere in your Fountain file will cause it to import the contents of the path at that location.  The include paths used are *relative to the file they're written in*.

### Gender

Meander comes with the ability to analyse the genders of your characters, giving you a detailed print-out of how they break down across the whole screenplay and whose voices are heard the most.

Calling —

    $ meander gender some_film.fountain

— will output a terminal-friendly version of the stats for that file (and its included files, if applicable).

![Screenshot of a computer terminal window displaying a breakdown of the lines spoken by characters in the film "Big Fish", with specific focus on their genders](images/meander_gender.webp)

The information backing this analysis comes from a custom [boneyard](https://fountain.io/syntax#section-bone) comment[^1] in the root file of your screenplay.

You can put it anywhere, so if you want to shove it way down at the end, Meander doesn't mind.  If you write in more than one table, it will always use the first one in the text, and tables in included, child files will be ignored at any depth; in English, that means the table must be in the file you target with the `gender` command.

```c
/*
    [gender.male]
    Ashby
    Jenks

    [gender.female]
    Rosemary
    Sissix
    Kizzy

    [gender.<custom>]
    Dr. Chef

    [gender.ignore]
    Market Stall Owner
*/
```

Characters will be assigned the gender from the heading they reside under.  Any number of genders can be specified, useful for non-binary and queer characters, as well as for atypical instances like non-humans in science fiction.

In fact, `male` and `female` aren't programmed in literally — every gender is "custom" in Meander, so you can use whatever terms you'd like instead.

The only reserved word is `ignore`.  Characters assigned to `ignore` will be left out of consideration in the breakdown, useful for single-appearance characters or special cases like "the crowd" shouting back at a main player that probably shouldn't count toward any totals.

Any characters found in the screenplay but _not_ in the gender table will be reported as "unknown" and classified in the statistics under that additional group.

Characters can also have multiple names — `Jess` and `Young Jess`, for example.  By writing each name in with a pipe separating them (see the table example above), all instances of the character's appearances under different names will be combined and handled as if they are one.  The report will use the *first* name provided as **the** name.

Only include the actual gender data in the boneyard, with at least one `[gender.x]` header as the first non-whitespace text inside.  Whitespace, indentation and letter casings are not considered.

### Data

The data command generates a JSON file containing the content of and data about a given Fountain file.

    $ meander data [some_film.fountain] [data.json]

This is provided as a useful (but lossy[^2]) data exchange format.  Rather than conversion to other screenplay tools, this is intended for use with non-screenplay software, such as furnishing production-tracking tools with screenplay metadata or dumping statistics into spreadsheets.

The resulting JSON blob is a dictionary containing four entries:

+ `meta` — information about the version of Meander and the JSON format.
+ `title` — a dictionary of the title page entries.
+ `characters` — a list of all characters in the screenplay, their alternate names and gender from the gender analysis table, as well as the number of lines they actually speak.
+ `content` — a syntactic breakdown list of the screenplay content, with each paragraph or dialogue entry, etc., tagged by its type.

### Convert

Meander can also convert certain formats from other formats into plain-text —

+ `.pdf` files
+ `.fdx` files from [Final Draft](https://www.finaldraft.com)
+ `.highland` files from [Highland 2](https://highland2.app)

```sh
$ meander convert input.fdx
```

Meander will detect the input format (and report back if it doesn't know what to do with it), then output a Fountain file (or files) alongside the original with a matching file name.  You can also override the output path with another argument, as with other commands.

Each of these conversions has some caveats —

#### ⚠️ PDF Files

PDF melting is a notoriously *weird* problem.  There is no guarantee that any text is stored linearly inside the file itself.  Meander *assumes it is*, which can lead to some strange output depending on which program created the file.

In trying to produce valid Fountain syntax, Meander also assumes that the input file actually looks like a screenplay.  For example, it takes into account the spacing of lines on the page to figure out if a chunk of text is dialogue or not.  This may be error prone if the spacing is unexpected or otherwise unlike a 'standard' screenplay.

In the event of major issues, Meander also has a raw text extraction mode which will simply rip all text out of the file without trying to make guesses.

    $ meander convert file.pdf --raw

#### ⚠️ Final Draft Files

For Final Draft, Meander parses the XML structure and attempts to write out a decent approximation in Fountain.  It also adds force-characters to text that it knows Fountain would not recognise as its Final Draft designation.

However, only a limited number of files have been tested, none of which have contained more advanced Final Draft features like page-locking, colours and versioning, which will likely cause Meander to stumble.

(If you're a Final Draft user and can provide example files that demonstrate any issues with Meander's conversion, please reach out!)

#### Highland Files

Meander has no trouble converting Highland files.

The only noteworthy thing is that Highland's `{{include}}` system works slightly differently to Meander's, in that it internally stores its references as macOS filesystem IDs, which are useless on other platforms.  This allows Highland users to include files from all over their filesystem without worrying about relative locations or keeping track of file paths.

Meander handles conversion (and rewriting the include paths) of all these file connections automatically, but in order for it to find everything, it has to go manually look for the files to extract or reference. This means they need to be placed together (with the starting file at the highest level of any folders) and it should have no trouble.

## Render Flags

### Scenes

One necessity when formatting screenplays is the numbering of scenes.  In Fountain, this is done by tacking `#12#` (for example) to the end of a scene heading to denote it as the twelfth.

However, Meander offers more options during rendering —

    $ meander -s input
    $ meander --scene input

+ `input`, the default, simply takes the original `#12#` markers exactly as they are in the input files.
+ `remove` ignores all scene numbers and doesn't print them.  It's as if they never existed.
+ `generate` creates new scene numbers, ignoring existing ones, starting from `1`.  They also increment correctly across included files.

If you choose to use `input`, you're not limited to numbers either — you can go mad with stuff like `#1.3-A#`, provided you write them all in yourself.

### Formats

Meander also offers different formatting options.  Right now, it comes with —

- `screenplay`
- `stageplay`
- `manuscript`
- `graphicnovel`
- `document`

These formats can be specified as part of the title page, in the form `format: screenplay`, but the command line flag will take priority.

    $ meander -f screenplay
    $ meander --format screenplay

(Although, `screenplay` is the default — you don't need to explicitly specify it anywhere).

### Paper Sizes

Meander also supports different paper sizes:

- `A4`
- `USLetter`

Again, the paper size may be included as part of the title page, in the same form `paper: A4`.

    $ meander -p A4
    $ meander --paper A4

Controversially, `A4` is the default.

### Hidden Syntaxes

In some templates, certain syntaxes are hidden by default.  Most of them are intended for use during the writing process for reminders, alternate versions, outlining, bookmarking, etc.

For the screenplay template, these include —

+ `# sections`
+ `= synopses`
+ `[[notes]]`

(For the document and graphic novel templates, Sections are used for headings and page/panel markings respectively.)

During the creative process, printing a draft to take away and read and mull over is incredibly valuable — and so are your notes.

Running Meander with the relevant flags —

    $ meander --notes --synopses --sections

— will ensure one (or all) get printed in distinguished colours, designed to make them obvious when reading.

## Syntax Extensions

As mentioned at the outset of this ridiculously long document, Meander incorporates some neat features of other Fountain tools.

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
go mod tidy
go build -ldflags "-s -w" -trimpath ./source
```

These commands will fetch the dependencies, which are extremely minimal (see just below) and then build the smallest possible binary.  With that, you're done.  There should be a shiny executable in the , all ready to run.

Great care has been taken to minimise the use of libraries in Meander for future-proofedness and maintainability.  We currently only rely on —

+ `gopdf` — [source](https://github.com/signintech/gopdf), which is how Meander writes its PDF files.
+ `isatty` — [source](https://github.com/mattn/go-isatty), which is just used to detect whether we can use colours in terminal outputs.

## Experimental Features

### Starred Revisions

Using version control diffs and tags, Meander can provide starred revision features displaying changes since an arbitrary historical point, allowing screenwriters to automatically mark changes.

Using tags as the historical anchor allows any number of Git/Mercurial revisions between the writer-defined screenplay revisions.  Unlike commits, tags can be moved around.

    $ meander film.fountain -r <tag>

This loads the Fountain files via the relevant version control tool and places stars in the margins based on the results of a diff output between the working copy and the historical tag you specify.  This means two things:

- It's a little slower (hundreds rather than tens of milliseconds), because it has to make external calls to other programs.
- Mercurial or Git become dependencies and must be installed on the system; it's not merely enough to have the repository history.

### Language Support

Meander needs no major work to support other European languages: the parser is already designed to be extended with additional matches for all language-driven identifiers (`int/ext`, `to:`, etc.), though only the standard English Fountain versions are programmed in.

Automatic tags like `(more)` and `(CONT'D)` can be specified in the title page —

    + `more tag: (more)`
    + `cont tag: (CONT'D)`

Therefore... as far as Latin/European languages are concerned, Meander is fully internationalised.  We *could* extend the syntax to include non-English matches, but at the cost of compatibility with other tools.

Technically, Fountain is 'fully internationalised' in that all syntaxes can be forced.  It runs counter to the philosophy of Fountain in general, which is to be natural to write in, but *it can be done* with the tools at hand and is already guaranteed compatible with other Fountain tools.

...but this is Anglo-centric.  Some experimentation is necessary.

The other major issue is that Meander has prioritised portability and cross-platform consistency and its font of choice only supports Latin and European languages, meaning that Cyrillic or Asian versions of Meander would, at minimum, require recompilation.

## Attribution

The `{{include}}` syntax feature was originally from the tiny Python utility [Mountain](https://github.com/mjrusso/mountain), where it used the note syntax `[[include]]`.

Highland would then borrow this idea, using curly braces instead.  Meander has adopted the latter for compatibility, but it still felt important to thank Mountain where they did not.

[^1]: "Magic comments" are generally to be avoided, but this was intentionally designed to play nicely with other Fountain tools while ensuring the gender table can still travel with the screenplay, instead of being fed in by a separate file.

[^2]: 'Lossy' here means inline formatting like \*italics\* are removed.  I've yet to encounter an NLE, production tracker or spreadsheet program that understands Markdown formatters, so Meander prefers to remove them.  This simplifies the JSON file immensely, as each paragraph is single, clean string and no additional work is required to reassemble them in a target format or program.