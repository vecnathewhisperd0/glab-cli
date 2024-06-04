package git

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/cli/internal/config"
	"gitlab.com/gitlab-org/cli/internal/run"
)

func Test_StackRemoveRef(t *testing.T) {
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

func Test_StackLast(t *testing.T) {
	tests := []struct {
		name     string
		mockRefs map[string]StackRef
		expected StackRef
		wantErr  bool
	}{
		{
			name: "Find last ref",
			mockRefs: map[string]StackRef{
				"sha1": {Next: "sha2", SHA: "sha1"},
				"sha2": {Prev: "sha1", Next: "sha3", SHA: "sha2"},
				"sha3": {Prev: "sha2", SHA: "sha3"},
			},
			expected: StackRef{Prev: "sha2", SHA: "sha3"},
			wantErr:  false,
		},
		{
			name:     "No refs",
			mockRefs: map[string]StackRef{},
			expected: StackRef{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Stack{Refs: tt.mockRefs}
			got, err := s.Last()

			require.Equal(t, got, tt.expected)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func Test_StackFirst(t *testing.T) {
	tests := []struct {
		name     string
		mockRefs map[string]StackRef
		expected StackRef
		wantErr  bool
	}{
		{
			name: "Find last ref",
			mockRefs: map[string]StackRef{
				"sha1": {Next: "sha2", SHA: "sha1"},
				"sha2": {Prev: "sha1", Next: "sha3", SHA: "sha2"},
				"sha3": {Prev: "sha2", SHA: "sha3"},
			},
			expected: StackRef{Next: "sha2", SHA: "sha1"},
			wantErr:  false,
		},
		{
			name:     "No refs",
			mockRefs: map[string]StackRef{},
			expected: StackRef{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Stack{Refs: tt.mockRefs}
			got, err := s.First()

			require.Equal(t, got, tt.expected)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func Test_StackEmpty(t *testing.T) {
	s := Stack{Refs: make(map[string]StackRef)}
	if !s.Empty() {
		t.Errorf("Expected empty stack, but got non-empty")
	}

	s.Refs["sha"] = StackRef{}
	if s.Empty() {
		t.Errorf("Expected non-empty stack, but got empty")
	}
}

func Test_StackRemoveBranch(t *testing.T) {
	tests := []struct {
		name    string
		stack   Stack
		ref     StackRef
		wantErr bool
	}{
		{
			name: "remove single ref",
			stack: Stack{
				Title: "test-stack",
				Refs:  map[string]StackRef{"sha1": {SHA: "sha1", Branch: "branch123"}},
			},
			ref: StackRef{SHA: "sha1", Branch: "branch123"},
		},
		{
			name: "remove first ref",
			stack: Stack{
				Title: "test-stack",
				Refs: map[string]StackRef{
					"sha1": {SHA: "sha1", Next: "sha2", Branch: "branch123"},
					"sha2": {SHA: "sha2", Prev: "sha1", Branch: "branch456"},
				},
			},
			ref: StackRef{SHA: "sha1", Next: "sha2", Branch: "branch123"},
		},
		{
			name: "remove middle ref",
			stack: Stack{
				Title: "test-stack",
				Refs: map[string]StackRef{
					"sha1": {SHA: "sha1", Next: "sha2", Branch: "branch123"},
					"sha2": {SHA: "sha2", Prev: "sha1", Next: "sha3", Branch: "branch456"},
					"sha3": {SHA: "sha3", Prev: "sha2", Branch: "branch789"},
				},
			},
			ref: StackRef{SHA: "sha2", Prev: "sha1", Next: "sha3", Branch: "branch456"},
		},
		{
			name: "remove last ref",
			stack: Stack{
				Title: "test-stack",
				Refs: map[string]StackRef{
					"sha1": {SHA: "sha1", Next: "sha2", Branch: "branch123"},
					"sha2": {SHA: "sha2", Prev: "sha1", Branch: "branch456"},
				},
			},
			ref: StackRef{SHA: "sha2", Prev: "sha1", Branch: "branch456"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InitGitRepoWithCommit(t)

			gitAddRemote := GitCommand("remote", "add", "origin", "http://gitlab.com/gitlab-org/cli.git")
			_, err := run.PrepareCmd(gitAddRemote).Output()
			require.Nil(t, err)

			createBranches(t, tt.stack.Refs)

			err = tt.stack.RemoveBranch(tt.ref)

			require.Nil(t, err)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.Nil(t, err)

				showref := GitCommand("show-ref", "--verify", "--quiet", "refs/heads/"+tt.ref.Branch)
				_, err := run.PrepareCmd(showref).Output()
				require.Error(t, err)
			}
		})
	}
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
		wantErr  bool
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
		{
			name: "with bad start ref data",
			args: args{title: "sweet-title-123"},
			stacks: []StackRef{
				{SHA: "123", Prev: "", Next: "456"},
				{SHA: "456", Prev: "", Next: ""},
			},
			expected: Stack{},
			wantErr:  true,
		},
		{
			name: "with bad end ref data",
			args: args{title: "sweet-title-123"},
			stacks: []StackRef{
				{SHA: "123", Prev: "", Next: ""},
				{SHA: "456", Prev: "123", Next: ""},
			},
			expected: Stack{},
			wantErr:  true,
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

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
			}

			require.Equal(t, stack, tt.expected)
		})
	}
}

func Test_adjustAdjacentRefs(t *testing.T) {
	type args struct {
		title  string
		adjust StackRef
	}
	tests := []struct {
		name     string
		args     args
		stacks   []StackRef
		expected Stack
		wantErr  bool
	}{
		{
			name: "with multiple files",
			args: args{
				title:  "sweet-title-123",
				adjust: StackRef{SHA: "456", Prev: "123", Next: "789"},
			},
			stacks: []StackRef{
				{SHA: "456", Prev: "123", Next: "789"},
				{SHA: "123", Prev: "", Next: "456"},
				{SHA: "789", Prev: "456", Next: ""},
			},
			expected: Stack{
				Refs: map[string]StackRef{
					"123": {SHA: "123", Prev: "", Next: "789"},
					"456": {SHA: "456", Prev: "123", Next: "789"},
					"789": {SHA: "789", Prev: "123", Next: ""},
				},
				Title: "sweet-title-123",
			},
		},
		{
			name: "with multiple files, beginning ref",
			args: args{
				title:  "sweet-title-123",
				adjust: StackRef{SHA: "123", Prev: "", Next: "456"},
			},
			stacks: []StackRef{
				{SHA: "456", Prev: "123", Next: "789"},
				{SHA: "123", Prev: "", Next: "456"},
				{SHA: "789", Prev: "456", Next: ""},
			},
			expected: Stack{
				Refs: map[string]StackRef{
					"123": {SHA: "123", Prev: "", Next: "456"},
					"456": {SHA: "456", Prev: "", Next: "789"},
					"789": {SHA: "789", Prev: "456", Next: ""},
				},
				Title: "sweet-title-123",
			},
		},
		{
			name: "with multiple files, end ref",
			args: args{
				title:  "sweet-title-123",
				adjust: StackRef{SHA: "789", Prev: "456", Next: ""},
			},
			stacks: []StackRef{
				{SHA: "123", Prev: "", Next: "456"},
				{SHA: "456", Prev: "123", Next: "789"},
				{SHA: "789", Prev: "456", Next: ""},
			},
			expected: Stack{
				Refs: map[string]StackRef{
					"123": {SHA: "123", Prev: "", Next: "456"},
					"456": {SHA: "456", Prev: "123", Next: ""},
					"789": {SHA: "789", Prev: "456", Next: ""},
				},
				Title: "sweet-title-123",
			},
		},
		{
			name: "with 1 file",
			args: args{
				title:  "sweet-title-123",
				adjust: StackRef{SHA: "123", Prev: "", Next: ""},
			},
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

			originalStack, err := GatherStackRefs(tt.args.title)
			require.Nil(t, err)

			err = originalStack.adjustAdjacentRefs(tt.args.adjust)
			require.Nil(t, err)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
			}

			require.Equal(t, tt.expected.Refs, originalStack.Refs)
		})
	}
}

func Test_validateStackRefs(t *testing.T) {
	tests := []struct {
		name    string
		stack   Stack
		wantErr bool
	}{
		{
			name: "valid stack",
			stack: Stack{
				Refs: map[string]StackRef{
					"1": {SHA: "1", Prev: "", Next: "2"},
					"2": {SHA: "2", Prev: "1", Next: "3"},
					"3": {SHA: "3", Prev: "2", Next: ""},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple start refs",
			stack: Stack{
				Refs: map[string]StackRef{
					"1": {SHA: "1", Prev: "", Next: "2"},
					"2": {SHA: "2", Prev: "", Next: "3"},
					"3": {SHA: "3", Prev: "2", Next: ""},
				},
			},
			wantErr: true,
		},
		{
			name: "multiple end refs",
			stack: Stack{
				Refs: map[string]StackRef{
					"1": {SHA: "1", Prev: "", Next: "2"},
					"2": {SHA: "2", Prev: "1", Next: ""},
					"3": {SHA: "3", Prev: "2", Next: ""},
				},
			},
			wantErr: true,
		},
		{
			name: "empty stack",
			stack: Stack{
				Refs: map[string]StackRef{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStackRefs(tt.stack)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func createBranches(t *testing.T, refs map[string]StackRef) {
	for _, ref := range refs {
		err := CheckoutNewBranch(ref.Branch)
		require.Nil(t, err)
	}
}
