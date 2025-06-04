package main

import "os"

func notMain() {
	os.Exit(1) // No error, as this is not the main function
}
