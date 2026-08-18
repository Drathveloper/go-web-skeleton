// Command scaffold generates the parts of this project that are the same every
// time. It ships inside the template and inside every project generated from
// it, so `scaffold module` keeps working after `scaffold new`.
package main

import (
	"errors"
	"fmt"
	"os"
)

var ErrUnknownCommand = errors.New("unknown command")

const usage = `scaffold — generate a project or a CRUD module

  scaffold new    --name <project> --module <path> --out <dir> [--roles a,b]
  scaffold module --context <ctx> --name <entity> [--plural <plural>]
                  [--field name:type[:required]]... [--roles role]...

Field types: string text int uint bool date datetime money email ref=<entity>
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "scaffold:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		_, _ = fmt.Fprint(os.Stderr, usage)
		return nil
	}
	switch args[0] {
	case "module":
		return runModule(args[1:])
	case "new":
		return runNew(args[1:])
	case "help", "-h", "--help":
		_, _ = fmt.Fprint(os.Stdout, usage)
		return nil
	default:
		return fmt.Errorf("%w: %q", ErrUnknownCommand, args[0])
	}
}
