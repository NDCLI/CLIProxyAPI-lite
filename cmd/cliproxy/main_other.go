//go:build !windows

package main

import "fmt"

func main() {
	fmt.Println("The Lumina tray launcher is available only on Windows release packages.")
}
