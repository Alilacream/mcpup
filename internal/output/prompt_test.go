package output

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestConfirmDefaultNoOnEmptyInput(t *testing.T) {
	SetTTY(true)
	t.Cleanup(func() { SetTTY(false) })

	in := strings.NewReader("\n")
	var out bytes.Buffer

	ok, err := Confirm(in, &out, "Enable on clients now?", false)
	if err != nil {
		t.Fatalf("confirm returned error: %v", err)
	}
	if ok {
		t.Fatalf("expected default false on empty input")
	}
}

func TestConfirmRepromptsOnInvalidInput(t *testing.T) {
	SetTTY(true)
	t.Cleanup(func() { SetTTY(false) })

	in := strings.NewReader("maybe\nn\n")
	var out bytes.Buffer

	ok, err := Confirm(in, &out, "Enable on clients now?", true)
	if err != nil {
		t.Fatalf("confirm returned error: %v", err)
	}
	if ok {
		t.Fatalf("expected explicit no to return false")
	}
	if !strings.Contains(out.String(), "Please answer y or n.") {
		t.Fatalf("expected invalid-input hint in prompt output")
	}
}

func TestVisibleWindow(t *testing.T) {
	tests := []struct {
		name       string
		total      int
		cursor     int
		maxVisible int
		want       [2]int
	}{
		{
			name:       "all fit",
			total:      4,
			cursor:     0,
			maxVisible: 10,
			want:       [2]int{0, 4},
		},
		{
			name:       "start of long list",
			total:      30,
			cursor:     0,
			maxVisible: 8,
			want:       [2]int{0, 8},
		},
		{
			name:       "middle of long list",
			total:      30,
			cursor:     15,
			maxVisible: 8,
			want:       [2]int{11, 19},
		},
		{
			name:       "end of long list",
			total:      30,
			cursor:     29,
			maxVisible: 8,
			want:       [2]int{22, 30},
		},
		{
			name:       "cursor out of range",
			total:      20,
			cursor:     99,
			maxVisible: 6,
			want:       [2]int{14, 20},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			start, end := visibleWindow(tc.total, tc.cursor, tc.maxVisible)
			got := [2]int{start, end}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("visibleWindow(%d, %d, %d) = %v, want %v", tc.total, tc.cursor, tc.maxVisible, got, tc.want)
			}
		})
	}
}

func TestFilterOptionIndices(t *testing.T) {
	options := []string{
		"github                GitHub - repos and PRs",
		"gitlab                GitLab - merge requests",
		"google-calendar       Calendar integration",
		"notion                Notes and docs",
	}

	tests := []struct {
		name  string
		query string
		want  []int
	}{
		{
			name:  "empty query returns all",
			query: "",
			want:  []int{0, 1, 2, 3},
		},
		{
			name:  "case insensitive match",
			query: "GIT",
			want:  []int{0, 1},
		},
		{
			name:  "match by description text",
			query: "calendar",
			want:  []int{2},
		},
		{
			name:  "no results",
			query: "stripe",
			want:  []int{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := filterOptionIndices(options, tc.query)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("filterOptionIndices(%q) = %v, want %v", tc.query, got, tc.want)
			}
		})
	}
}
