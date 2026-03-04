package output

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

// ErrNotInteractive is returned when a prompt cannot run because
// the session is not an interactive terminal.
var ErrNotInteractive = errors.New("not an interactive terminal")

// Confirm prompts the user for a yes/no answer.
// When not interactive it returns defaultYes without prompting.
func Confirm(in io.Reader, out io.Writer, question string, defaultYes bool) (bool, error) {
	if !isTTY || in == nil {
		return defaultYes, nil
	}

	hint := "y/N"
	if defaultYes {
		hint = "Y/n"
	}
	scanner := bufio.NewScanner(in)
	for {
		fmt.Fprintf(out, "%s %s %s ", Yellow("?"), Bold(question), Dim("["+hint+"]"))
		if !scanner.Scan() {
			fmt.Fprintln(out)
			return defaultYes, scanner.Err()
		}
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))

		switch answer {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		case "":
			return defaultYes, nil
		default:
			fmt.Fprintf(out, "%s\n", Dim("Please answer y or n."))
		}
	}
}

// Select prompts the user to choose from a list using arrow keys.
// Returns the index of the selected option.
// When not interactive it returns ErrNotInteractive.
func Select(in *os.File, out io.Writer, question string, options []string) (int, error) {
	if len(options) == 0 {
		return -1, errors.New("no options to select from")
	}
	if !isTTY || in == nil {
		return -1, ErrNotInteractive
	}

	restore, err := enableRawMode(int(in.Fd()))
	if err != nil {
		return -1, fmt.Errorf("enable raw mode: %w", err)
	}
	defer restore()

	selected := 0
	maxVisible := maxVisibleOptions(in, 12)
	buf := make([]byte, 3)

	// Hide cursor during selection.
	fmt.Fprint(out, "\033[?25l")

	printedLines := printSelectUI(out, question, options, selected, maxVisible)

	for {
		n, readErr := in.Read(buf)
		if readErr != nil {
			fmt.Fprint(out, "\033[?25h")
			return -1, readErr
		}

		switch {
		case n == 1 && buf[0] == 3: // Ctrl+C
			fmt.Fprint(out, "\033[?25h")
			fmt.Fprintln(out)
			return -1, errors.New("interrupted")

		case n == 1 && (buf[0] == '\r' || buf[0] == '\n'): // Enter
			eraseLines(out, printedLines)
			fmt.Fprintf(out, "%s %s %s\r\n", Green(SymbolOK), Bold(question), Cyan(options[selected]))
			fmt.Fprint(out, "\033[?25h")
			return selected, nil

		case n == 3 && buf[0] == '\033' && buf[1] == '[' && buf[2] == 'A': // Up
			if selected > 0 {
				selected--
			} else {
				selected = len(options) - 1
			}
			eraseLines(out, printedLines)
			printedLines = printSelectUI(out, question, options, selected, maxVisible)

		case n == 3 && buf[0] == '\033' && buf[1] == '[' && buf[2] == 'B': // Down
			if selected < len(options)-1 {
				selected++
			} else {
				selected = 0
			}
			eraseLines(out, printedLines)
			printedLines = printSelectUI(out, question, options, selected, maxVisible)
		}
	}
}

