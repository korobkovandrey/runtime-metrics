package main

import (
	"flag"
	"net"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var muxOsArgs sync.Mutex

func TestRunAndQuitMain(t *testing.T) {
	t.Run("run", func(t *testing.T) {
		runAndQuitMain(t, []string{"test"})
	})
	t.Run("pprof", func(t *testing.T) {
		list, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		require.NoError(t, list.Close())
		addr := list.Addr().String()
		runAndQuitMain(t, []string{"test", "-pprof", addr})
	})
	t.Run("grpc", func(t *testing.T) {
		runAndQuitMain(t, []string{"test", "-grpc"})
	})
}

func runAndQuitMain(t *testing.T, args []string) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		muxOsArgs.Lock()
		origArgs := os.Args
		os.Args = args
		muxOsArgs.Unlock()
		defer func() {
			muxOsArgs.Lock()
			os.Args = origArgs
			muxOsArgs.Unlock()
		}()
		flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
		currentDir, err := os.Getwd()
		require.NoError(t, err)
		t.Chdir("../..")
		main()
		t.Chdir(currentDir)
		done <- struct{}{}
	}()
	time.Sleep(time.Second)
	require.NoError(t, syscall.Kill(syscall.Getpid(), syscall.SIGTERM))
	select {
	case <-time.After(5 * time.Second):
		assert.Fail(t, "timeout")
	case <-done:
	}
}
