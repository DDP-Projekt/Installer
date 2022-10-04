package main

import "fmt"

func main() {
	InfoF("This is a simple info %s %d", "hi", 222)
	DoneF("I am done with this shit %s %d", "bye", 420)
	WarnF("I'm warning ya, I will %s %d", "do it", 23232)
	ErrorF("Oh fuck, oh God, oh Fuck %s %d", "sdf", 3434)
	InfoF("press ENTER to continue...")
	fmt.Scanln()
}
