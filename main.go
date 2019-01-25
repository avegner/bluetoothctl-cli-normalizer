package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type command struct {
	name, params, desc string
}

var (
	menus = map[string][]command{
		"main": {
			{"list", "", "List available controllers"},
			{"show", "[ctrl]", "Controller information"},
			{"select", "<ctrl>", "Select default controller"},
			{"devices", "", "List available devices"},
			{"paired-devices", "", "List paired devices"},
			{"system-alias", "<name>", "Set controller alias"},
			{"reset-alias", "", "Reset controller alias"},
			{"power", "<on/off>", "Set controller power"},
			{"pairable", "<on/off>", "Set controller pairable mode"},
			{"discoverable", "<on/off>", "Set controller discoverable mode"},
			{"agent", "<on/off/capability>", "Enable/disable agent with given capability"},
			{"default-agent", "", "Set agent as the default one"},
			{"advertise", "<on/off/type>", "Enable/disable advertising with given type"},
			{"set-alias", "<alias>", "Set device alias"},
			{"scan", "<on/off>", "Scan for devices"},
			{"info", "[dev]", "Device information"},
			{"pair", "[dev]", "Pair with device"},
			{"trust", "[dev]", "Trust device"},
			{"untrust", "[dev]", "Untrust device"},
			{"block", "[dev]", "Block device"},
			{"unblock", "[dev]", "Unblock device"},
			{"remove", "<dev>", "Remove device"},
			{"connect", "<dev>", "Connect device"},
			{"disconnect", "[dev]", "Disconnect device"},
		},
		"advertise": {
			{"set-uuids", "[uuid1 uuid2 ...]", "Set advertise uuids"},
			{"set-service", "[uuid] [data=xx xx ...]", "Set advertise service data"},
			{"set-manufacturer", "[id]", "Set advertise manufacturer data"},
			{"set-tx-power", "<on/off>", "Enable/disable TX power to be advertised"},
			{"set-name", "<on/off/name>", "Enable/disable local name to be advertised"},
			{"set-appearance", "<value>", "Set custom appearance to be advertised"},
			{"set-duration", "<seconds>", "Set advertise duration"},
			{"set-timeout", "<seconds>", "Set advertise timeout"},
		},
		"scan": {
			{"uuids", "[all/uuid1 uuid2 ...]", "Set/Get UUIDs filter"},
			{"rssi", "[rssi]", "Set/Get RSSI filter, and clears pathloss"},
			{"pathloss", "[pathloss]", "Set/Get Pathloss filter, and clears RSSI"},
			{"transport", "[transport]", "Set/Get transport filter"},
			{"duplicate-data", "[on/off]", "Set/Get duplicate data filter"},
			{"clear", "[uuids/rssi/pathloss/transport/duplicate-data]", "Clears discovery filter"},
		},
		"gatt": {
			{"list-attributes", "[dev]", "List attributes"},
			{"select-attribute", "<attribute/UUID>", "Select attribute"},
			{"attribute-info", "[attribute/UUID]", "Attribute info"},
			{"read", "", "Read attribute value"},
			{"write", "<data=xx xx ...>", "Write attribute value"},
			{"acquire-write", "", "Acquire Write file descriptor"},
			{"release-write", "", "Release Write file descriptor"},
			{"acquire-notify", "", "Acquire Notify file descriptor"},
			{"release-notify", "", "Release Notify file descriptor"},
			{"notify", "<on/off>", "Notify attribute value"},
			{"register-application", "[UUID ...]", "Register profile to connect"},
			{"unregister-application", "", "Unregister profile"},
			{"register-service", "<UUID>", "Register application service"},
			{"unregister-service", "<UUID/object>", "Unregister application service"},
			{"register-characteristic", "<UUID> <Flags=read,write,notify...>", "Register application characteristic"},
			{"unregister-characteristic", "<UUID/object>", "Unregister application characteristic"},
			{"register-descriptor", "<UUID> <Flags=read,write...>", "Register application descriptor"},
			{"unregister-descriptor", "<UUID/object>", "Unregister application descriptor"},
		},
	}

	commonCmds = []command{
		{"version", "", "Print version"},
		{"exit", "", "Exit"},
		{"quit", "", "Quit"},
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
	// TODO: track whether bluetoothctl is still alive
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
			printHelp()
			_, _ = stdin.Write([]byte("\n"))
			continue
		}

	loop:
		for m, cs := range menus {
			for _, c := range cs {
				if isCmd(l, c.name) {
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

func isCmd(cline, cname string) bool {
	return strings.Split(cline, " ")[0] == cname
}

func jumpToSubMenu(stdin io.WriteCloser, src, dst string) string {
	if src != dst {
		if src != "main" {
			// back to main menu from submenus
			_, _ = stdin.Write([]byte("back\n"))
			src = "main"
		}
		if dst != "main" {
			// from main menu to submenus
			stdin.Write([]byte("menu " + dst + "\n"))
		}
	}

	return dst
}

func printHelp() {
	pl := func(cmd *command) {
		fmt.Fprintf(os.Stderr, "  - %-30s %-50s %s\n", cmd.name, cmd.params, cmd.desc)
	}

	for m, cs := range menus {
		fmt.Fprintln(os.Stderr, "---------------")
		fmt.Fprintln(os.Stderr, m+":")
		for _, c := range cs {
			pl(&c)
		}
	}

	fmt.Fprintln(os.Stderr, "---------------")
	fmt.Fprintln(os.Stderr, "common:")
	for _, c := range commonCmds {
		pl(&c)
	}
}
