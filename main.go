package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var (
	xrandrCmd         = flag.NewFlagSet("xrandr", flag.ExitOnError)
	externalMonitor   = xrandrCmd.Bool("external", false, "capture external monitors")
	integratedMonitor = xrandrCmd.Bool("integrated", false, "capture external monitors")
	filterActive      = xrandrCmd.Bool("filter-active", false, "filter-out active monitors")

	regexResolutions = `^\s+(\d+x\d+i?)\s+(\d+\.\d+)(\*?\+?)?.*`
	regexMonitors    = `^(\w+)\s((dis)?connected).*`
)

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("not enough parameters")
		os.Exit(1)
	}
	switch args[0] {
	case "xrandr":
		err := xrandrCmd.Parse(args[1:])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = xrandr()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	default:
		fmt.Printf("unknown command: %v\n", args[0])
		os.Exit(1)
	}
}

func xrandr() error {
	reMonitors := regexp.MustCompile(regexMonitors)
	reResolution := regexp.MustCompile(regexResolutions)

	cmd := exec.Command("xrandr")
	stdout, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("unable to run xrandr: %w", err)
	}

	reader := bytes.NewReader(stdout)
	scanner := bufio.NewScanner(reader)
	var monitors []string
	for scanner.Scan() {
		text := scanner.Text()

		// Find connected monitors
		monitor := reMonitors.FindStringSubmatch(text)
		if len(monitor) == 4 && monitor[2] == "connected" {
			if *externalMonitor && !strings.HasPrefix(monitor[1], "e") {
				monitors = append(monitors, monitor[1])
			}
			if *integratedMonitor && strings.HasPrefix(monitor[1], "e") {
				monitors = append(monitors, monitor[1])
			}
		}

		// Find active monitors using the active resolution
		resolution := reResolution.FindStringSubmatch(text)
		if *filterActive && len(resolution) == 4 {
			if resolution[3] != "" && len(monitors) > 0 {
				monitors = monitors[:len(monitors)-1]
			}
		}
	}

	if scanner.Err() != nil {
		return fmt.Errorf("failed to scan output: %w", scanner.Err())
	}

	if len(monitors) > 0 {
		fmt.Println(strings.Join(monitors, " "))
	}

	return nil
}
