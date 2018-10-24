# GoGo

`gogo` compiles and restarts Go applications when code changes.

## Usage

`go install -i github.com/ww9/gogo`

`gogo main.go`

## Options

```txt
  -all					reloads whenever any file changes instead of only .go files
  -bin string			name of generated binary file (default ".gogo")
  -buildargs string		additional go build arguments
  -watchdir string		path to monitor for file changes (default ".")
  -builddir string		path to build files from (defaults to -watchdir)
  -excludedirs value	relative directories to skip monitoring for file changes. multiple paths can be specified by repeating the -excludedirs flag
  -godep				use godep when building
  -logprefix string		log prefix (default "gogo")
  -runargs string		arguments passed when running the program
```

# Changelog

* Removed dependency of [github.com/urfave/cli](https://github.com/urfave/cli) in favor of `flag` from standard library

* Removed Builder and Runner interfaces which were implemented by only one struct each

* Added `-runargs` cli argument to allow passing arguments when running the program

* Renamed from goreload to gogo

## Credits

This is heavily modified fork from [acoshift/goreload](https://github.com/acoshift/goreload) which itself is a fork from [codegangsta/gin](https://github.com/codegangsta/gin). All credits go the their contributors.