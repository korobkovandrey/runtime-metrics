package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMainFunction(t *testing.T) {
	main()
	_, err := os.Stat("version.go")
	require.NoError(t, err)

	content, err := os.ReadFile("version.go")
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "buildVersion")
	assert.Contains(t, contentStr, "buildDate")
	assert.Contains(t, contentStr, "buildCommit")
	assert.Contains(t, contentStr, "DO NOT EDIT")
	info, err := os.Stat("version.go")
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
	require.NoError(t, os.Remove("version.go"))
}

func TestRunCommand(t *testing.T) {
	output, err := runCommand("echo", "test")
	require.NoError(t, err, "Expected no error")
	assert.Equal(t, "test\n", output, "Expected output 'test'")
	_, err = runCommand("nonexistentcommand")
	assert.Error(t, err, "Expected error for nonexistent command")
}
