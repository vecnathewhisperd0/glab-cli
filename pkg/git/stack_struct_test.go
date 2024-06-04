package git

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/cli/internal/config"
)

func Test_RemoveRef(t *testing.T) {
	type args struct {
		stack  Stack
		remove StackRef
	}
	tests := []struct {
		name     string
		args     args
		expected map[string]StackRef
	}{
		{
			name: "with multiple files",
			args: args{
				remove: StackRef{SHA: "456", Prev: "123", Next: "789"},
				stack: Stack{
					Title: "sweet-title-123",
					Refs: map[string]StackRef{
						"123": {SHA: "123", Prev: "", Next: "456"},
						"456": {SHA: "456", Prev: "123", Next: "789"},
						"789": {SHA: "789", Prev: "456", Next: ""},
					}}},
			expected: map[string]StackRef{
				"123": {SHA: "123", Prev: "", Next: "789"},
				"789": {SHA: "789", Prev: "123", Next: ""},
			},
		},
		{
			name: "with 1 file",
			args: args{
				stack: Stack{
					Title: "sweet-title-123",
					Refs:  map[string]StackRef{"123": {SHA: "123", Prev: "", Next: ""}}},
				remove: StackRef{SHA: "123", Prev: "", Next: ""},
			},
			expected: map[string]StackRef{},
		},
		{
			name: "large number",
			args: args{
				stack: Stack{
					Title: "title-123",
					Refs: map[string]StackRef{
						"1":  {SHA: "1", Prev: "", Next: "2"},
						"2":  {SHA: "2", Prev: "1", Next: "3"},
						"3":  {SHA: "3", Prev: "2", Next: "4"},
						"4":  {SHA: "4", Prev: "3", Next: "5"},
						"5":  {SHA: "5", Prev: "4", Next: "6"},
						"6":  {SHA: "6", Prev: "5", Next: "7"},
						"7":  {SHA: "7", Prev: "6", Next: "8"},
						"8":  {SHA: "8", Prev: "7", Next: "9"},
						"9":  {SHA: "9", Prev: "8", Next: "10"},
						"10": {SHA: "10", Prev: "9", Next: "11"},
						"11": {SHA: "11", Prev: "10", Next: "12"},
						"12": {SHA: "12", Prev: "11", Next: "13"},
						"13": {SHA: "13", Prev: "12", Next: ""},
					}},
				remove: StackRef{SHA: "11", Prev: "10", Next: "12"},
			},
			expected: map[string]StackRef{
				"1":  {SHA: "1", Prev: "", Next: "2"},
				"2":  {SHA: "2", Prev: "1", Next: "3"},
				"3":  {SHA: "3", Prev: "2", Next: "4"},
				"4":  {SHA: "4", Prev: "3", Next: "5"},
				"5":  {SHA: "5", Prev: "4", Next: "6"},
				"6":  {SHA: "6", Prev: "5", Next: "7"},
				"7":  {SHA: "7", Prev: "6", Next: "8"},
				"8":  {SHA: "8", Prev: "7", Next: "9"},
				"9":  {SHA: "9", Prev: "8", Next: "10"},
				"10": {SHA: "10", Prev: "9", Next: "12"},
				"12": {SHA: "12", Prev: "10", Next: "13"},
				"13": {SHA: "13", Prev: "12", Next: ""},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := InitGitRepo(t)

			err := createRefFiles(tt.args.stack.Refs, tt.args.stack.Title)
			require.Nil(t, err)

			err = tt.args.stack.RemoveRef(tt.args.remove)
			require.Nil(t, err)

			require.Equal(t, tt.expected, tt.args.stack.Refs)

			wantpath := path.Join(dir, stackLocation, tt.args.remove.Branch, ".json")
			require.False(t, config.CheckFileExists(wantpath))
		})
	}
}

func Test_RemoveBranch(t *testing.T) {
	// TODO: write test
}

func Test_adjustAdjacentRefs(t *testing.T) {
	// TODO: write test
}

func Test_Last(t *testing.T) {
	// TODO: write test
}

func Test_First(t *testing.T) {
	// TODO: write test
}

func Test_GatherStackRefs(t *testing.T) {
	type args struct {
		title string
	}
	tests := []struct {
		name     string
		args     args
		stacks   []StackRef
		expected Stack
	}{
		{
			name: "with multiple files",
			args: args{title: "sweet-title-123"},
			stacks: []StackRef{
				{SHA: "456", Prev: "123", Next: "789"},
				{SHA: "123", Prev: "", Next: "456"},
				{SHA: "789", Prev: "456", Next: ""},
			},
			expected: Stack{
				Refs: map[string]StackRef{
					"123": {SHA: "123", Prev: "", Next: "456"},
					"456": {SHA: "456", Prev: "123", Next: "789"},
					"789": {SHA: "789", Prev: "456", Next: ""},
				},
				Title: "sweet-title-123",
			},
		},
		{
			name: "with 1 file",
			args: args{title: "sweet-title-123"},
			stacks: []StackRef{
				{SHA: "123", Prev: "", Next: ""},
			},
			expected: Stack{
				Refs: map[string]StackRef{
					"123": {SHA: "123", Prev: "", Next: ""},
				},
				Title: "sweet-title-123",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InitGitRepo(t)

			for _, stack := range tt.stacks {
				err := AddStackRefFile(tt.args.title, stack)
				require.Nil(t, err)
			}

			stack, err := GatherStackRefs(tt.args.title)
			require.Nil(t, err)

			require.Equal(t, stack, tt.expected)
		})
	}
}
