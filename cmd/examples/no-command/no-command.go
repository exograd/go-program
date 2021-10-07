package main

import (
	"fmt"

	"github.com/galdor/go-program"
)

func main() {
	p := program.NewProgram("no-command",
		"an example program without any command")

	p.AddGlobalFlag("", "flag-a", "a long flag")
	p.AddGlobalFlag("b", "", "a short flag")
	p.AddGlobalOption("c", "option-c", "value", "foo",
		"an option with both a short and long name")

	p.AddArgument("arg-1", "the first argument")
	p.AddArgument("arg-2", "the second argument")
	p.AddTrailingArgument("arg-3", "all trailing arguments")

	p.SetMain(main2)

	p.Start()
}

func main2(p *program.Program) {
	fmt.Printf("flag-a: %v\n", p.IsOptionSet("flag-a"))
	fmt.Printf("b: %v\n", p.IsOptionSet("b"))
	fmt.Printf("option-c: %s\n", p.OptionValue("option-c"))

	fmt.Printf("arg-1: %s\n", p.ArgumentValue("arg-1"))
	fmt.Printf("arg-2: %s\n", p.ArgumentValue("arg-2"))
	fmt.Printf("arg-3:")
	for _, value := range p.TrailingArgumentValues("arg-3") {
		fmt.Printf(" %s", value)
	}
	fmt.Printf("\n")
}
