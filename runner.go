package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"
)

type Runner struct {
	bin       string
	args      []string
	writer    io.Writer
	command   *exec.Cmd
	starttime time.Time
}

func NewRunner(bin string, args ...string) *Runner {
	return &Runner{
		bin:       bin,
		args:      args,
		writer:    ioutil.Discard,
		starttime: time.Now(),
	}
}

func (r *Runner) Run() (*exec.Cmd, error) {
	if r.needsRefresh() {
		r.Kill()
	}

	if r.command == nil || r.Exited() {
		err := r.runBin()
		if err != nil {
			log.Print("Error running: ", err)
		}
		time.Sleep(250 * time.Millisecond)
		return r.command, err
	}
	return r.command, nil
}

func (r *Runner) Info() (os.FileInfo, error) {
	return os.Stat(r.bin)
}

func (r *Runner) SetWriter(writer io.Writer) {
	r.writer = writer
}

func (r *Runner) Kill() error {
	if r.command != nil && r.command.Process != nil {
		done := make(chan error)
		go func() {
			r.command.Wait()
			close(done)
		}()

		//Trying a "soft" kill first
		if runtime.GOOS == "windows" {
			if err := r.command.Process.Kill(); err != nil {
				return err
			}
		} else if err := r.command.Process.Signal(os.Interrupt); err != nil {
			return err
		}

		//Wait for our process to die before we return or hard kill after 3 sec
		select {
		case <-time.After(3 * time.Second):
			if err := r.command.Process.Kill(); err != nil {
				log.Println("failed to kill: ", err)
			}
		case <-done:
		}
		r.command = nil
	}

	return nil
}

func (r *Runner) Exited() bool {
	return r.command != nil && r.command.ProcessState != nil && r.command.ProcessState.Exited()
}

func (r *Runner) runBin() error {
	r.command = exec.Command(r.bin, r.args...)
	stdout, err := r.command.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := r.command.StderrPipe()
	if err != nil {
		return err
	}

	err = r.command.Start()
	if err != nil {
		return err
	}

	r.starttime = time.Now()

	go io.Copy(r.writer, stdout)
	go io.Copy(r.writer, stderr)
	go r.command.Wait()

	return nil
}

func (r *Runner) needsRefresh() bool {
	info, err := r.Info()
	if err != nil {
		return false
	}
	return info.ModTime().After(r.starttime)
}
