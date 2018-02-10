package runner

import (
	"os"
	"os/exec"
	"os/signal"

	"github.com/go-errors/errors"
	"github.com/itchio/butler/runner/macutil"
)

type appRunner struct {
	params *RunnerParams
}

var _ Runner = (*appRunner)(nil)

func newAppRunner(params *RunnerParams) (Runner, error) {
	ar := &appRunner{
		params: params,
	}
	return ar, nil
}

func (ar *appRunner) Prepare() error {
	// nothing to prepare
	return nil
}

func (ar *appRunner) Run() error {
	params := ar.params
	consumer := params.Consumer

	var args = []string{
		"-W",
		params.FullTargetPath,
		"--args",
	}
	args = append(args, params.Args...)

	consumer.Infof("App bundle is (%s)", params.FullTargetPath)

	binaryPath, err := macutil.GetExecutablePath(params.FullTargetPath)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	consumer.Infof("Actual binary is (%s)", binaryPath)

	cmd := exec.CommandContext(params.Ctx, "open", args...)
	// I doubt this matters
	cmd.Dir = params.Dir
	cmd.Env = params.Env
	// 'open' does not relay stdout or stderr, so we don't
	// even bother setting them

	go func() {
		// catch SIGINT and kill the child if needed
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		consumer.Infof("Signal handler installed...")

		// Block until a signal is received.
		s := <-c
		consumer.Warnf("Got signal:", s)

		consumer.Warnf("Killing app...")
		// TODO: kill the actual binary, not the app
		cmd := exec.Command("pkill", "-f", binaryPath)
		err := cmd.Run()
		if err != nil {
			consumer.Errorf("While killing: ", err.Error())
		}
		os.Exit(0)
	}()

	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, 0)
	}

	return nil
}
