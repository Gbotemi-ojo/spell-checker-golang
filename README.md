# Spell Checker (sc)

A Command Line Interface (CLI) written in Go for checking plain text documents for typos. This project is structured to be easily extendable for custom dictionaries, Markdown/HTML parsing, and graphical interfaces.

## Features

- **Fast Lookups:** Uses an in-memory hash map for `O(1)` word lookups against a dictionary.
- **CI/CD Ready:** Exits with a non-zero status code (`1`) if typos are found, making it ideal for automated pipelines.
- **Flexible Input:** Accepts text via stdin piping or direct file/directory paths.
- **Line & Column Tracking:** Accurately reports the exact location (Line and 1-based Column) of detected typos.

## Installation

Ensure you have [Go](https://go.dev/) installed. Clone the repository and build the binary:

```bash
cd sc
go mod tidy
go build -o sc.exe .