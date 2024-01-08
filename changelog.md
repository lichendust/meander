# Changelog

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
