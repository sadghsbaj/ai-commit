package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Spinner represents a simple CLI spinner
type Spinner struct {
	stopChan chan struct{}
}

// StartSpinner starts a loading spinner with the given message.
func StartSpinner(message string) *Spinner {
	s := &Spinner{
		stopChan: make(chan struct{}),
	}
	go func() {
		frames := []string{"|", "/", "-", "\\"}
		i := 0
		for {
			select {
			case <-s.stopChan:
				fmt.Print("\r\033[K") // Clear line
				return
			default:
				fmt.Printf("\r\033[36m%s\033[0m %s", frames[i%len(frames)], message)
				time.Sleep(100 * time.Millisecond)
				i++
			}
		}
	}()
	return s
}

// Stop stops the spinner.
func (s *Spinner) Stop() {
	close(s.stopChan)
}

// SendDesktopNotification uses notify-send on Linux desktop.
func SendDesktopNotification(title, message string) {
	cmd := exec.Command("notify-send", title, message)
	_ = cmd.Run() // Ignore errors, as it's optional
}

// PromptUser asks the user for action based on the suggestion.
// Returns 'a' for accept, 'r' for reject, 'c' for comment.
func PromptUser(suggestion string) (rune, string) {
	fmt.Printf("\n\033[1;32mGenerated Commit Message:\033[0m\n\n%s\n\n", suggestion)

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\033[1;33mAction: [a]ccept, [r]eject, [c]omment & retry: \033[0m")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		input = strings.ToLower(input)

		if input == "a" || strings.HasPrefix(input, "a\n") || strings.HasPrefix(input, "a\r") {
			return 'a', ""
		} else if input == "r" || strings.HasPrefix(input, "r\n") || strings.HasPrefix(input, "r\r") {
			return 'r', ""
		} else if input == "c" || strings.HasPrefix(input, "c\n") || strings.HasPrefix(input, "c\r") {
			fmt.Print("\033[1;36mWhat should be changed? \033[0m")
			comment, _ := reader.ReadString('\n')
			return 'c', strings.TrimSpace(comment)
		} else {
			fmt.Println("Invalid input, please enter a, r, or c.")
		}
	}
}

// Color Print functions
func PrintError(msg string, err error) {
	if err != nil {
		fmt.Printf("\033[1;31mError: %s: %v\033[0m\n", msg, err)
	} else {
		fmt.Printf("\033[1;31mError: %s\033[0m\n", msg)
	}
}

func PrintSuccess(msg string) {
	fmt.Printf("\033[1;32m%s\033[0m\n", msg)
}
