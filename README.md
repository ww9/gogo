# GoGo

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

## Changelog

* Removed dependency of [github.com/urfave/cli](https://github.com/urfave/cli) in favor of `flag` from standard library

* Deleted Builder and Runner interfaces which were implemented by only one struct each

* Added `-runargs` cli argument to allow passing arguments when running the program

## Credits

This is a heavily modified fork of [acoshift/goreload](https://github.com/acoshift/goreload) which itself is a fork of [codegangsta/gin](https://github.com/codegangsta/gin).