# Meander and Fonts

## Contents
<!-- MarkdownTOC autolink=true -->

- [These Fonts](#these-fonts)
- [Reasoning](#reasoning)
- [Recompiling](#recompiling)
- [The OFL and Font Distribution](#the-ofl-and-font-distribution)

<!-- /MarkdownTOC -->

[ofl_link]: https://scripts.sil.org/cms/scripts/page.php?site_id=nrsi&id=OFL

Rather than linking to fonts installed by a user on their operating system, Meander is distributed with a single font family bundled directly into the binary, which cannot be changed or overridden — unless you recompile.

This decision also leads to a weird semi-edge case in the [SIL Open Font License][ofl_link], and the steps Meander takes to handle this are [outlined here too](the-ofl-and-font-distribution).

## These Fonts

Before we begin: you can obtain a copy of *Courier Prime*  from [Google Fonts](https://fonts.google.com/specimen/Courier+Prime) or from the creators at [Quote-Unquote Apps](https://quoteunquoteapps.com/courierprime/).

## Reasoning

The reasoning behind bundling a font family rather than simply pointing to a distribution is consistency.  Meander aims to support a high number of operating systems and provide exactly equivalent output on all of them, with great care taken to ensure this is the case.

Due to the complexity of font formats and their variadic installations on different operating systems, this is functionally impossible.

Meander supports every operating system and instruction set its parent language's compiler, [Go](https://go.dev), supports, which is all the ones you've heard of and several you haven't.

This is an explicit, immutable goal of Meander, and it achieves this by being extremely strict about what it links to and what libraries and APIs it depends on.

Introducing 'proper' font sourcing on a platform requires explicit, tailored support and the introduction of large dependencies, additional compilers and the loss of other features.

Not all platforms, either morally or technically, support TTF fonts, let alone provide a way to install and retrieve them with a simple interface and API.  There's very little consistency in the *existence* of mechanisms for retrieving fonts, let alone semantic consistency in how they might work, or even internal consistency in how fonts are stored and named and (looking at you Windows).

By using a 'canonical' Meander font, for which we've chosen *Courier Prime Sans*, we can ensure that every platform we target can be guaranteed to work, we remove the need to validate hundreds of variable-quality fonts and makes Meander 100% portable and idempotent.

To quote a tragically quotable scumbag:

> Any customer can have a [Meander] painted any color that he wants so long as it is [Courier Prime Sans].

## Recompiling

As mentioned, the relative simplicity of compilation *because* of these choices — and the open nature of Meander — means that you can, in fact, just recompile it with your own choice of **monospace** font in seconds.

(No, really, Meander isn't designed at all for variable-width fonts and creates some very funny output if you try.)

## The OFL and Font Distribution

The [SIL Open Font License][ofl_link][^1] is weird and vague.  In several cases, the FAQ contradicts the licence text itself and more than once internally contradicts itself within the same document.

The way in which Meander distributes the fonts that are bundled within the binary by the compiler resides in a poorly-defined area of the OFL.  From detailed reading, I consider Meander's form of distribution to be 'bundling', which is a term that the OFL defines through context clues alone.  Helpful.

To that end, the utmost care has been taken with Meander to bundle and distribute its accompanying fonts (under the philosophical goals of total platform agnosticism defined above) in the most compliant way possible within the oft-wishy washy framework of the OFL.  Meander tries, where possible, to follow the spirit of the OFL as well as the letter.

FAQ 1.10 expressly admits that the OFL.txt text does not need to accompany the fonts when they are bundled within software, Meander does not distribute them nakedly in its binary release packages (see `tool/build.sh`).

This is upheld by Meander deliberately to prevent user confusion: a copy of the OFL without font files immediately visible to the user is more likely to cause furrowed brows than anything else.

FAQ 1.21, on the subject of closed distributions that contain fonts, states that the user cannot be legally prevented from extracting the fonts themselves; their open licence supercedes the software with with they're distributed.

To that effect, Meander provides a 'vomit' command in the form of —

	meander fonts

— which will cause Meander to write its bundled fonts *and* their OFL and OFL-FAQ text to disk, thereby providing a complete, licence-compliant means of distribution.  Meander, then, acts as an unnecessarily complex zip file in this case.

Meander also provides a `credit` command that provides full attribution for the bundled fonts.

[^1]: Their URL structure is also insane and that link is definitely going to die, so shout at me if it's ever broken.