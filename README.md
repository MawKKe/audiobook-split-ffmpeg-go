# audiobook-split-ffmpeg-go

Split audiobook file into per-chapter files using chapter metadata and ffmpeg.

Useful in situations where your preferred audio player does not support chapter metadata.

**NOTE**: Works only if the input file actually has chapter metadata (see example below)

**NOTE**: Feature-wise this program/library is identical to
https://github.com/MawKKe/audiobook-split-ffmpeg except this one is written in
Go instead of Python. "Why???" you might ask? Well, I was learning Go
and need a project...

**NOTE**: this is a quick rewrite and test coverage is not that great; there might be bugs
not present in the Python version.

# Example

Let's say, you have an audio file `mybook.m4b`, for which `ffprobe -i mybook.m4b`
shows the following:

    Chapter #0:0: start 0.000000, end 1079.000000
    Metadata:
      title           : Chapter Zero
    Chapter #0:1: start 1079.000000, end 2040.000000
    Metadata:
      title           : Chapter One
    Chapter #0:2: start 2040.000000, end 2878.000000
    Metadata:
      title           : Chapter Two
    Chapter #0:3: start 2878.000000, end 3506.000000

Then, running:

    $ audiobook-split-ffmpeg-go --infile mybook.m4b --outdir /tmp/foo

..produces the following files:
- `/tmp/foo/001 - Chapter Zero.m4b`
- `/tmp/foo/002 - Chapter One.m4b`
- `/tmp/foo/003 - Chapter Two.m4b`

You may then play these files with your preferred application.

# Install

To install the main executable:

    $ go install github.com/MawKKe/audiobook-split-ffmpeg-go/cmd/audiobook-split-ffmpeg-go@latest

This should place the executable into your user's `$GOPATH/bin/`. If that path is in your `$PATH`,
you are good to go. Next, see `Usage` below.

However, if you want to use the library in your projects, run:

    $ go get github.com/MawKKe/audiobook-split-ffmpeg-go

See the file `cmd/audiobook-split-ffmpeg-go/main.go` for hints how to use the library.

# Usage

See the help:

    $ audiobook-split-ffmpeg-go -h

In the simplest case you can just call

    $ audiobook-split-ffmpeg-go --infile /path/to/audio.m4b --outdir foo

Note that this script will never overwrite files in `foo/`, so you must delete conflicting
files manually (or specify some other empty/nonexistent directory)

The chapter titles will be included in the filenames if they are available in
the chapter metadata. You may prevent this behaviour with flag `--no-use-title-as-filename`,
in which case the filenames will include the input file basename instead (this
is useful is your metadata is crappy or malformed, for example).

You may specify how many parallel `ffmpeg` jobs you want with command line param `--concurrency`.
The default concurrency is equal to the number of cores available. Note that at some point increasing
the concurrency might not increase the throughput. (We specifically instruct `ffmpeg` to NOT perform
re-encoding, so most of the processing work consists of copying the existing encoded audio data from the
input file to the output file(s) - this kind of processing is more I/O bounded than CPU-bounded).

# Dependencies
The project was developed with Go version 1.18, but it *should* compile with earlier versions.
You might be able to compile the project with earlier releases by adjusting the version in file `go.mod`.

This application has no 3rd party library dependencies, as everything is
implemented using the Go standard library. However, the script assumes
the that the following system-executables are available somewhere in your `$PATH`:

- `ffmpeg`
- `ffprobe`

For Ubuntu, these can be installed with `apt install ffmpeg`.

# Development and Testing

The Go tooling handles dependencies via 'go get', although this project
requires no external Go libraries at the moment.

To start working on the code, clone the repo:

    $ git clone https://github.com/MawKKe/audiobook-split-ffmpeg-go && audiobook-split-ffmpeg

To build the main binary:

    $ make go

or manually:

    $ go build cmd/audiobook-split-ffmpeg-go

To run tests:

    $ go test

(TODO: test coverage needs some improvement)

The provided `Makefile` has some useful targets defined for easier development and testing.
You should check it out.

# Features

- The script does not transcode/re-encode the audio data. This speeds up the processing, but has
  the possibility of creating mangled audio in some rare cases (let me know if this happens).

- This script will instruct ffmpeg to write metadata in the resulting chapter files, including:
  - `track`: the chapter number; in the format X/Y, where X = chapter number, Y = total num of chapters.
  - `title`: the chapter title, as-is (if available)

- The chapter numbers are included in the output file names, padded with zeroes so that all
  numbers are of equal length. This makes the files much easier to sort by name.

- The work is parallelized to speed up the processing.

# License

Copyright 2022 Markus Holmström (MawKKe)

The works under this repository are licenced under Apache License 2.0.
See file `LICENSE` for more information.

# Contributing

This project is hosted at https://github.com/MawKKe/audiobook-split-ffmpeg-go

You are welcome to leave bug reports, fixes and feature requests. Thanks!


