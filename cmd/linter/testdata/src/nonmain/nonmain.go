package nonmain

import (
	"log"
	"os"
)

func f() {
	log.Fatal("cannot init") // want "log.Fatal used outside main function of main package"
	os.Exit(1)               // want "os.Exit used outside main function of main package"
}
