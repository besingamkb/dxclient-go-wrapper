package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	imageName := "hcl/dx/client"
	imageTag := "95_CF223_20240905-0159"

	// parse command-line arguments
	args := os.Args[1:]

	containerFolder := "/dxclient/store"
	for i, arg := range args {
		if isPath(arg) {
			localPath := arg
			baseName := filepath.Base(localPath)
			args[i] = filepath.Join(containerFolder, baseName)
		}
	}

	// check for dependencies
	// basically check if docker is registered properly on the PATH
	if err := checkDependencies("docker"); err != nil {
		fmt.Println(err)
		return
	}

	// handle environment variables
	volumeDir := os.Getenv("VOLUME_DIR")
	if volumeDir == "" {
		volumeDir = "store"
	}

	containerRuntime := os.Getenv("CONTAINER_RUNTIME")
	if containerRuntime == "" {
		containerRuntime = "docker"
	}

	// determine if running in a TTY
	ttyFlag := ""
	if isTTY() {
		ttyFlag = "-it"
	}

	// create volume directory if they don't exist
	if err := os.MkdirAll(volumeDir, 0777); err != nil {
		fmt.Println("Error createing volume directory:", err)
		return
	}

	// generate volume parameters
	volumeParams := fmt.Sprintf("-v \"%s/%s://dxclient/store\":Z", getCurrentDirectory(), volumeDir)

	// compose docker command
	dockerCmd := fmt.Sprintf("%s run -e VOLUME_DIR=\"%s\" -u %d %s %s --network=host --platform linux/amd64 --name dxclient --rm %s:%s ./bin/dxclient %s", containerRuntime, volumeDir, os.Getuid(), ttyFlag, volumeParams, imageName, imageTag, strings.Join(args, " "))
	fmt.Println(dockerCmd)

	// Execute the command based on the OS
	executeCommand(dockerCmd)

	cleanupFiles(args, volumeDir)
}

func executeCommand(dockerCmd string) {
	var cmd *exec.Cmd

	// Check the OS type
	switch runtime.GOOS {
	case "windows":
		// Use cmd.exe on Windows
		cmd = exec.Command("cmd.exe", "/C", dockerCmd)
	default:
		// Use sh on Unix-like systems (macOS, Linux)
		cmd = exec.Command("sh", "-c", dockerCmd)
	}

	// Run the command and capture output
	out, err := cmd.CombinedOutput()

	// Print the output and handle errors
	fmt.Println(string(out))
	if err != nil {
		fmt.Println("Error executing Docker command: ", err)
	}
}

// return error if "docker" command not found
func checkDependencies(cmd string) error {
	_, err := exec.LookPath(cmd)
	return err
}

// this is to check if the stdin is from a terminal and is currently connected
func isTTY() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeCharDevice) != 0
}

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir((os.Args[0])))
	if err != nil {
		fmt.Println("Error getting current directory: ", err)
		return ""
	}
	return strings.ReplaceAll(dir, "\\", "/")
}

func cleanupFiles(args []string, volumeDir string) {
	for _, arg := range args {
		if _, err := os.Stat(arg); err == nil {
			absPath := filepath.Join(volumeDir, filepath.Base(arg))
			err := os.Remove(absPath)
			if err != nil {
				fmt.Printf("Error removing file :%s: %s\n", absPath, err)
			}
		}
	}
}

func isPath(arg string) bool {
	_, err := os.Stat(arg)
	return err == nil
}
