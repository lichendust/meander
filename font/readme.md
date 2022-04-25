# Meander and Fonts

## Contents
<!-- MarkdownTOC autolink=true -->

- [A Note on Font Licenses](#a-note-on-font-licenses)
- [Technical Reasoning](#technical-reasoning)
- [Font Seeking](#font-seeking)
- [Conclusion](#conclusion)

<!-- /MarkdownTOC -->

## A Note on Font Licenses

[ofl_link]: https://scripts.sil.org/cms/scripts/page.php?site_id=nrsi&id=OFL

The copy of the Meander source contained in this repository will, if compiled, embed the font family *Courier Prime* inside its standalone binaries.  These fonts are provided here in compliance with the [SIL Open Font License][ofl_link].

Because the [OFL][ofl_link] permits "embedding" in software but is ambiguous as to whether the *type* of embedding performed at compilation time in this repository is in violation of its rules on "Reserved Names", the compiled binaries do not make visible reference to the Reserved Name "Courier Prime [Sans]".

(Though this feels rather more like following the letter of the license rather than its spirit, which I don't like.)

Meander standalone binaries are distributed with the *Courier Prime Sans* OFL license files, which use the Reserved Name insofar as to attribute the original copyright holders; an unambiguous requirement of the OFL regardless of the other considerations above.

## Technical Reasoning

The reasoning behind embedding a font family rather than simply pointing to a distribution is consistency.  Meander aims to support a high number of operating systems and provide exactly equivalent output on all of them, with great care taken to ensure this is the case.

Due to the complexity of font formats and their variadic installations on different operating systems, this is functionally impossible.

Seeking compatible fonts and collecting the requisite style variations from alternate files is incredibly difficult without directly accessing the platform's database through a system-level API.

Aside from exponentially increasing the difficulty of supporting any single platform, it would require CGo and any number of platform-level C/C++ compilers to become dependencies of Meander's compilation.  Any single C compiler is, even to the experienced, a nightmare to make play nicely with unfamiliar repositories — before we mention needing to get Go, CC/GCC and/or Windows Libraries playing nicely together, for instance.

Compiling Meander has been intentionally designed to be as simple and accessible as possible to enable anyone to recompile their own versions with personalised changes, using their own copies of the source.  Even the least technical user should meet with very little difficulty when following compilation instructions provided with the source.

## Font Seeking

Meander offers an experimental method of font-seeking as an alternative to the embedded files *and* operating-system specific support, but for the following reasons, it is also not a solution to the problem.

Because Meander does not use the target system's font database or link to operating system libraries, it must therefore algorithmically locate the files by searching known directories and making an educated guess when identifying the font and related style files in the chosen family.

Generally it "works", but with some notable exceptions like *Courier New* on Windows, which Microsoft has helpfully named, in the default installation, `cour.ttf`.  This plays utter havoc with the results if other Courier derivatives are installed.  Also `couri.ttf`, the italic style, is heuristically closer to the word "Courier" than the base font `cour.ttf`, which leads to a situation where the entire document's italics are inverted.

Also, Meander requires TTF files, due to a limitation in one of its [core libraries](https://github.com/signintech/gopdf).  The algorithmic font seeking described above therefore becomes difficult on the OTF-focused Linux — generally, when people have a font installed, they expect software to be able to use it.

Worse though, it *guarantees* Free/NetBSD, Plan9 and others are simply locked out of the feature because their font formats are too far removed to be used without significant effort from us and the user.  The font seek command is actually disabled on all but the "big three" for this reason.

(I have considered on-the-fly conversions of OTF (or others) into TTF, but they seem out of the window due to OpenType being a superset of TrueType — most TTFs are easily converted to OTF, but not the other way around.)

## Conclusion

In short, instead of having to tailor a custom, specific lookup for every known monospace font (per user request) for every operating system, we choose to embed a sensible default instead.  This allows us to provide better support to all operating systems, ease of use for the casual user who wants to make minor changes for their own purposes and make compilation and long-term maintenance significantly simpler.