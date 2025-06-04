package main

import "os"

func main() {
	func() {
		os.Exit(1) // want "direct call to os.Exit is prohibited in the main function of the main package"
	}()
}
