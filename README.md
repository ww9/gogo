[![License: Unlicense](https://img.shields.io/badge/license-Unlicense-blue.svg)](http://unlicense.org/) [![Go Report Card](https://goreportcard.com/badge/github.com/ww9/gogo)](https://goreportcard.com/report/github.com/ww9/gogo)

# gogo 🏃

`gogo` compiles and restarts Go applications when code changes.

## Usage

`go install -i github.com/ww9/gogo`

`gogo main.go`

## Options

```txt
  -all				reloads whenever any file changes instead of only .go files
  -bin string			name of generated binary file (default ".gogo")
  -buildargs string		additional go build arguments
  -watchdir string		path to monitor for file changes (default ".")
  -builddir string		path to build files from (defaults to -watchdir)
  -excludedir value		directories to skip monitoring
  -godep			use godep when building
  -logprefix string		log prefix (default "gogo")
  -runargs string		arguments passed when running the program
```

## Features & Todo

- [x] Watches a directory and its subdirectories for .go file changes
- [x] Recompiles and restarts the Go app when file changes are detected
- [X] Supports listening to all file changes rather than only .go files using -all
- [x] Tested on Windows 10
- [x] Prevents annoying Windows [firewall prompt](annoying_windows_network_prompt.png) that happens when using `go run` since it always compile to the same binary file name
- [ ] Option to delete compiled file after exiting `gogo`. Perhaps enabled by default even
- [ ] Add optional debounce/delay for when multiple files change simultaneously (git branch change and other tooling)
- [ ] Remove godep support (thanks for everything godep team ❤)
- [ ] Add go.mod file
- [ ] Add --files argument to allow filtering files being watched using [glob](https://en.wikipedia.org/wiki/Glob_(programming)) matching
- [ ] Write tests including real file system usage
- [ ] Test on popular Linux distros, BSD and OSX

## Rather not do

- [ ] Config file: Passing arguments gets old but running `gogo` just works for most projects. Also Makefiles and shellscripts can fullfill this role.

## Changelog

* Removed dependency of [github.com/urfave/cli](https://github.com/urfave/cli) in favor of `flag` from standard library

* Deleted Builder and Runner interfaces which were implemented by only one struct each

* Added `-runargs` cli argument to allow passing arguments when running the program

## Credits

This is a heavily modified fork of [acoshift/goreload](https://github.com/acoshift/goreload) which itself is a fork of [codegangsta/gin](https://github.com/codegangsta/gin).

## License

[The Unlicense](http://unlicense.org/), [Public Domain](https://gist.github.com/ww9/4c4481fb7b55186960a34266078c88b1). As free as it gets.