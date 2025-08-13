package cli

import (
	"errors"
	"testing"
)

func TestCheckRebootStatus(t *testing.T) {
	tests := []struct {
		name           string
		job            HostJob
		mockUptime     int64
		mockUptimeErr  error
		expectedResult bool
		expectedError  string
	}{
		{
			name: "successful reboot - uptime decreased",
			job: HostJob{
				Result: "initiated reboot, uptime: 1000",
			},
			mockUptime:     500,
			expectedResult: true,
			expectedError:  "",
		},
		{
			name: "failed reboot - uptime still high after max duration",
			job: HostJob{
				Result: "initiated reboot, uptime: 1000",
			},
			mockUptime:     1400, // 1000 + 400 > maxRebootDuration (300)
			expectedResult: false,
			expectedError:  "system is still running after the reboot was initiated",
		},
		{
			name: "reboot in progress - within max duration",
			job: HostJob{
				Result: "initiated reboot, uptime: 1000",
			},
			mockUptime:     1200, // 1000 + 200 < maxRebootDuration (300)
			expectedResult: false,
			expectedError:  "",
		},
		{
			name: "invalid status format - missing prefix",
			job: HostJob{
				Result: "some other status",
			},
			expectedResult: false,
			expectedError:  "status data is not in the expected format",
		},
		{
			name: "invalid status format - wrong number of parts",
			job: HostJob{
				Result: "initiated reboot, uptime: 1000, extra",
			},
			expectedResult: false,
			expectedError:  "status data is not in the expected format",
		},
		{
			name: "invalid status format - non-numeric uptime",
			job: HostJob{
				Result: "initiated reboot, uptime: abc",
			},
			expectedResult: false,
			expectedError:  "status data is not in the expected format",
		},
		{
			name: "error getting current uptime",
			job: HostJob{
				Result: "initiated reboot, uptime: 1000",
			},
			mockUptimeErr:  errors.New("failed to get uptime"),
			expectedResult: false,
			expectedError:  "error getting uptime: failed to get uptime",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the getUptimeFunc for this test
			originalGetUptimeFunc := getUptime
			getUptime = func() (int64, error) {
				return tt.mockUptime, tt.mockUptimeErr
			}
			// Restore the original function after the test
			defer func() {
				getUptime = originalGetUptimeFunc
			}()

			result, err := checkRebootStatus(tt.job)

			if result != tt.expectedResult {
				t.Errorf("checkRebootStatus() result = %v, want %v", result, tt.expectedResult)
			}

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("checkRebootStatus() error = %v, want nil", err)
				}
			} else {
				if err == nil {
					t.Errorf("checkRebootStatus() error = nil, want %v", tt.expectedError)
				} else if err.Error() != tt.expectedError {
					t.Errorf("checkRebootStatus() error = %v, want %v", err.Error(), tt.expectedError)
				}
			}
		})
	}
}
