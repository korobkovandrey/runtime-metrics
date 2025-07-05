package main

import (
	"flag"
	"net"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunAndQuitMain(t *testing.T) {
	done := make(chan struct{})
	go func() {
		currentDir, err := os.Getwd()
		require.NoError(t, err)
		t.Chdir("../..")
		list, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		addr := list.Addr().String()
		list1, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		addr1 := list1.Addr().String()
		origArgs := os.Args
		defer func() { os.Args = origArgs }()
		os.Args = []string{"test", "-a", addr, "-g", addr1}
		flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
		require.NoError(t, list.Close())
		require.NoError(t, list1.Close())
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
