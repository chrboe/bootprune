# bootprune

A tool to remove old kernels (and their associated files) from `/boot`.

## Installation

```
$ go get github.com/chrboe/bootprune
```

## Usage

```
$ bootprune
```

The tool scans `/boot` for Linux kernels (currently using the very primitive
heuristic of checking if the file starts with `vmlinuz-`) and deducts a list of
installed kernel versions from that. Then the user gets to choose which versions
to keep and which to delete.

Currently only "interactive" mode is supported. It is similar to `git rebase -i`,
in that it opens the user's default editor and lets them pick which kernel
versions to remove.

## TODO

* Have an "automated" mode (without the interactive prompt)
* Way better kernel detection heuristic
* Less hardcoded paths
