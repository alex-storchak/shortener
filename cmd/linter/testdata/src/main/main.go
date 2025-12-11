package main

import (
	"log"
	"os"
)

func main() {
	log.Fatal("fatal from main")
	os.Exit(1)

	log.Println("test")
	panic("test") // want "panic function used in expression"
}

func f1() {
	log.Fatal("bad") // want "log.Fatal used outside main function of main package"
}

func f2() {
	os.Exit(1) // want "os.Exit used outside main function of main package"
}

type S struct{}

func (S) method() {
	log.Fatal("in method") // want "log.Fatal used outside main function of main package"
}

func helper() {
	log.Println("test")
}
