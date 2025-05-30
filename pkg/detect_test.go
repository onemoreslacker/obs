package pkg_test

import (
	"testing"

	"github.com/es-debug/backend-academy-2024-go-template/config"
	"github.com/es-debug/backend-academy-2024-go-template/pkg"
	"github.com/stretchr/testify/require"
)

func TestServiceFromURL(t *testing.T) {
	tests := map[string]struct {
		link            string
		expectedService string
		wantErr         bool
	}{
		"detected github": {
			link:            "https://github.com/example/repo",
			expectedService: config.GitHub,
			wantErr:         false,
		},
		"detected stackoverflow": {
			link:            "https://stackoverflow.com/questions/1",
			expectedService: config.StackOverflow,
			wantErr:         false,
		},
		"unknown service": {
			link:            "https://youtube.com",
			expectedService: "",
			wantErr:         true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualService, err := pkg.ServiceFromURL(test.link)

			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedService, actualService)
		})
	}
}
