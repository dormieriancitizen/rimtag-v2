package main

import (
	"fmt"
	"os"
	"os/exec"
)

func ToddsClean(config Config) error {
	toEncode := []string{config.LocalModSrc, config.SteamModSrc}
	for _, path := range toEncode {
		cmd := exec.Command("todds", "-cl", "-v", "-p", "-r", "Textures", "-t", path)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				return (fmt.Errorf("todds exited with code %d\n", exitErr.ExitCode()))
			} else {
				return (fmt.Errorf("failed to run todds: %v\n", err))
			}
		}
	}
	return nil
}

func ToddsEncode(config Config) error {
	toEncode := []string{config.LocalModSrc, config.SteamModSrc}
	for _, path := range toEncode {
		cmd := exec.Command("todds", "-f", "BC1", "-v", "-p", "-af", "BC7", "-on", "-vf", "-fs", "-r", "Textures", "-t", path)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				return (fmt.Errorf("todds exited with code %d\n", exitErr.ExitCode()))
			} else {
				return (fmt.Errorf("failed to run todds: %v\n", err))
			}
		}
	}
	return nil
}
