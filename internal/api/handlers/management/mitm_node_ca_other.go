//go:build !windows

package management

import "os/exec"

func hiddenMITMCommand(name string, args ...string) *exec.Cmd { return exec.Command(name, args...) }

func setMITMNodeCA(string) error   { return nil }
func clearMITMNodeCA(string) error { return nil }
