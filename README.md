# gini

gini is a utility to make some basic operations on ini files.

```
Tool to get/set key from an ini file.

Usage:
  gini [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  get         retrieve a key from an ini file
  help        Help about any command
  set         add/update key/value
  version     print version of gini

Flags:
  -h, --help       help for gini
      --i string   init file to read/update
      --k string   key to read or update
      --s string   section of ini file (can be empty)

Use "gini [command] --help" for more information about a command.
```

The **set** command have two more options:

```
gini set -h
add/update key/value in the desired section (can be empty)

Usage:
  gini set [flags]

Flags:
      --c          create file if no present
  -h, --help       help for set
      --v string   value to set

Global Flags:
      --i string   init file to read/update
      --k string   key to read or update
      --s string   section of ini file (can be empty)
```

# Some examples 

```
$ gini get --k key --i tests/test.ini
value
$ gini get --k key2 --i tests/test.ini --s section
value2
$ gini get --k keyThatDoNotExists --i tests/test.ini
$ echo $?
0
```

# Install

## Option 1

* Download the release
* Install the binary in /usr/local/bin 

## Option 2: with asdf

```
asdf plugin-add gini https://github.com/sgaunet/asdf-gini.git
asdf install gini latest
```

## Option 3: With brew

```
brew tap sgaunet/tools
brew install mdtohtml
```

# Development

This project is using :

* golang 1.19+
* [task for development](https://taskfile.dev/#/)
* docker
* [docker buildx](https://github.com/docker/buildx)
* docker manifest
* [goreleaser](https://goreleaser.com/)
* [Principle package of the CLI](https://pkg.go.dev/gopkg.in/ini.v1?utm_source=godoc)

The docker image is only created to simplify the copy of gini in another docker image.

# Tests

Tests are executed with [venom](https://github.com/ovh/venom).

```
task tests
```
