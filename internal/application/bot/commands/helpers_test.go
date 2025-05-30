package commands_test

import (
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/internal/application/bot/commands"
	"github.com/stretchr/testify/require"
)

func TestValidateLink(t *testing.T) {
	tests := map[string]struct {
		link    string
		wantErr bool
	}{
		"valid github": {
			link:    "https://github.com/user/repo",
			wantErr: false,
		},
		"valid stackoverflow": {
			link:    "https://stackoverflow.com/questions/12345",
			wantErr: false,
		},
		"invalid http": {
			link:    "http://example.com",
			wantErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := commands.ValidateLink(test.link)
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMatchTags(t *testing.T) {
	tests := map[string]struct {
		got      []string
		desired  []string
		expected bool
	}{
		"different order": {
			got:      []string{"go", "dev"},
			desired:  []string{"dev", "go"},
			expected: true,
		},
		"lacking tag": {
			got:      []string{"go"},
			desired:  []string{"go", "dev"},
			expected: false,
		},
		"completely different": {
			got:      []string{"go"},
			desired:  []string{"python"},
			expected: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.expected, commands.MatchTags(test.got, test.desired))
		})
	}
}

func TestMatchFilters(t *testing.T) {
	tests := map[string]struct {
		got      []string
		desired  []string
		expected bool
	}{
		"different order": {
			got:      []string{"user:admin", "repo:go"},
			desired:  []string{"repo:go", "user:admin"},
			expected: true,
		},
		"lacking filter": {
			got:      []string{"user:admin"},
			desired:  []string{"user:admin", "lang:go"},
			expected: false,
		},
		"completely different": {
			got:      []string{"lang:go"},
			desired:  []string{"user:admin"},
			expected: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.expected, commands.MatchFilters(test.got, test.desired))
		})
	}
}

func TestValidateAck(t *testing.T) {
	tests := map[string]struct {
		input   string
		wantErr bool
	}{
		"valid lowercase": {
			input:   "yes",
			wantErr: false,
		},
		"valid uppercase": {
			input:   "NO",
			wantErr: false,
		},
		"maybe": {
			input:   "maybe",
			wantErr: true,
		},
		"empty": {
			input:   "",
			wantErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := commands.ValidateAck(test.input)
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateTags(t *testing.T) {
	tests := map[string]struct {
		input   string
		wantErr bool
	}{
		"valid tags": {
			input:   "go dev",
			wantErr: false,
		},
		"empty": {
			input:   "",
			wantErr: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := commands.ValidateTags(test.input)
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateFilters(t *testing.T) {
	tests := map[string]struct {
		input   string
		wantErr bool
	}{
		"valid": {
			input:   "user:admin repo:go",
			wantErr: false,
		},
		"missing semicolon": {
			input:   "useradmin repo:go",
			wantErr: true,
		},
		"single bad format": {
			input:   "badformat",
			wantErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := commands.ValidateFilters(test.input)
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
