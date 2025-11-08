package installer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ildx/merlin/internal/models"
)

func TestParseSelection(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxIndex  int
		want      []int
		wantError bool
	}{
		{
			name:     "single number",
			input:    "1",
			maxIndex: 5,
			want:     []int{0},
		},
		{
			name:     "multiple numbers",
			input:    "1 3 5",
			maxIndex: 5,
			want:     []int{0, 2, 4},
		},
		{
			name:     "range",
			input:    "1-3",
			maxIndex: 5,
			want:     []int{0, 1, 2},
		},
		{
			name:     "mixed",
			input:    "1 3-5 7",
			maxIndex: 10,
			want:     []int{0, 2, 3, 4, 6},
		},
		{
			name:     "with extra spaces",
			input:    "1  3   5",
			maxIndex: 5,
			want:     []int{0, 2, 4},
		},
		{
			name:      "out of bounds",
			input:     "1 10",
			maxIndex:  5,
			wantError: true,
		},
		{
			name:      "invalid range",
			input:     "5-3",
			maxIndex:  5,
			wantError: true,
		},
		{
			name:      "invalid number",
			input:     "abc",
			maxIndex:  5,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSelection(tt.input, tt.maxIndex)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if len(got) != len(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
				return
			}
			
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("got %v, want %v", got, tt.want)
					return
				}
			}
		})
	}
}

func TestSelectPackages(t *testing.T) {
	packages := []models.BrewPackage{
		{Name: "package1", Description: "First package"},
		{Name: "package2", Description: "Second package"},
		{Name: "package3", Description: "Third package"},
	}

	tests := []struct {
		name      string
		input     string
		wantCount int
		wantError bool
	}{
		{
			name:      "select all",
			input:     "all\n",
			wantCount: 3,
		},
		{
			name:      "select none",
			input:     "none\n",
			wantCount: 0,
		},
		{
			name:      "select by numbers",
			input:     "1 3\n",
			wantCount: 2,
		},
		{
			name:      "select by range",
			input:     "1-2\n",
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader(tt.input)
			output := &bytes.Buffer{}
			
			selected, err := SelectPackages(packages, "Test Packages", input, output)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if len(selected) != tt.wantCount {
				t.Errorf("got %d packages, want %d", len(selected), tt.wantCount)
			}
		})
	}
}

func TestConfirmInstallation(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		formulaeCount  int
		casksCount     int
		wantConfirmed  bool
	}{
		{
			name:          "confirm with y",
			input:         "y\n",
			formulaeCount: 5,
			casksCount:    3,
			wantConfirmed: true,
		},
		{
			name:          "confirm with yes",
			input:         "yes\n",
			formulaeCount: 5,
			casksCount:    3,
			wantConfirmed: true,
		},
		{
			name:          "decline with n",
			input:         "n\n",
			formulaeCount: 5,
			casksCount:    3,
			wantConfirmed: false,
		},
		{
			name:          "decline with no",
			input:         "no\n",
			formulaeCount: 5,
			casksCount:    3,
			wantConfirmed: false,
		},
		{
			name:          "decline with empty",
			input:         "\n",
			formulaeCount: 5,
			casksCount:    3,
			wantConfirmed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := strings.NewReader(tt.input)
			output := &bytes.Buffer{}
			
			confirmed, err := ConfirmInstallation(tt.formulaeCount, tt.casksCount, input, output)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if confirmed != tt.wantConfirmed {
				t.Errorf("got confirmed=%v, want %v", confirmed, tt.wantConfirmed)
			}
		})
	}
}

