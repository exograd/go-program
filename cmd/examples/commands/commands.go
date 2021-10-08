// Copyright (c) 2021 Nicolas Martyanoff <khaelin@gmail.com>
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY
// SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF OR
// IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

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

	p.AddCommand("foo", "foo command", cmdFoo)
	p.AddCommandFlag("foo", "d", "flag-d", "a command flag")
	p.AddCommandArgument("foo", "arg-1", "the first argument")
	p.AddCommandArgument("foo", "arg-2", "the second argument")
	p.AddCommandTrailingArgument("foo", "arg-3", "all trailing arguments")

	p.AddCommand("bar", "bar command", cmdBar)

	p.Start()
}

func cmdFoo(p *program.Program) {
	p.Info("running command foo")

	fmt.Printf("flag-a: %v\n", p.IsOptionSet("flag-a"))
	fmt.Printf("b: %v\n", p.IsOptionSet("b"))
	fmt.Printf("option-c: %s\n", p.OptionValue("option-c"))
	fmt.Printf("flag-d: %v\n", p.IsOptionSet("flag-d"))

	fmt.Printf("arg-1: %s\n", p.ArgumentValue("arg-1"))
	fmt.Printf("arg-2: %s\n", p.ArgumentValue("arg-2"))
	fmt.Printf("arg-3:")
	for _, value := range p.TrailingArgumentValues("arg-3") {
		fmt.Printf(" %s", value)
	}
	fmt.Printf("\n")
}

func cmdBar(p *program.Program) {
	p.Info("running command bar")

	fmt.Printf("flag-a: %v\n", p.IsOptionSet("flag-a"))
	fmt.Printf("b: %v\n", p.IsOptionSet("b"))
	fmt.Printf("option-c: %s\n", p.OptionValue("option-c"))
}