// SearchSelect prompts the user to type and filter options while navigating with arrows.
// Returns the index of the selected option in the original options slice.
// When not interactive it returns ErrNotInteractive.
func SearchSelect(in *os.File, out io.Writer, question string, options []string) (int, error) {
	if len(options) == 0 {
		return -1, errors.New("no options to select from")
	}
	if !isTTY || in == nil {
		return -1, ErrNotInteractive
	}

	restore, err := enableRawMode(int(in.Fd()))
	if err != nil {
		return -1, fmt.Errorf("enable raw mode: %w", err)
	}
	defer restore()

	query := ""
	filtered := filterOptionIndices(options, query)
	cursor := 0
	maxVisible := maxVisibleOptions(in, 10)
	buf := make([]byte, 8)

	// Hide cursor during selection.
	fmt.Fprint(out, "\033[?25l")

	printedLines := printSearchSelectUI(out, question, options, filtered, query, cursor, maxVisible)

	for {
		n, readErr := in.Read(buf)
		if readErr != nil {
			fmt.Fprint(out, "\033[?25h")
			return -1, readErr
		}

		changed := false
		for i := 0; i < n; i++ {
			b := buf[i]

			// Arrow keys (escape sequences).
			if b == '\033' && i+2 < n && buf[i+1] == '[' {
				switch buf[i+2] {
				case 'A': // Up
					if len(filtered) > 0 {
						if cursor > 0 {
							cursor--
						} else {
							cursor = len(filtered) - 1
						}
						changed = true
					}
				case 'B': // Down
					if len(filtered) > 0 {
						if cursor < len(filtered)-1 {
							cursor++
						} else {
							cursor = 0
						}
						changed = true
					}
				}
				i += 2
				continue
			}

			switch b {
			case 3: // Ctrl+C
				fmt.Fprint(out, "\033[?25h")
				fmt.Fprintln(out)
				return -1, errors.New("interrupted")
			case '\r', '\n': // Enter
				if len(filtered) == 0 {
					continue
				}
				chosen := filtered[cursor]
				eraseLines(out, printedLines)
				fmt.Fprintf(out, "%s %s %s\r\n", Green(SymbolOK), Bold(question), Cyan(options[chosen]))
				fmt.Fprint(out, "\033[?25h")
				return chosen, nil
			case 127, 8: // Backspace/Delete
				if query != "" {
					_, size := utf8.DecodeLastRuneInString(query)
					if size > 0 && size <= len(query) {
						query = query[:len(query)-size]
						filtered = filterOptionIndices(options, query)
						if cursor >= len(filtered) {
							cursor = 0
						}
						changed = true
					}
				}
			default:
				if b >= 32 && b <= 126 { // printable ASCII
					query += string(b)
					filtered = filterOptionIndices(options, query)
					cursor = 0
					changed = true
				}
			}
		}

		if changed {
			eraseLines(out, printedLines)
			printedLines = printSearchSelectUI(out, question, options, filtered, query, cursor, maxVisible)
		}
	}
}

func printSelectUI(w io.Writer, question string, options []string, selected int, maxVisible int) int {
	start, end := visibleWindow(len(options), selected, maxVisible)
	lines := 0

	fmt.Fprintf(w, "%s %s\r\n", Yellow("?"), Bold(question))
	lines++

	if start > 0 {
		fmt.Fprintf(w, "    %s\r\n", Dim(fmt.Sprintf("... %d above", start)))
		lines++
	}
	for i := start; i < end; i++ {
		opt := options[i]
		if i == selected {
			fmt.Fprintf(w, "  %s %s\r\n", Cyan(SymbolArrow), Cyan(opt))
		} else {
			fmt.Fprintf(w, "    %s\r\n", Dim(opt))
		}
		lines++
	}
	if end < len(options) {
		fmt.Fprintf(w, "    %s\r\n", Dim(fmt.Sprintf("... %d more", len(options)-end)))
		lines++
	}

	return lines
}

func printSearchSelectUI(w io.Writer, question string, options []string, filtered []int, query string, cursor int, maxVisible int) int {
	lines := 0
	fmt.Fprintf(w, "%s %s\r\n", Yellow("?"), Bold(question))
	lines++

	queryDisplay := query
	if queryDisplay == "" {
		queryDisplay = Dim("(type to search)")
	}
	fmt.Fprintf(w, "  %s %s\r\n", Bold("Search:"), Cyan(queryDisplay))
	lines++
	fmt.Fprintf(w, "  %s\r\n", Dim("↑/↓ move  enter select  backspace delete"))
	lines++

	if len(filtered) == 0 {
		fmt.Fprintf(w, "    %s\r\n", Dim("No matches"))
		lines++
		return lines
	}

	start, end := visibleWindow(len(filtered), cursor, maxVisible)
	if start > 0 {
		fmt.Fprintf(w, "    %s\r\n", Dim(fmt.Sprintf("... %d above", start)))
		lines++
	}
	for i := start; i < end; i++ {
		origIdx := filtered[i]
		opt := options[origIdx]
		if i == cursor {
			fmt.Fprintf(w, "  %s %s\r\n", Cyan(SymbolArrow), Cyan(opt))
		} else {
			fmt.Fprintf(w, "    %s\r\n", Dim(opt))
		}
		lines++
	}
	if end < len(filtered) {
		fmt.Fprintf(w, "    %s\r\n", Dim(fmt.Sprintf("... %d more", len(filtered)-end)))
		lines++
	}
	return lines
}

