package features

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"

	"github.com/PDOK/gokoala/internal/ogc/features/domain" // SRID is used
	"github.com/stretchr/testify/assert"
)

// Helps to mock exec.Command. Inspired by https://www.joeshaw.org/testing-with-os-exec-and-testmain/
func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Fprint(os.Stdout, os.Getenv("STDOUT"))
	i, _ := strconv.Atoi(os.Getenv("EXIT_CODE"))
	os.Exit(i)
}

func TestShouldSwapXY(t *testing.T) {
	// Save original functions and restore them after all tests in this function are done
	originalCmdFunc := execCommand
	originalLookPathFunc := execLookPath
	defer func() {
		execCommand = originalCmdFunc
		execLookPath = originalLookPathFunc
	}()

	type args struct {
		name            string
		srid            domain.SRID
		mockLookPathErr error
		mockCmdOutput   string
		mockCmdExitCode int
		expectedSwap    bool
		expectedError   bool
		expectedErrMsg  string
	}
	tests := []args{
		{
			name:            "should swap - first axis direction is east",
			srid:            domain.SRID(28992),
			mockCmdOutput:   `{"coordinate_system":{"axis":[{"direction":"east"}, {"direction":"north"}]}}`,
			mockCmdExitCode: 0,
			expectedSwap:    false,
			expectedError:   false,
		},
		{
			name:            "should not swap - first axis direction is north",
			srid:            domain.SRID(4258),
			mockCmdOutput:   `{"coordinate_system":{"axis":[{"direction":"north"}, {"direction":"east"}]}}`,
			mockCmdExitCode: 0,
			expectedSwap:    true,
			expectedError:   false,
		},
		{
			name:            "error - projinfo not found (LookPath fails)",
			srid:            domain.SRID(1000),
			mockLookPathErr: errors.New("simulated LookPath error: command not found"),
			expectedError:   true,
			expectedErrMsg:  "projinfo command not found in PATH",
		},
		{
			name:            "error - projinfo command execution fails (non-zero exit)",
			srid:            domain.SRID(2000),
			mockCmdOutput:   "error from command",
			mockCmdExitCode: 1, // Simulate command failure
			expectedError:   true,
			expectedErrMsg:  "failed to execute projinfo command",
		},
		{
			name:            "error - projinfo output is invalid JSON",
			srid:            domain.SRID(3000),
			mockCmdOutput:   "this is not valid json",
			mockCmdExitCode: 0,
			expectedError:   true,
			expectedErrMsg:  "failed to parse projinfo output",
		},
		{
			name:            "error - projinfo output JSON lacks axis",
			srid:            domain.SRID(4000),
			mockCmdOutput:   `{"coordinate_system":{}}`,
			mockCmdExitCode: 0,
			expectedError:   true,
			expectedErrMsg:  "invalid projinfo output: axis not found",
		},
		{
			name:            "error - projinfo output JSON has empty axis",
			srid:            domain.SRID(5000),
			mockCmdOutput:   `{"coordinate_system":{"axis":[]}}`,
			mockCmdExitCode: 0,
			expectedError:   true,
			expectedErrMsg:  "invalid projinfo output: axis not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks for execLookPath and execCommand for this specific test case
			execLookPath = func(file string) (string, error) {
				if file == projInfoTool {
					if tc.mockLookPathErr != nil {
						return "", tc.mockLookPathErr
					}
					return "/test/projinfo", nil
				}
				return originalLookPathFunc(file)
			}

			execCommand = func(name string, arg ...string) *exec.Cmd {
				// Check if the command is the one we intend to mock by name.
				// It could be "projinfo" or the dummy path "/test/projinfo".
				if name == projInfoTool || name == "/test/projinfo" {
					cs := []string{"-test.run=TestHelperProcess", "--", name}
					cmd := exec.Command(os.Args[0], cs...) //nolint:gosec    // Use test binary itself
					cmd.Env = []string{
						"GO_WANT_HELPER_PROCESS=1",
						"STDOUT=" + tc.mockCmdOutput,
						"EXIT_CODE=" + strconv.Itoa(tc.mockCmdExitCode),
					}
					return cmd
				}
				return originalCmdFunc(name, arg...)
			}

			swap, err := ShouldSwapXY(tc.srid)

			if tc.expectedError {
				assert.Contains(t, err.Error(), tc.expectedErrMsg, "Test: %s. Error message mismatch for SRID %d", tc.name, tc.srid)
			} else {
				assert.Equal(t, tc.expectedSwap, swap, "Test: %s. Swap value mismatch for SRID %d", tc.name, tc.srid)
			}
		})
	}
}
