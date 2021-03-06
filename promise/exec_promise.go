package promise

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type ExecType int

const (
	ExecChange ExecType = iota
	ExecTest
)

func (t ExecType) Name() string {
	switch t {
	case ExecChange:
		return "change"
	case ExecTest:
		return "test"
	default:
		return "unknown"
	}
}

func (t ExecType) ReportResult(logger *Logger, result bool) {
	if t == ExecChange {
		logger.Changes++
	}

	if t == ExecTest {
		logger.Tests++
	}
}

func (t ExecType) String() string {
	return t.Name()
}

type ExecPromise struct {
	Type      ExecType
	Arguments []Argument
}

func (p ExecPromise) New(children []Promise, args []Argument) (Promise, error) {
	return ExecPromise{Type: p.Type, Arguments: args}, nil
}

func (p ExecPromise) getCommand(arguments []Constant, ctx *Context) (*exec.Cmd, error) {
	cmd := p.Arguments[0].GetValue(arguments, &ctx.Vars)
	largs := p.Arguments[1:]

	args := []string{}
	for _, argument := range largs {
		args = append(args, argument.GetValue(arguments, &ctx.Vars))
	}

	command := exec.Command(cmd, args...)

	if ctx.InDir != "" {
		fs, err := os.Stat(ctx.InDir)
		if err != nil {
			return nil, fmt.Errorf("(indir) error: %s", err.Error())
		}
		if !fs.IsDir() {
			return nil, fmt.Errorf("(indir) not a directory: %s", ctx.InDir)
		}
		command.Dir = ctx.InDir
	} else {
		command.Dir = os.Getenv("PWD")
	}

	command.Env = os.Environ()
	for _, v := range ctx.Env {
		command.Env = append(command.Env, v)
	}

	return command, nil
}

func (p ExecPromise) Desc(arguments []Constant) string {
	if len(p.Arguments) == 0 {
		return "(" + p.Type.Name() + ")"
	}

	cmd := p.Arguments[0].GetValue(arguments, &Variables{})
	largs := p.Arguments[1:]

	args := make([]string, len(largs))
	for i, v := range largs {
		args[i] = v.GetValue(arguments, &Variables{})
	}

	return "(" + p.Type.Name() + " <" + cmd + " [" + strings.Join(args, ", ") + "] >)"
}

func (p ExecPromise) Eval(arguments []Constant, ctx *Context) bool {
	command, err := p.getCommand(arguments, ctx)
	if err != nil {
		ctx.Logger.Stderr.Write([]byte(err.Error()))
		return false
	}
	command.Stdout = ctx.Logger.Stdout
	command.Stderr = ctx.Logger.Stderr

	ctx.Logger.Info.Write([]byte(strings.Join(command.Args, " ") + "\n"))

	err = command.Run()

	result := (err == nil)
	p.Type.ReportResult(&ctx.Logger, result)
	return result
}

/////////////////////////////

type PipePromise struct {
	Execs []ExecPromise
}

func (p PipePromise) New(children []Promise, args []Argument) (Promise, error) {

	execs := []ExecPromise{}

	for _, c := range children {
		switch t := c.(type) {
		case ExecPromise:
			execs = append(execs, t)
		default:
			return nil, fmt.Errorf("only (test) or (change) promises allowed inside (pipe) promise")
		}
	}

	return PipePromise{execs}, nil
}

func (p PipePromise) Desc(arguments []Constant) string {
	retval := "(pipe"
	for _, v := range p.Execs {
		retval += " " + v.Desc(arguments)
	}
	return retval + ")"
}

func (p PipePromise) Eval(arguments []Constant, ctx *Context) bool {
	commands := []*exec.Cmd{}
	cstrings := []string{}

	for _, v := range p.Execs {
		cmd, err := v.getCommand(arguments, ctx)
		if err != nil {
			ctx.Logger.Stderr.Write([]byte(err.Error()))
			return false
		}

		cstrings = append(cstrings, strings.Join(cmd.Args, " "))
		commands = append(commands, cmd)
	}

	for i, command := range commands[:len(commands)-1] {
		out, err := command.StdoutPipe()
		if err != nil {
			ctx.Logger.Stderr.Write([]byte(err.Error()))
			return false
		}
		command.Start()
		commands[i+1].Stdin = out
	}

	ctx.Logger.Info.Write([]byte(strings.Join(cstrings, " | ") + "\n"))

	last_cmd := commands[len(commands)-1]
	last_cmd.Stdout = ctx.Logger.Stdout
	last_cmd.Stderr = ctx.Logger.Stderr

	err := last_cmd.Run()

	for _, command := range commands[:len(commands)-1] {
		command.Wait()
	}

	if err != nil {
		return false
	} else {
		return true
	}
}
