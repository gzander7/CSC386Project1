// Package main provides a simple interactive shell program in Go.
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// main is the entry point of the program.
func main() {
	// Create a reader to read input from standard input
	reader := bufio.NewReader(os.Stdin)

	// Loop indefinitely to continuously accept user input
	for {
		// Print the shell prompt.
		fmt.Print("GagesGoShell> ")

		// Read input from the user until a newline character
		input, err := reader.ReadString('\n')
		if err != nil {
			// If there is an error reading input print the error message and continue to the next iteration
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			continue
		}

		// Remove newline character from input
		input = strings.TrimSpace(input)

		// Split input into command and arguments
		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue // Skip empty input lines
		}

		// Check for commands
		switch parts[0] {
		case "ls":
			// Execute ls command with arguments
			cmd := exec.Command("ls", parts[1:]...)
			cmd.Stdout = os.Stdout // Set command's standard output to os.Stdout
			cmd.Stderr = os.Stderr // Set command's standard error to os.Stderr
			err := cmd.Run()       // Run the command
			if err != nil {
				fmt.Println("Error:", err) // Print error message if command execution fails
			}
		case "wc":
			// Execute wc command with arguments
			cmd := exec.Command("wc", parts[1:]...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				fmt.Println("Error:", err)
			}
		case "mkdir":
			// Execute mkdir command with arguments
			cmd := exec.Command("mkdir", parts[1:]...)
			err := cmd.Run()
			if err != nil {
				fmt.Println("Error:", err)
			}
		case "cp":
			// Execute cp command with arguments
			cmd := exec.Command("cp", parts[1:]...)
			err := cmd.Run()
			if err != nil {
				fmt.Println("Error:", err)
			}
		case "mv":
			// Execute mv command with arguments
			cmd := exec.Command("mv", parts[1:]...)
			err := cmd.Run()
			if err != nil {
				fmt.Println("Error:", err)
			}
		case "cd":
			// Change directory to the specified path
			if len(parts) < 2 {
				fmt.Println("Usage: cd [directory]")
				continue
			}
			err := os.Chdir(parts[1])
			if err != nil {
				fmt.Println("Error:", err)
			}
		case "whoami":
			// Print users name and user ID
			fmt.Println("User:", os.Getenv("USER"))
		case "exit":
			// Exit the shell
			os.Exit(0)
		default:
			// Print message for unsupported commands
			fmt.Println("Command not supported:", parts[0])
		}
	}
}