// Input prompts the user for free-text input with an optional default value.
// Returns the default when not interactive or the user presses enter without typing.
func Input(in io.Reader, out io.Writer, question string, defaultVal string) (string, error) {
	if !isTTY || in == nil {
		return defaultVal, nil
	}

	hint := ""
	if defaultVal != "" {
		hint = " " + Dim("("+defaultVal+")")
	}
	fmt.Fprintf(out, "%s %s%s ", Yellow("?"), Bold(question), hint)

	scanner := bufio.NewScanner(in)
	if !scanner.Scan() {
		fmt.Fprintln(out)
		return defaultVal, scanner.Err()
	}
	answer := strings.TrimSpace(scanner.Text())
	if answer == "" {
		return defaultVal, nil
	}
	return answer, nil
}

// MultiSelect prompts the user to toggle multiple options with space and confirm with enter.
// preSelected sets which options start checked (nil means all unchecked).
// Returns indices of selected options.
func MultiSelect(in *os.File, out io.Writer, question string, options []string, preSelected []bool) ([]int, error) {
	if len(options) == 0 {
		return nil, errors.New("no options to select from")
	}
	if !isTTY || in == nil {
		return nil, ErrNotInteractive
	}

	restore, err := enableRawMode(int(in.Fd()))
	if err != nil {
		return nil, fmt.Errorf("enable raw mode: %w", err)
	}
	defer restore()

	cursor := 0
	checked := make([]bool, len(options))
	if preSelected != nil {
		copy(checked, preSelected)
	}
	maxVisible := maxVisibleOptions(in, 10)
	buf := make([]byte, 3)

	fmt.Fprint(out, "\033[?25l")
	printedLines := printMultiSelectUI(out, question, options, checked, cursor, maxVisible)

	for {
		n, readErr := in.Read(buf)
		if readErr != nil {
			fmt.Fprint(out, "\033[?25h")
			return nil, readErr
		}

		switch {
		case n == 1 && buf[0] == 3: // Ctrl+C
			fmt.Fprint(out, "\033[?25h")
			fmt.Fprintln(out)
			return nil, errors.New("interrupted")

		case n == 1 && buf[0] == ' ': // Space toggles
			checked[cursor] = !checked[cursor]
			eraseLines(out, printedLines)
			printedLines = printMultiSelectUI(out, question, options, checked, cursor, maxVisible)

		case n == 1 && buf[0] == 'a': // 'a' toggles all
			allChecked := true
			for _, c := range checked {
				if !c {
					allChecked = false
					break
				}
			}
			for i := range checked {
				checked[i] = !allChecked
			}
			eraseLines(out, printedLines)
			printedLines = printMultiSelectUI(out, question, options, checked, cursor, maxVisible)

		case n == 1 && (buf[0] == '\r' || buf[0] == '\n'): // Enter confirms
			eraseLines(out, printedLines)
			var selected []int
			var names []string
			for i, c := range checked {
				if c {
					selected = append(selected, i)
					names = append(names, options[i])
				}
			}
			summary := strings.Join(names, ", ")
			if summary == "" {
				summary = Dim("(none)")
			}
			fmt.Fprintf(out, "%s %s %s\r\n", Green(SymbolOK), Bold(question), Cyan(summary))
			fmt.Fprint(out, "\033[?25h")
			return selected, nil

		case n == 3 && buf[0] == '\033' && buf[1] == '[' && buf[2] == 'A': // Up
			if cursor > 0 {
				cursor--
			} else {
				cursor = len(options) - 1
			}
			eraseLines(out, printedLines)
			printedLines = printMultiSelectUI(out, question, options, checked, cursor, maxVisible)

		case n == 3 && buf[0] == '\033' && buf[1] == '[' && buf[2] == 'B': // Down
			if cursor < len(options)-1 {
				cursor++
			} else {
				cursor = 0
			}
			eraseLines(out, printedLines)
			printedLines = printMultiSelectUI(out, question, options, checked, cursor, maxVisible)
		}
	}
}

