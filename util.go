package ghost_postgres

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/jpillora/backoff"
)

func interruptThenKill(cmd *exec.Cmd, wait time.Duration) error {
	finished := make(chan error)
	go func(cmd *exec.Cmd, finished chan error) {
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			finished <- err
			return
		}
		if err := cmd.Wait(); err != nil {
			finished <- err
			return
		}
		finished <- nil
	}(cmd, finished)
	go func(duration time.Duration, finished chan error) {
		time.Sleep(duration)
		finished <- fmt.Errorf("Process didn't exit cleanly after %s", duration)
	}(wait, finished)
	err := <-finished
	if err != nil {
		// Ignore errors from these commands
		cmd.Process.Kill()
		cmd.Process.Release()
	}
	return err
}

func waitForService(db *sql.DB) error {
	// TODO configurable
	b := &backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    5 * time.Second,
		Factor: 2,
		Jitter: false,
	}
	var err error
	for err, dur := db.Ping(), b.Duration(); err != nil && dur < b.Max; err, dur = db.Ping(), b.Duration() {
		time.Sleep(dur)
	}
	return err
}
