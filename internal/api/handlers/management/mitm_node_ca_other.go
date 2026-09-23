//go:build !windows

package management

func setMITMNodeCA(string) error   { return nil }
func clearMITMNodeCA(string) error { return nil }