// SearchMultiSelect prompts the user to filter options by typing, toggle with space,
// and confirm with enter. preSelected sets which options start checked.
// Returns indices of selected options in the original options slice.
func SearchMultiSelect(in *os.File, out io.Writer, question string, options []string, preSelected []bool) ([]int, error) {
	if len(options) == 0 {
		return nil, errors.New("no options to select from")
	}
	if !isTTY || in == nil {
		return nil, ErrNotInteractive
	}

	restore, err := enableRawMode(int(in.Fd()))
	if err != nil {
		return nil, fmt.Errorf("enable raw mode: %w", err)
	}
	defer restore()

	query := ""
	filtered := filterOptionIndices(options, query)
	cursor := 0
	checked := make([]bool, len(options))
	if preSelected != nil {
		copy(checked, preSelected)
	}
	maxVisible := maxVisibleOptions(in, 9)
	buf := make([]byte, 16)

	fmt.Fprint(out, "\033[?25l")
	printedLines := printSearchMultiSelectUI(out, question, options, filtered, checked, query, cursor, maxVisible)

	for {
		n, readErr := in.Read(buf)
		if readErr != nil {
			fmt.Fprint(out, "\033[?25h")
			return nil, readErr
		}

		changed := false
		for i := 0; i < n; i++ {
			b := buf[i]

			// Arrow keys (escape sequences).
			if b == '\033' && i+2 < n && buf[i+1] == '[' {
				switch buf[i+2] {
				case 'A': // Up
					if len(filtered) > 0 {
						if cursor > 0 {
							cursor--
						} else {
							cursor = len(filtered) - 1
						}
						changed = true
					}
				case 'B': // Down
					if len(filtered) > 0 {
						if cursor < len(filtered)-1 {
							cursor++
						} else {
							cursor = 0
						}
						changed = true
					}
				}
				i += 2
				continue
			}

			switch b {
			case 3: // Ctrl+C
				fmt.Fprint(out, "\033[?25h")
				fmt.Fprintln(out)
				return nil, errors.New("interrupted")
			case '\r', '\n': // Enter confirms
				eraseLines(out, printedLines)
				var selected []int
				var names []string
				for idx, c := range checked {
					if c {
						selected = append(selected, idx)
						names = append(names, options[idx])
					}
				}
				summary := strings.Join(names, ", ")
				if summary == "" {
					summary = Dim("(none)")
				}
				fmt.Fprintf(out, "%s %s %s\r\n", Green(SymbolOK), Bold(question), Cyan(summary))
				fmt.Fprint(out, "\033[?25h")
				return selected, nil
			case ' ': // Space toggles current filtered item
				if len(filtered) > 0 {
					idx := filtered[cursor]
					checked[idx] = !checked[idx]
					changed = true
				}
			case 'a', 'A': // Toggle all
				allChecked := true
				for _, c := range checked {
					if !c {
						allChecked = false
						break
					}
				}
				for i := range checked {
					checked[i] = !allChecked
				}
				changed = true
			case 127, 8: // Backspace/Delete
				if query != "" {
					_, size := utf8.DecodeLastRuneInString(query)
					if size > 0 && size <= len(query) {
						query = query[:len(query)-size]
						filtered = filterOptionIndices(options, query)
						if cursor >= len(filtered) {
							cursor = 0
						}
						changed = true
					}
				}
			default:
				if b >= 32 && b <= 126 { // printable ASCII
					query += string(b)
					filtered = filterOptionIndices(options, query)
					cursor = 0
					changed = true
				}
			}
		}

		if changed {
			eraseLines(out, printedLines)
			printedLines = printSearchMultiSelectUI(out, question, options, filtered, checked, query, cursor, maxVisible)
		}
	}
}

