package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// SelectFromList displays numbered items and prompts the user to pick one.
// Returns the 0-based index of the selected item.
func SelectFromList(label string, items []string) (int, error) {
	if len(items) == 0 {
		return 0, fmt.Errorf("no items to select from")
	}
	if len(items) == 1 {
		fmt.Printf("  Auto-selected: %s\n", items[0])
		return 0, nil
	}

	for i, item := range items {
		fmt.Printf("  %2d. %s\n", i+1, item)
	}

	for {
		fmt.Printf("\n%s [1-%d]: ", label, len(items))
		input, err := readLine()
		if err != nil {
			return 0, err
		}

		n, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil || n < 1 || n > len(items) {
			fmt.Printf("  Please enter a number between 1 and %d\n", len(items))
			continue
		}
		return n - 1, nil
	}
}

// ReadInput prompts the user and reads a line of input
func ReadInput(prompt string) (string, error) {
	fmt.Print(prompt)
	return readLine()
}

func readLine() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
