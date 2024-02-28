package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	imageName := "dxclient"
	imageTag := "local"

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
	ttyFlag := "-it"
	if !isTTY() {
		ttyFlag = ""
	}

	// create volume directory if they don't exist
	if err := os.MkdirAll(volumeDir, 0755); err != nil {
		fmt.Println("Error createing volume directory:", err)
		return
	}

	// generate volume parameters
	volumeParams := fmt.Sprintf("-v \"%s/%s://dxclient/store\":Z", getCurrentDirectory(), volumeDir)

	// compose docker command
	dockerCmd := fmt.Sprintf("run -e VOLUME_DIR=\"%s\" -u %d %s %s --network=host --platform linux/amd64 --name dxclient --rm %s:%s ./bin/dxclient %s", volumeDir, os.Getuid(), ttyFlag, volumeParams, imageName, imageTag, strings.Join(args, " "))
	fmt.Println(dockerCmd)

	out, err := exec.Command(containerRuntime, dockerCmd).CombinedOutput()
	if err != nil {
		fmt.Println("Error executing Docker command: ", err)
		return
	}

	fmt.Println(string(out))
	cleanupFiles(args, volumeDir)
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