func printMultiSelectUI(w io.Writer, question string, options []string, checked []bool, cursor int, maxVisible int) int {
	start, end := visibleWindow(len(options), cursor, maxVisible)
	lines := 0

	fmt.Fprintf(w, "%s %s\r\n", Yellow("?"), Bold(question))
	lines++

	if start > 0 {
		fmt.Fprintf(w, "    %s\r\n", Dim(fmt.Sprintf("... %d above", start)))
		lines++
	}
	for i := start; i < end; i++ {
		opt := options[i]
		box := "[ ]"
		if checked[i] {
			box = "[" + Green(SymbolOK) + "]"
		}
		if i == cursor {
			fmt.Fprintf(w, "  %s %s %s\r\n", Cyan(SymbolArrow), box, Cyan(opt))
		} else {
			fmt.Fprintf(w, "    %s %s\r\n", box, Dim(opt))
		}
		lines++
	}
	if end < len(options) {
		fmt.Fprintf(w, "    %s\r\n", Dim(fmt.Sprintf("... %d more", len(options)-end)))
		lines++
	}
	fmt.Fprintf(w, "  %s\r\n", Dim("space: toggle  a: all  enter: confirm"))
	lines++

	return lines
}

func printSearchMultiSelectUI(
	w io.Writer,
	question string,
	options []string,
	filtered []int,
	checked []bool,
	query string,
	cursor int,
	maxVisible int,
) int {
	lines := 0
	fmt.Fprintf(w, "%s %s\r\n", Yellow("?"), Bold(question))
	lines++

	queryDisplay := query
	if queryDisplay == "" {
		queryDisplay = Dim("(type to search)")
	}
	fmt.Fprintf(w, "  %s %s\r\n", Bold("Search:"), Cyan(queryDisplay))
	lines++

	if len(filtered) == 0 {
		fmt.Fprintf(w, "    %s\r\n", Dim("No matches"))
		lines++
		fmt.Fprintf(w, "  %s\r\n", Dim("type: filter  backspace: delete  enter: confirm"))
		lines++
		return lines
	}

	start, end := visibleWindow(len(filtered), cursor, maxVisible)
	if start > 0 {
		fmt.Fprintf(w, "    %s\r\n", Dim(fmt.Sprintf("... %d above", start)))
		lines++
	}
	for i := start; i < end; i++ {
		origIdx := filtered[i]
		opt := options[origIdx]
		box := "[ ]"
		if checked[origIdx] {
			box = "[" + Green(SymbolOK) + "]"
		}
		if i == cursor {
			fmt.Fprintf(w, "  %s %s %s\r\n", Cyan(SymbolArrow), box, Cyan(opt))
		} else {
			fmt.Fprintf(w, "    %s %s\r\n", box, Dim(opt))
		}
		lines++
	}
	if end < len(filtered) {
		fmt.Fprintf(w, "    %s\r\n", Dim(fmt.Sprintf("... %d more", len(filtered)-end)))
		lines++
	}
	fmt.Fprintf(w, "  %s\r\n", Dim("↑/↓ move  space: toggle  a: all  backspace: delete  enter: confirm"))
	lines++
	return lines
}

// eraseLines moves the cursor up n lines, clearing each one,
// leaving the cursor at the start of the topmost cleared line.
func eraseLines(w io.Writer, n int) {
	for i := 0; i < n; i++ {
		fmt.Fprint(w, "\033[1A") // move up
		fmt.Fprint(w, "\033[2K") // clear line
	}
	fmt.Fprint(w, "\r")
}

func maxVisibleOptions(in *os.File, fallback int) int {
	maxVisible := fallback
	if in != nil {
		if rows, _, ok := terminalSize(in.Fd()); ok {
			// Keep some headroom for prompt/summaries and shell prompt.
			candidate := rows - 6
			if candidate > 0 {
				maxVisible = candidate
			}
		}
	}
	if maxVisible < 5 {
		maxVisible = 5
	}
	return maxVisible
}

func visibleWindow(total, cursor, maxVisible int) (start, end int) {
	if total <= 0 {
		return 0, 0
	}
	if maxVisible <= 0 || total <= maxVisible {
		return 0, total
	}
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= total {
		cursor = total - 1
	}

	start = cursor - maxVisible/2
	if start < 0 {
		start = 0
	}
	end = start + maxVisible
	if end > total {
		end = total
		start = end - maxVisible
	}
	return start, end
}

func filterOptionIndices(options []string, query string) []int {
	if query == "" {
		all := make([]int, len(options))
		for i := range options {
			all[i] = i
		}
		return all
	}
	lower := strings.ToLower(strings.TrimSpace(query))
	idxs := make([]int, 0)
	for i, opt := range options {
		if strings.Contains(strings.ToLower(opt), lower) {
			idxs = append(idxs, i)
		}
	}
	return idxs
}
