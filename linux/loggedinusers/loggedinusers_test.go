package linux_loggedinusers

import (
	"testing"
)

const testCase = `ewillems pts/0 2023-10-01 10:00 (host1)
ewillems pts/1 2023-10-01 10:05 (host2)
ewillems pts/2 2023-10-01 10:10 (host3)
`

const testCaseConsole = `ewillems tty1 2023-10-01 10:00
ewillems tty2 2023-10-01 10:05
ewillems pts/1 2023-10-01 10:05 (host2)
ewillems pts/2 2023-10-01 10:10 (host3)
`

const testCaseNoUsers = ``

const testCaseInvalid = `ewillems`

func TestParseLoggedInUsers(t *testing.T) {
	users := parseLoggedInUsers(testCase)
	if len(users) != 3 {
		t.Errorf("expected 3 users, got %d", len(users))
	}
}
func TestParseLoggedInUsersConsole(t *testing.T) {
	users := parseLoggedInUsers(testCaseConsole)
	if len(users) != 4 {
		t.Errorf("expected 4 users, got %d", len(users))
	}
}

func TestParseLoggedInUsersNoUsers(t *testing.T) {
	users := parseLoggedInUsers(testCaseNoUsers)
	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}
}

func TestParseLoggedInUsersInvalid(t *testing.T) {
	users := parseLoggedInUsers(testCaseInvalid)
	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}
}
