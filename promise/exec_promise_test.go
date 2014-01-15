package promise

import (
	"testing"
	"bytes"
	"strconv"
)

func TestExecPromise(t *testing.T) {
	var promise Promise
	promise = ExecPromise{ExecTest,[]Argument{
		Constant{"/bin/echo"},
		Constant{"Hello"},
		ArgGetter{0}}}

	var sout,serr bytes.Buffer
	ctx := Context{}
	ctx.Logger = Logger{Stdout: &sout, Stderr: &serr, Info: &sout}

	res := promise.Eval([]Constant{},&ctx)
	equals(t, true, res)
	equals(t, strconv.Itoa(24), strconv.Itoa(len(sout.String())))
	equals(t, "/bin/echo Hello \nHello \n", sout.String())
	equals(t, strconv.Itoa(0), strconv.Itoa(len(serr.String())))
}

func TestPipePromise(t *testing.T) {
	exec1 := ExecPromise{ExecTest, []Argument{Constant{"/bin/echo"},
		Constant{"hello world"}}}
	exec2 := ExecPromise{ExecChange, []Argument{Constant{"/usr/bin/rev"}}}

	var promise Promise

	promises := []ExecPromise{exec1, exec2}
	promise = PipePromise{promises}

	var sout,serr bytes.Buffer
	ctx := Context{}
	ctx.Logger = Logger{Stdout: &sout, Stderr: &serr, Info: &sout}

	res := promise.Eval([]Constant{}, &ctx)
	equals(t, true, res)
	equals(t, strconv.Itoa(49), strconv.Itoa(len(sout.String())))
	equals(t, "/bin/echo hello world | /usr/bin/rev\ndlrow olleh\n", sout.String())
	equals(t, strconv.Itoa(0), strconv.Itoa(len(serr.String())))
}

func TestExecReporting(t *testing.T) {
	arguments := []Argument {
		Constant{"/bin/echo"},
		Constant{"Hello"},
		ArgGetter{0},
	}

	var tests = []struct {
		promise ExecPromise
		changes int
	}{
		{ExecPromise{ExecChange, arguments},1},
		{ExecPromise{ExecTest, arguments},0},
	}

	for _,test := range tests {
		var sout,serr bytes.Buffer
		ctx := Context{}
		ctx.Logger = Logger{Stdout: &sout, Stderr: &serr, Info: &sout}

		res := test.promise.Eval([]Constant{},&ctx)
		equals(t, true, res)
		equals(t, strconv.Itoa(test.changes), strconv.Itoa(ctx.Logger.Changes))
	}
}
