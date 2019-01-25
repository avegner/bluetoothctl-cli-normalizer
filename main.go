package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

var (
	menus = map[string][]string{
		"main": {
			"list",
			"show",
			"select",
			"devices",
			"paired-devices",
			"system-alias",
			"reset-alias",
			"power",
			"pairable",
			"discoverable",
			"agent",
			"default-agent",
			"advertise",
			"set-alias",
			"scan",
			"info",
			"pair",
			"trust",
			"untrust",
			"block",
			"unblock",
			"remove",
			"connect",
			"disconnect",
		},
		"advertise": {
			"set-uuids",
			"set-service",
			"set-manufacturer",
			"set-tx-power",
			"set-name",
			"set-appearance",
			"set-duration",
			"set-timeout",
		},
		"scan": {
			"uuids",
			"rssi",
			"pathloss",
			"transport",
			"duplicate-data",
			"clear",
		},
		"gatt": {
			"list-attributes",
			"select-attribute",
			"attribute-info",
			"read",
			"write",
			"acquire-write",
			"release-write",
			"acquire-notify",
			"release-notify",
			"notify",
			"register-application",
			"unregister-application",
			"register-service",
			"unregister-service",
			"register-characteristic",
			"unregister-characteristic",
			"register-descriptor",
			"unregister-descriptor",
		},
	}
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cmd := getCmd("bluetoothctl")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe: %v", err)
	}

	// run bluetoothctl in a separate thread
	errc := make(chan error)
	go func() {
		errc <- runCmd(cmd)
	}()

	// TODO: pass tabs and arrows as well
	s := bufio.NewScanner(os.Stdin)
	menu := "main"
	for s.Scan() {
		l := s.Text()

		if isCmd(l, "quit") || isCmd(l, "exit") {
			break
		}

		// ignore submenu control commands
		if isCmd(l, "menu") || isCmd(l, "back") {
			_, _ = stdin.Write([]byte("\n"))
			continue
		}

		// print help for all menus
		if isCmd(l, "help") {
			menu = printHelp(stdin, menu)
			continue
		}

	loop:
		for m, cs := range menus {
			for _, c := range cs {
				if isCmd(l, c) {
					menu = jumpToSubMenu(stdin, menu, m)
					break loop
				}
			}
		}

		_, _ = stdin.Write([]byte(l + "\n"))
	}

	// close stdin to make bluetoothctl stop
	_ = stdin.Close()
	return <-errc
}

func getCmd(name string) *exec.Cmd {
	cmd := exec.Command(name, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}

func runCmd(cmd *exec.Cmd) error {
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run: %v", err)
	}
	return nil
}

func isCmd(l, c string) bool {
	return strings.Split(l, " ")[0] == c
}

func jumpToSubMenu(stdin io.WriteCloser, s, d string) string {
	if s != d {
		if s != "main" {
			// back to main menu from submenus
			_, _ = stdin.Write([]byte("back\n"))
			s = "main"
		}
		if d != "main" {
			// from main menu to submenus
			stdin.Write([]byte("menu " + d + "\n"))
		}
	}

	return d
}

// TODO: eliminate duplications
func printHelp(stdin io.WriteCloser, menu string) string {
	for m := range menus {
		menu = jumpToSubMenu(stdin, menu, m)
	}
	return menu
}
