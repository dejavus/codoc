# Codoc

Code and document viewer for Go.

## Install

`git clone https://github.com/dejavus/codoc.git`

## Install dependencies

`dep ensure`

## Run

`go run cmd/cli/main.go --package {target_dir} --exclude {exclude} --server`

## Parameters

- package: Path of package for parse.
- exclude: Relative paths that you want to exclude. Use comma as separator for multiple exclude paths.
- server: Run as server (default port is 8000).