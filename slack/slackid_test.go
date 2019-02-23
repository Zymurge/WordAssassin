package slack

import (
	"testing"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	type args struct {
		in0 string
	}
	tests := []struct {
		name      string
		args      args
		wantValid bool
		wantErr   string // "" denotes no error expected
	}{
		{
			name: "Positive",
			args: args{"WVALIDNAME"},
			wantValid: true,
			wantErr: "",
		},
		{
			name: "@ prefix should fail",
			args: args{"@UBVALID"},
			wantValid: false,
			wantErr: "must start with ",
		},
		{
			name: "embedded non-alphanum should fail",
			args: args{"U not valid"},
			wantValid: false,
			wantErr: "consist of only alphanumberics",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValid, err := Validate(tt.args.in0)
			require.Equal(t, tt.wantValid, gotValid, "Expected valid verdict mismatch")
			if err != nil {
				require.False(t, tt.wantValid, "Expected valid verdict, but received %v", err)
				require.Contains(t, err.Error(), tt.wantErr, "Error should contain expected message")
			} else {
				require.True(t, tt.wantValid, "Expected valid verdict, but received 'false' with no reason")
				require.Nil(t, err, "Error should not be returned, but received %v", err)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		in0 string
	}
	tests := []struct {
		name      string
		args      args
		wantSID   SlackID
		wantErr   string // "" denotes no error expected
	}{
		{
			name: "Positive",
			args: args{"WVALIDNAME"},
			wantSID: SlackID("WVALIDNAME"),
			wantErr: "",
		},
		{
			name: "@ prefix should fail",
			args: args{"@UBVALID"},
			wantSID: "",
			wantErr: "must start with ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSID, err := New(tt.args.in0)
			require.Equal(t, tt.wantSID, gotSID, "Expected valid response mismatch")
			if err != nil {
				require.Contains(t, err.Error(), tt.wantErr, "Error should contain expected message")
			}
		})
	}
}
