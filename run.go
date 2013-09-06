package main

import (
	"io"
	"os"
	"log"
	"fmt"
	"bufio"
	
	llconf_io "github.com/mruediger/llconf/io"
	"github.com/mruediger/llconf/parse"
	"github.com/mruediger/llconf/promise"
)

var run = &Command{
	Name: "run",
	Usage: "run   [arguments...] [folder]",
	Run: execRun,
}

var run_cfg struct{
	input string
	promise string
	verbose bool
	dryrun bool
}

func init() {
	run.Flag.BoolVar(&run_cfg.verbose, "verbose", false, "enable verbose output")
	run.Flag.BoolVar(&run_cfg.dryrun, "dry-run", false, "just parse the promise and not check it")
	run.Flag.StringVar(&run_cfg.promise, "promise", "done", "the promise that will be used as root")
}

func execRun(args []string, logi, loge *log.Logger) {
	switch len(args) {
	case 0:
		run_cfg.input = ""
		fmt.Println("no input folder specified, reading from stdin")
	case 1:
		run_cfg.input = args[0]
	default:
		fmt.Fprintf(os.Stderr, "argument count mismatch")
		os.Exit(1)
	}
	
	input,err := openInput(run_cfg.input)
	if err != nil {
		loge.Printf("could not open %q: %v\n", run_cfg.input, err)
		return
	}

	globals := map[string]string{}
	promises,err := parse.ParsePromises(input,&globals)
	if err != nil {
		loge.Printf("error while parsing input: %v\n", err)
		return
	}

	p,promise_present := promises[run_cfg.promise]
	if !promise_present {
		loge.Printf("specified goal (%s) not found in config\n", run_cfg.promise)
	}

	if run_cfg.dryrun {
		return
	}

	logger := promise.Logger{LogWriter{logi}, LogWriter{loge}}
	
	success := p.Eval([]promise.Constant{}, &logger)
	if success {
		logi.Println("evaluation successful\n")
	} else {
		loge.Println("error during evaluation")		
	}
}

func openInput( source string ) (io.RuneReader, error) {
	if source == "" {
		input := bufio.NewReader( os.Stdin )
		return input,nil
	} else {
		input,err := llconf_io.NewFolderRuneReader( run_cfg.input )
		return &input,err
	}
}
