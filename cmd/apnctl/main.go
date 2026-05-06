package main

import (
	"flag"
	"fmt"
	"os"
)

const aospURL = "https://android.googlesource.com/device/sample/+/main/etc/apns-full-conf.xml?format=TEXT"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "apnctl:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		usage(os.Stderr)
		return flag.ErrHelp
	}

	switch args[0] {
	case "fetch":
		return runFetch(args[1:])
	case "convert":
		return runConvert(args[1:])
	case "find":
		return runFind(args[1:])
	case "list":
		return runList(args[1:])
	case "stats":
		return runStats(args[1:])
	case "validate":
		return runValidate(args[1:])
	case "patch":
		return runPatch(args[1:])
	case "inspect":
		return runInspect(args[1:])
	case "build":
		return runBuild(args[1:])
	case "help", "-h", "--help":
		usage(os.Stdout)
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}
