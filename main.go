package main

import (
	"errors"
	"fmt"

	"github.com/acoshift/goreload/lib"
	shellwords "github.com/mattn/go-shellwords"
	"gopkg.in/urfave/cli.v1"

	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/0xAX/notificator"
)

var (
	startTime     = time.Now()
	logger        = log.New(os.Stdout, "[goreload] ", 0)
	buildError    error
	colorGreen    = string([]byte{27, 91, 57, 55, 59, 51, 50, 59, 49, 109})
	colorRed      = string([]byte{27, 91, 57, 55, 59, 51, 49, 59, 49, 109})
	colorReset    = string([]byte{27, 91, 48, 109})
	notifier      = notificator.New(notificator.Options{AppName: "Goreload Build"})
	notifications = false
)

func main() {
	app := cli.NewApp()
	app.Name = "goreload"
	app.Usage = "A live reload utility for Go web applications."
	app.Action = MainAction
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:   "port,p",
			Value:  3000,
			EnvVar: "GIN_PORT",
			Usage:  "port for the proxy server",
		},
		cli.IntFlag{
			Name:   "appPort,a",
			Value:  3001,
			EnvVar: "BIN_APP_PORT",
			Usage:  "port for the Go web server",
		},
		cli.StringFlag{
			Name:   "bin,b",
			Value:  "gin-bin",
			EnvVar: "GIN_BIN",
			Usage:  "name of generated binary file",
		},
		cli.StringFlag{
			Name:   "path,t",
			Value:  ".",
			EnvVar: "GIN_PATH",
			Usage:  "Path to watch files from",
		},
		cli.StringFlag{
			Name:   "build,d",
			Value:  "",
			EnvVar: "GIN_BUILD",
			Usage:  "Path to build files from (defaults to same value as --path)",
		},
		cli.StringSliceFlag{
			Name:   "excludeDir,x",
			Value:  &cli.StringSlice{},
			EnvVar: "GIN_EXCLUDE_DIR",
			Usage:  "Relative directories to exclude",
		},
		cli.BoolFlag{
			Name:   "all",
			EnvVar: "GIN_ALL",
			Usage:  "reloads whenever any file changes, as opposed to reloading only on .go file change",
		},
		cli.BoolFlag{
			Name:   "godep,g",
			EnvVar: "GIN_GODEP",
			Usage:  "use godep when building",
		},
		cli.StringFlag{
			Name:   "buildArgs",
			EnvVar: "GIN_BUILD_ARGS",
			Usage:  "Additional go build arguments",
		},
		cli.StringFlag{
			Name:   "logPrefix",
			EnvVar: "GIN_LOG_PREFIX",
			Usage:  "Log prefix",
			Value:  "gin",
		},
		cli.BoolFlag{
			Name:   "notifications",
			EnvVar: "GIN_NOTIFICATIONS",
			Usage:  "Enables desktop notifications",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:      "run",
			ShortName: "r",
			Usage:     "Run the gin proxy in the current working directory",
			Action:    MainAction,
		},
	}

	app.Run(os.Args)
}

func MainAction(c *cli.Context) {
	port := c.GlobalInt("port")
	all := c.GlobalBool("all")
	appPort := strconv.Itoa(c.GlobalInt("appPort"))
	keyFile := c.GlobalString("keyFile")
	certFile := c.GlobalString("certFile")
	logPrefix := c.GlobalString("logPrefix")
	notifications = c.GlobalBool("notifications")

	logger.SetPrefix(fmt.Sprintf("[%s] ", logPrefix))

	// Set the PORT env
	os.Setenv("PORT", appPort)

	wd, err := os.Getwd()
	if err != nil {
		logger.Fatal(err)
	}

	buildArgs, err := shellwords.Parse(c.GlobalString("buildArgs"))
	if err != nil {
		logger.Fatal(err)
	}

	buildPath := c.GlobalString("build")
	if buildPath == "" {
		buildPath = c.GlobalString("path")
	}
	builder := gin.NewBuilder(buildPath, c.GlobalString("bin"), c.GlobalBool("godep"), wd, buildArgs)
	runner := gin.NewRunner(filepath.Join(wd, builder.Binary()), c.Args()...)
	runner.SetWriter(os.Stdout)
	proxy := gin.NewProxy(builder, runner)

	config := &gin.Config{
		Port:     port,
		ProxyTo:  "http://localhost:" + appPort,
		KeyFile:  keyFile,
		CertFile: certFile,
	}

	err = proxy.Run(config)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Printf("Listening on port %d\n", port)

	shutdown(runner)

	// build right now
	build(builder, runner, logger)

	// scan for changes
	scanChanges(c.GlobalString("path"), c.GlobalStringSlice("excludeDir"), all, func(path string) {
		runner.Kill()
		build(builder, runner, logger)
	})
}

func build(builder gin.Builder, runner gin.Runner, logger *log.Logger) {
	logger.Println("Building...")

	if notifications {
		notifier.Push("Build Started!", "Building "+builder.Binary()+"...", "", notificator.UR_NORMAL)
	}
	err := builder.Build()
	if err != nil {
		buildError = err
		logger.Printf("%sBuild failed%s\n", colorRed, colorReset)
		fmt.Println(builder.Errors())
		buildErrors := strings.Split(builder.Errors(), "\n")
		if notifications {
			if err := notifier.Push("Build FAILED!", buildErrors[1], "", notificator.UR_CRITICAL); err != nil {
				logger.Println("Notification send failed")
			}
		}
	} else {
		buildError = nil
		logger.Printf("%sBuild finished%s\n", colorGreen, colorReset)
		runner.Run()

		if notifications {
			if err := notifier.Push("Build Succeded", "Build Finished!", "", notificator.UR_CRITICAL); err != nil {
				logger.Println("Notification send failed")
			}
		}
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

func shutdown(runner gin.Runner) {
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
