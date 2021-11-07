package cmd

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNewRootCmd(t *testing.T) {
	cmd := NewRootCmd()
	assert.Equal(t, "mockery", cmd.Name())
}

func TestGetFilters(t *testing.T) {
	tt := []struct {
		name            string
		interfaceNames  []string
		expectedFilters int
		expectedRegex   bool
	}{
		{
			name:            "Only one interface",
			interfaceNames:  []string{"myInterfaceName"},
			expectedFilters: 1,
			expectedRegex:   false,
		},
		{
			name:            "Only one interface using regex",
			interfaceNames:  []string{".*myInterfaceRegex"},
			expectedFilters: 1,
			expectedRegex:   true,
		},
		{
			name:            "Multiple interfaces without regex",
			interfaceNames:  []string{"multiple", "interface", "names"},
			expectedFilters: 3,
			expectedRegex:   false,
		},
		{
			name:            "Multiple interfaces using regex",
			interfaceNames:  []string{".multiple", "interface", "regex", ".*interface"},
			expectedFilters: 4,
			expectedRegex:   true,
		},
	}

	log := zerolog.New(os.Stderr)

	for _, tc := range tt {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			filters, hasRegex := getFilters(tc.interfaceNames, log)
			assert.Equal(t, len(filters), tc.expectedFilters)
			assert.Equal(t, hasRegex, tc.expectedRegex)
		})
	}
}
