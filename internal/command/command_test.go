package command

import "testing"

func TestCommand_Narg(t *testing.T) {
	tests := []struct {
		name      string
		fn        func([]string) error
		args      []string
		wantedErr string
	}{
		{
			name:      "runUpdate",
			fn:        runUpdate,
			args:      []string{"subcommand"},
			wantedErr: "usage: tokensnap update [-tokens-dir] [-snapshot-file] [-snapshot-dir]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn(tt.args)

			if err == nil {
				t.Fatalf("expected error, got nil")
			}

			if err.Error() != tt.wantedErr {
				t.Errorf("want: %q, got: %q", tt.wantedErr, err.Error())
			}
		})
	}
}
