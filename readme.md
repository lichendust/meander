# 📝 Meander

Meander is a tiny, single-binary, portable utility for the production writing markup language [Fountain](https://fountain.io).

![A stylised image of the word 'Meander' on a beige paper background, which says 'plain-text screenwriting for everyone, everywhere' beneath it](https://stuff.lichendust.com/media/meander.webp)

Meander has a focus on beautiful formatting on the page, as well as being available and fully functional on as large a number of platforms as possible — most of the highly-regarded industry standard tools are prohibitively expensive simply by virtue of only being available on Apple devices.

Instead, Meander lets you write wherever you like, on whatever platform you like, with any plain-text editor you like.  Or, like some of us, on a bunch of them at once.  You can install it anywhere, run it anywhere and take it anywhere, by design.  It's licensed under the GPL 3.0, to ensure it remains available and modifiable.

Building on top of Fountain, Meander can create all sorts of production documents, including —

- Screenplays
- Stageplays
- Novel Manuscripts

Meander also extends the core syntax with simple, clever and worthwhile features to make it more powerful.

You can [download Meander from itch.io](https://lichendust.itch.io/meander) under a 'pay what you want' model, which includes free.

[<img height="60px" src="https://static.itch.io/images/badge-color.svg">](https://lichendust.itch.io/meander)

In spite of this quite scary table of contents, Meander is *extremely* simple to use.  There's just a lot to cover!

## Contents

<!-- MarkdownTOC autolink="true" -->

- [Usage](#usage)
- [Basic Commands](#basic-commands)
    - [Render](#render)
    - [Merge](#merge)
    - [Gender](#gender)
    - [Data](#data)
    - [Convert](#convert)
- [Render Flags](#render-flags)
    - [Scenes](#scenes)
    - [Formats](#formats)
    - [Paper Sizes](#paper-sizes)
    - [Hidden Syntaxes](#hidden-syntaxes)
- [Syntax Extensions](#syntax-extensions)
    - [Text Styling](#text-styling)
    - [Modifiers](#modifiers)
        - [Includes](#includes)
        - [Headers / Footers](#headers--footers)
    - [Counters](#counters)
    - [Title Page](#title-page)
- [Compilation](#compilation)
- [Editor Support](#editor-support)
- [Future Plans](#future-plans)
- [Attribution](#attribution)

<!-- /MarkdownTOC -->

## Usage

Meander is very simple to use.  Render your first screenplay with —

    meander myfilm.fountain

The output, in this case `myfilm.pdf` will be placed alongside the original.

You can then get *really* adventurous by naming the PDF file yourself —

    meander myfilm.fountain "My Magnum Opus.pdf"

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

Meander's help command is extremely powerful and provides detailed information about every command, flag and setting available within Meander, as well as useful resources like a built-in cheat-sheet for Fountain —

    meander help fountain

### Render

    meander render

The default, implied option.  Creates a PDF of your input file.  See [Render Flags](#render-flags) for the myriad additional options.

### Merge

    meander merge source.fountain [output.fountain]

Meander supports a multi-file workflow.  Merging collapses your multi-file screenplay into a single text file.  The render command does this automatically, but merging allows you to output the combined plain-text as needed.

You'll need to use the `include` syntax in your screenplay to insert another file —

    include: scenes/some_file.fountain

Included paths are *always* relative to the file in which they're written.

### Gender

Meander comes with the ability to analyse the genders of your characters, giving you a detailed print-out of how they break down across the whole screenplay and whose voices are heard the most.

Calling —

    meander gender some_film.fountain

— will output a terminal-friendly version of the stats for that file (and its included files, if applicable).

![Screenshot of a computer terminal window displaying a breakdown of the lines spoken by characters in the film "Big Fish", with specific focus on their genders](https://stuff.lichendust.com/media/meander_gender.webp)

The information backing this analysis comes from a custom [boneyard](https://fountain.io/syntax#section-bone) comment[^1] in the root file of your screenplay.

```c
/*
    [gender.male]
    Ashby | Ashby Santoso | Captain Santoso
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

Characters will be assigned the gender from the heading they reside under.  Any word can used to define a gender: this means you can represent non-binary and queer characters, as well as non-humans in science fiction.  Oh and non-English writers aren't stuck with their reports saying 'Male' and 'Female'.

The only reserved word is `ignore`.  Characters assigned to `ignore` will be left out of consideration in the breakdown, useful for single-appearance characters or special cases like 'the crowd' shouting back at a main player that probably shouldn't count toward any totals.

Any characters found in the screenplay but _not_ in the gender table will be reported as 'unknown' and classified in the statistics under that additional group.

Characters can also have multiple names — `Ashby` and his occasional full name `Ashby Santoso`, for example.  By writing each name in with a pipe separating them (see the table example above), all instances of the character's appearances under different names will be combined and handled as if they are one.  The report will use the *first* name provided as **the** name.

Only include the actual gender data in the boneyard, with at least one `[gender.x]` header as the first non-whitespace text inside.  Whitespace, indentation and letter casings are not considered: the way the name is written in the table is how it will appear in the output.

You can put the gender table anywhere, so if you want to shove it way down at the end, Meander doesn't mind.  If you supply more than one table (such as across multiple included files), those new characters will be combined.

> For instance, you could create a stub file that `includes` and therefore outputs the statistics for a whole season of episodes.

~~Existing characters are not changed to prevent confusion: always define a character in a single location.~~  This was changed in `v0.3.0`, because the pros actually outweighed the cons of this.

### Data

The data command generates a JSON file containing the content of and data about a given Fountain file.

    meander data [some_film.fountain] [data.json]

This is provided as a useful data exchange format.  Rather than conversion to other screenplay tools, this is intended for use with non-screenplay software, such as furnishing production-tracking tools with screenplay metadata or dumping statistics into spreadsheets.

The resulting JSON blob is a dictionary containing four entries —

+ `meta` — information about the version of Meander and the JSON format.
+ `title` — a dictionary of the title page entries.
+ `characters` — a list of all characters in the screenplay, their alternate names and gender from the gender analysis table, as well as the number of lines they actually speak.
+ `content` — a syntactic breakdown list of the screenplay content, with each paragraph or dialogue entry, etc., tagged by its type.

### Convert

Meander can convert `.fdx` files from Final Draft to Fountain.

    meander convert input.fdx

You can override the output path with another argument, as with other commands.

Meander parses the XML structure and attempts to write out a decent approximation in Fountain.  It also adds force-characters to text that it knows Fountain would not recognise as its Final Draft designation.

Because Final Draft has a Fountain importer, Meander does not *export* to `.fdx`.

> This command is currently considered experimental.  I have limited access to example `.fdx` files, especially those demonstrating complex features like page-locking.
>
> Note that Meander's non-standard [syntax extensions](#syntax-extensions) are a [known issue for importing with Final Draft](https://github.com/lichendust/meander/issues/3).  If this is a requirement, you should limit your use of any non-standard syntax for the time being.
>
> Please open an issue for any other import/export concerns.  If there is sufficient need, Meander can easily support its own export mode to support more Final Draft features.

## Render Flags

Almost everything in Meander has a default specification.  For the following sections, the item marked with an asterisk* is the default, implied parameter.  If that's the one you want, you don't need to specify it.

### Scenes

In Fountain, scene numbering is traditionally handled by tacking `#12#` (for example) to the end of a scene heading to denote it as the twelfth.

However, Meander offers more options during rendering —

    meander -s input
    meander --scene input

- `input` uses the original input markers from the text.*
- `remove` removes all of them from the output.
- `generate` creates a new sequence starting from `1`, which increments correctly across multiple files.

If you're not familiar with Fountain, if you choose to write in scene headings manually you're not limited to numbers; you can go mad with stuff like `#1.3-A#`.

### Formats

Meander also offers different formatting options.  Right now, it comes with —

- `screenplay`*
- `stageplay`
- `manuscript`
- `graphicnovel`
- `document`

These formats can be specified as part of the title page, in the form `format: screenplay`, but the command line flag will take priority.

    meander -f screenplay
    meander --format screenplay

### Paper Sizes

Meander also supports different paper sizes —

- `US Letter`*
- `US Legal`
- `A4`

Again, the paper size may be included as part of the title page, in the same form `paper: A4`.

    meander -p A4
    meander --paper A4

### Hidden Syntaxes

In some templates, certain syntaxes are hidden by default.  Most of them are intended for use during the writing process for reminders, alternate versions, outlining, bookmarking, etc.

For the screenplay template, these include —

+ `# sections`
+ `= synopses`
+ `[[notes]]`

(For the manuscript, document and graphic novel templates, Sections are used for chapters, headings and page/panel markings respectively, which means they're enabled by default.)

During the creative process, printing a draft to take away and read and mull over is incredibly valuable — and so are your notes.

Running Meander with the relevant flags —

    meander --notes --synopses --sections

— will ensure they remain printed.

## Syntax Extensions

### Text Styling

The core Fountain spec includes —

- `*italics*`
- `**bold**`
- `***bold italics***`
- `_underlines_`

Meander also includes —

- `~~strikethroughs~~`
- `+highlights+`

### Modifiers

#### Includes

You've already seen includes above, but just to re-iterate: you can import another Fountain file into the current one using the following line —

    include: some_file.fountain

The path is always relative to the file in which the include is written.

#### Headers / Footers

    header: Some Header
    footer: Some Footer

Headers and footers add their contents to the top and bottom of all subsequent pages starting from the page on which their declaration would appear.

You can also specify left, right or centre alignment by using pipe characters —

    header: left | right
    header: left | centre | right

    footer: left only
    footer: | centre only |
    footer: | right only

In fact, the default header for every Meander document is defined like so —

    header: | #PAGE.

(See also [Counters](#counters) below.)

They can also be stopped by leaving them empty —

    header:

You can set a header anywhere in the text, but it will only take effect on the following page: set a new header before a manual page-break then.

Headers and footers are also valid title page elements in Meander, so if you're just setting a single one for the entire document, you should do it this way — it's useful for maintaining compatibility with other Fountain tools.

### Counters

Sometimes, you might want a numerical counter for tracking values across a screenplay, independently of say, the scene numbers or the page count.

Meander's syntax for this is a pound sign `#` followed by a keyword of your choice.  This word should be made of only letters and underscores and is written in ALL CAPS by convention —

    There are #COUNTER apples in the box.

You can also start or reset any counter to an arbitrary value —

    #COUNTER:10

You can also employ alphabetical counters, by initialising them with a letter —

    #COUNTER:A

> Note: you cannot change a counter's type after it has begun.

There are also a small number of built-in counters that are available to use.  None of these counters may be modified or reset —

- `#PAGE` the current page number.
- `#SCENE` the current scene number (only available when using generative scene numbers).
- `#WORDCOUNT` the total word count of the document.

### Title Page

Fountain's title page consists of the following items —

- `title`
- `credit`
- `author`
- `source`
- `notes`
- `draft date`
- `copyright`
- `revision`
- `contact`
- `info`

Meander adds the following items —

- `paper`
- `format`
- `header`
- `footer`
- `more tag`
- `cont tag`

More and cont tags are used to override the default `(more)` and `(CONT'D)` text used when dialogue is broken across a page boundary.  You should specify them inclusive of brackets —

    more tag: (more)
    cont tag: (CONT'D)

Note that in Meander, title page elements are case insensitive and whitespace agnostic: `more tag:` is the same as `MORETAG:`.  This may not be true in every Fountain tool.

## Compilation

Building Meander is super easy.  Install [Go](https://golang.org) — check the `go.mod` file for the most up-to-date information on versions, then clone this repository and run:

```sh
go build -ldflags "-s -w" -trimpath ./source
```

This command will build the smallest possible binary.  With that, you're done.  There should be a shiny executable in your repository, all ready to run.

Great care has been taken to minimise the use of libraries in Meander for future-proofedness and maintainability.  We currently only rely on —

+ `gopdf` — [source](https://github.com/signintech/gopdf), which is how Meander writes its PDF files.
+ `isatty` — [source](https://github.com/mattn/go-isatty), which is just used to detect whether we can use colours in terminal outputs.

All current versions of dependencies are vendored into this repository to defend against unexpected deletion.  Each of these packages are redistributed under their original licenses as defined in each vendor subdirectory.

If you're building for an esoteric platform, like Plan9, Dragonfly, odd BSD flavours or even Android, you are strongly advised to compile Meander yourself.  Only you know the specifics of your hardware or choice of emulator.

Go can compile for all of these targets and more, and you can verify the list of compatible systems and architectures with —

```sh
go tool dist list
```

## Editor Support

While there are several generic packages available for screenwriting with Fountain available for most text editors, I have built first-party support for Meander, its syntax and a number of extra tools into a [Sublime Text package](https://github.com/lichendust/meander-sublime).

Meander for Sublime Text should also be treated as a reference implementation for other packages and further text editor support.

## Future Plans

I plan to add starred revisions, page-locking and expanded language support to Meander.

Most editorial changes of this variety require more syntactic changes to Fountain itself.  Most other tools get around this by moving the goalposts and putting your documents in some arbitrary binary format, which will **never happen** with Meander, because it violates the entire purpose of Fountain in the first place.

Please check the [Ideas tag](https://github.com/lichendust/meander/labels/idea) for my current proposals to this effect and give any suggestions or feedback you may have.

## Attribution

The `include` feature was originally from the tiny Python utility [Mountain](https://github.com/mjrusso/mountain), where it used the note syntax `[[include]]`.

[^1]: 'Magic comments' are generally to be avoided, but this was intentionally designed to play nicely with other Fountain tools while ensuring the gender table can still travel with the screenplay, instead of being fed in by a separate file.
