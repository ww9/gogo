package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	shellwords "github.com/mattn/go-shellwords"

	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var (
	startTime  = time.Now()
	logger     = log.New(os.Stdout, "[gogo] ", 0)
	buildError error
)

// We use long flag names because all flags are passed when running the program and we don't want to conflict with sane flag names
var flagAllFiles = flag.Bool("all", false, "reloads whenever any file changes instead of only .go files")
var flagBinaryFileName = flag.String("bin", ".gogo", "name of generated binary file")
var flagWatchDir = flag.String("watchdir", ".", "path to monitor for file changes")
var flagBuildDir = flag.String("builddir", "", "path to build files from (defaults to -watchdir)")
var flagExcludeDirs StringListArg
var flagGoDep = flag.Bool("godep", false, "use godep when building")
var flagBuildArgs = flag.String("buildargs", "", "additional go build arguments")
var flagRunArgs = flag.String("runargs", "", "arguments passed when running the program")
var flagLogPrefix = flag.String("logprefix", "gogo", "log prefix")

func main() {
	flag.Var(&flagExcludeDirs, "excludedir", "relative directories to skip monitoring for file changes. multiple paths can be specified by repeating the -excludedir flag")
	flag.Parse()
	if flagExcludeDirs == nil {
		flagExcludeDirs = StringListArg{}
	}

	logger.SetPrefix(fmt.Sprintf("[%s] ", *flagLogPrefix))
	wd, err := os.Getwd()
	if err != nil {
		logger.Fatal(err)
	}
	buildArgs, err := shellwords.Parse(*flagBuildArgs)
	if err != nil {
		logger.Fatal(err)
	}
	runArgs, err := shellwords.Parse(*flagRunArgs)
	if err != nil {
		logger.Fatal(err)
	}
	if *flagBuildDir == "" {
		*flagBuildDir = *flagWatchDir
	}
	builder := NewBuilder(*flagBuildDir, *flagBinaryFileName, *flagGoDep, wd, buildArgs)
	runner := NewRunner(filepath.Join(wd, builder.Binary()), runArgs...)
	runner.SetWriter(os.Stdout)

	exitGracefully(runner)

	build(builder, runner, logger)

	scanChanges(*flagWatchDir, flagExcludeDirs, *flagAllFiles, func(path string) {
		runner.Kill()
		build(builder, runner, logger)
	})
}

func build(builder *Builder, runner *Runner, logger *log.Logger) {
	err := builder.Build()
	if err != nil {
		buildError = err
		logger.Printf("Build failed\n")
		fmt.Println(builder.Errors())
	} else {
		buildError = nil
		runner.Run()
	}

	time.Sleep(100 * time.Millisecond)
}

type scanCallback func(path string)

func scanChanges(watchPath string, excludeDirs []string, allFiles bool, cb scanCallback) {
	for {
		filepath.Walk(watchPath, func(path string, info os.FileInfo, err error) error {
			if path == ".git" && info.IsDir() {
				return filepath.SkipDir
			}
			for _, x := range excludeDirs {
				if x == path {
					return filepath.SkipDir
				}
			}

			// ignore hidden files
			if filepath.Base(path)[0] == '.' {
				return nil
			}

			if (allFiles || filepath.Ext(path) == ".go") && info.ModTime().After(startTime) {
				cb(path)
				startTime = time.Now()
				return errors.New("done")
			}

			return nil
		})
		time.Sleep(500 * time.Millisecond)
	}
}

// exitGracefully listens for exit signal (usually when user presses CTRL+C) and gracefully close the running program before exiting.
func exitGracefully(runner *Runner) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-c
		log.Println("Got signal: ", s)
		err := runner.Kill()
		if err != nil {
			log.Print("Error killing: ", err)
		}
		os.Exit(1)
	}()
}

// StringListArg is used so we can parse and accumulate multiple values of the same cli flag.
// For example -excludedir "dir1" -excludedir "dir2"
type StringListArg []string

func (arg *StringListArg) String() string {
	if arg == nil {
		return ""
	}
	return strings.Join(*arg, ";")
}
func (arg *StringListArg) Set(value string) error {
	*arg = append(*arg, value)
	return nil
}
