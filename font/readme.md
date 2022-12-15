# Meander and Fonts

## Contents
<!-- MarkdownTOC autolink=true -->

- [A Note on Font Licenses](#a-note-on-font-licenses)
- [Technical Reasoning](#technical-reasoning)

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