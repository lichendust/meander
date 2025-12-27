# Changelog

## 0.3.0

In Progress

### Features

- Added starred revisions and an accompanying syntax: `@pink`.
- Added user-editable templating.
- Added plain-text archival mode.

### Bugs

- Fixed an edge case where punctuation could be orphaned by line-wrapping if the preceding word was a different font-style.
- Fixed a bug where certain markup characters were still treated as markup (and thus disappeared) when used in ways that should print regular characters.
- Fixed the final page not including a footer.
- Fixed an edge case where text colour did not reset to black.
- Fixed indentation styles being mistakenly applied to the second half of a paragraph after a page-break.
- Fixed indentation being mistakenly applied to extraordinarily short lines of dialogue or all-caps yelling.

### Internal

- Rewrote templating constructor to support templates.
- Rewrote argument parser to more easily support variable numbers of arguments after a flag.
- Updated dependencies to latest versions.
- Updated to latest Go version (no code changes).

## 0.2.3

Patch Release

### Bugs

- Fixed false-positive declarations (headers footers, etc.), once confirmed to be false, *always* reverting to action elements because of an obtuse control-flow fail.
- Fixed dual-dialogue page breaks sometimes creating two pages for themselves.
- Fixed dual-dialogue page breaks sometimes doing a fun bonus increment of the `#PAGE` value.
- Fixed dual-dialogue page-breaks sometimes omitting the header and footer.

## 0.2.2

Patch Release

### Bugs

- Fixed all Section levels using the Section 1 templating instead of their own.
- Fixed headers/footers being missing on a certain page-break case.
- Fixed array bound issue in parser — resolves #4.

## 0.2.1

Patch Release

### Bugs

- Fixed a dumb typo that caused generative scene numbers to print nonsense.

## 0.2.0

Meander's first public release!
