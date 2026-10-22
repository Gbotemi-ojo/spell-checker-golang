# Spell Checker (sc)

A Command Line Interface (CLI) written in Go for checking plain text documents for typos. This project is structured to be easily extendable for custom dictionaries, Markdown/HTML parsing, and graphical interfaces.

## Features

- **Fast Lookups:** Uses an in-memory hash map for `O(1)` word lookups against a dictionary.
- **Custom Words & Phrases:** Specify domain-specific words or multi-word phrases to ignore without triggering typo errors.
- **CI/CD Ready:** Exits with a non-zero status code (`1`) if typos are found, making it ideal for automated pipelines.
- **Flexible Input:** Accepts text via stdin piping or direct file/directory paths.
- **Line & Column Tracking:** Accurately reports the exact location (Line and 1-based Column) of detected typos.

## Installation

Ensure you have [Go](https://go.dev/) installed. Clone the repository and build the binary:

```bash
cd sc
go mod tidy
go build -o sc.exe .
```

## Usage

```bash
sc [-dict path] [-ignore words/phrases] [-ignore-file path] <file-or-directory>...
```

### Options

- `-dict <path>`: Path to a custom dictionary file (one word per line). Defaults to system dictionary paths or `SC_DICT_PATH` environment variable.
- `-ignore <string>`: Custom words or phrases that should not trigger an error. Can be comma-separated or repeated.
- `-ignore-file <path>`: Path to a file containing custom words and phrases (one per line). Comments starting with `#` and blank lines are ignored.
- **Automatic `.scignore`**: If `-ignore-file` is not specified and a `.scignore` file exists in the current directory, it is loaded automatically.

### Examples

#### Ignoring Custom Words

```bash
sc -dict words.txt -ignore "golang,kubernetes,devops" document.txt
```

#### Ignoring Multi-Word Phrases

Multi-word phrases will be ignored when they appear together in the text (even across line breaks or hyphens), while individual words in the phrase will still be flagged if used separately outside the phrase:

```bash
sc -dict words.txt -ignore "status quo,ad hoc,machine learning" document.txt
```

#### Using an Ignore File

Create a `.scignore` file (or specify any file with `-ignore-file`):

```text
# Domain terms
Kubernetes
DevOps

# Phrases
status quo
ad hoc
open source
```

Run the spell checker:

```bash
sc -dict words.txt -ignore-file .scignore document.txt
```

#### Checking via Stdin

```bash
cat document.txt | sc -dict words.txt -ignore "status quo"
```
