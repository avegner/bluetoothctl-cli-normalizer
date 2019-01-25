package main

import (
	"bufio"
	"bytes"
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
	if err := setTermParams(); err != nil {
		return fmt.Errorf("set term params: %v", err)
	}
	// TODO: catch signals to reset term params for all exit cases
	defer func() {
		if err := resetTermParams(); err != nil {
			fmt.Fprintf(os.Stderr, "reset term params: %v", err)
		}
	}()

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

	s := bufio.NewScanner(os.Stdin)
	s.Split(bufio.ScanBytes)
	menu := "main"
	buf := make([]byte, 128)
	for s.Scan() {
		bs := s.Bytes()
		if isCtlKey(bs) {
			_, _ = stdin.Write(bs)
			continue
		}
		if strings.Index(string(bs), "\n") == -1 {
			fmt.Fprint(os.Stdout, string(bs))
			buf = append(buf, bs...)
			continue
		}

		l := string(buf)
		buf = buf[:0]

		// exit commands
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
		// run command in required menu
		menu = chooseMenu(stdin, menu, l)
		_, _ = stdin.Write([]byte(l + "\n"))
	}

	// close stdin to make bluetoothctl stop
	// TODO: make a proper handling of errors
	_ = stdin.Close()
	return <-errc
}

func setTermParams() error {
	return exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1", "-echo").Run()
}

func resetTermParams() error {
	return exec.Command("stty", "-F", "/dev/tty", "-cbreak", "echo").Run()
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

func chooseMenu(stdin io.WriteCloser, menu, cline string) string {
loop:
	for m, cs := range menus {
		for _, c := range cs {
			if isCmd(cline, c.name) {
				if menu != m {
					if menu != "main" {
						// return to main menu
						_, _ = stdin.Write([]byte("back\n"))
						menu = "main"
					}
					if m != "main" {
						// choose menu
						stdin.Write([]byte("menu " + m + "\n"))
					}
				}
				menu = m
				break loop
			}
		}
	}

	return menu
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

func isCtlKey(bs []byte) bool {
	keys := [][]byte{
		// tab
		[]byte{0x09},
		// up
		[]byte{0x1B, 0x5B, 0x41},
		// down
		[]byte{0x1B, 0x5B, 0x42},
		// right
		[]byte{0x1B, 0x5B, 0x43},
		// left
		[]byte{0x1B, 0x5B, 0x44},
		// del
		[]byte{0x1B, 0x5B, 0x33, 0x7E},
	}

	for _, k := range keys {
		if bytes.Equal(k, bs) {
			return true
		}
	}

	return false
}
