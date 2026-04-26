package browser

import (
	"strings"
	"sync"
	"testing"
)

func TestParseNetscapeCookiesValid(t *testing.T) {
	input := ".example.com\tTRUE\t/\tTRUE\t1700000000\tsession_id\tabc123\n" +
		".example.com\tTRUE\t/path\tFALSE\t1700000001\tother_cookie\txyz789\n"

	cookies := ParseNetscapeCookies(input)
	if len(cookies) != 2 {
		t.Fatalf("expected 2 cookies, got %d", len(cookies))
	}

	c := cookies[0]
	if c.Domain != ".example.com" {
		t.Errorf("expected domain '.example.com', got %q", c.Domain)
	}
	if c.Path != "/" {
		t.Errorf("expected path '/', got %q", c.Path)
	}
	if !c.Secure {
		t.Error("expected Secure true")
	}
	if c.Expires != 1700000000 {
		t.Errorf("expected expires 1700000000, got %f", c.Expires)
	}
	if c.Name != "session_id" {
		t.Errorf("expected name 'session_id', got %q", c.Name)
	}
	if c.Value != "abc123" {
		t.Errorf("expected value 'abc123', got %q", c.Value)
	}

	c2 := cookies[1]
	if c2.Secure {
		t.Error("expected Secure false for second cookie")
	}
	if c2.Name != "other_cookie" {
		t.Errorf("expected name 'other_cookie', got %q", c2.Name)
	}
}

func TestParseNetscapeCookiesSkipsComments(t *testing.T) {
	input := "# Netscape HTTP Cookie File\n" +
		"# This is a comment\n" +
		".example.com\tTRUE\t/\tTRUE\t1700000000\tname\tvalue\n"

	cookies := ParseNetscapeCookies(input)
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie (comments skipped), got %d", len(cookies))
	}
	if cookies[0].Name != "name" {
		t.Errorf("expected name 'name', got %q", cookies[0].Name)
	}
}

func TestParseNetscapeCookiesSkipsEmptyLines(t *testing.T) {
	input := "\n\n.example.com\tTRUE\t/\tTRUE\t1700000000\tname\tvalue\n\n\n"

	cookies := ParseNetscapeCookies(input)
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie (empty lines skipped), got %d", len(cookies))
	}
}

func TestParseNetscapeCookiesReturnsEmptyForInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"only comments", "# comment\n# another"},
		{"too few fields", "domain\tpath\tvalue\n"},
		{"garbage", "this is not a cookie format"},
		{"only whitespace", "   \n   \n   "},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cookies := ParseNetscapeCookies(tc.input)
			if len(cookies) != 0 {
				t.Errorf("expected 0 cookies, got %d", len(cookies))
			}
		})
	}
}

func TestParseNetscapeCookiesHandlesMixedInput(t *testing.T) {
	input := "# Header\n" +
		"\n" +
		".valid.com\tTRUE\t/\tTRUE\t0\tcookie1\tval1\n" +
		"bad line\n" +
		".valid2.com\tFALSE\t/foo\tFALSE\t999\tcookie2\tval2\n"

	cookies := ParseNetscapeCookies(input)
	if len(cookies) != 2 {
		t.Fatalf("expected 2 cookies, got %d", len(cookies))
	}
}

func TestParseNetscapeCookiesZeroExpiry(t *testing.T) {
	input := ".example.com\tTRUE\t/\tTRUE\t0\tsession\tval\n"

	cookies := ParseNetscapeCookies(input)
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].Expires != 0 {
		t.Errorf("expected expires 0, got %f", cookies[0].Expires)
	}
}

func TestParseNetscapeCookiesSecureCaseInsensitive(t *testing.T) {
	input := ".example.com\tTRUE\t/\ttrue\t0\tname\tvalue\n"
	cookies := ParseNetscapeCookies(input)
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if !cookies[0].Secure {
		t.Error("expected Secure true (case insensitive)")
	}
}

func TestParseNetscapeCookiesWindowsLineEndings(t *testing.T) {
	input := ".example.com\tTRUE\t/\tTRUE\t0\tname1\tval1\r\n" +
		".example.com\tTRUE\t/\tFALSE\t0\tname2\tval2\r\n"

	cookies := ParseNetscapeCookies(input)
	// The split is on \n, so \r may remain in value. The function trims lines.
	if len(cookies) != 2 {
		t.Fatalf("expected 2 cookies, got %d", len(cookies))
	}
	// Values may have trailing \r due to TrimSpace on the line.
	if strings.TrimSpace(cookies[0].Name) != "name1" {
		t.Errorf("expected name 'name1', got %q", cookies[0].Name)
	}
}

func TestGetInstanceReturnsSingleton(t *testing.T) {
	// Reset the singleton for test isolation.
	once = sync.Once{}
	instance = nil
	defer func() {
		once = sync.Once{}
		instance = nil
	}()

	m1 := GetInstance()
	m2 := GetInstance()
	if m1 != m2 {
		t.Error("GetInstance() returned different pointers")
	}
}

func TestNewManagerReturnsSameAsGetInstance(t *testing.T) {
	once = sync.Once{}
	instance = nil
	defer func() {
		once = sync.Once{}
		instance = nil
	}()

	m1 := GetInstance()
	m2 := NewManager()
	if m1 != m2 {
		t.Error("NewManager() returned different pointer than GetInstance()")
	}
}

func TestManagerStartsInStatusStopped(t *testing.T) {
	once = sync.Once{}
	instance = nil
	defer func() {
		once = sync.Once{}
		instance = nil
	}()

	m := GetInstance()
	status, errMsg := m.GetStatus()
	if status != StatusStopped {
		t.Errorf("expected StatusStopped, got %s", status)
	}
	if errMsg != "" {
		t.Errorf("expected empty error message, got %q", errMsg)
	}
}

func TestManagerIsRunningReturnsFalseInitially(t *testing.T) {
	once = sync.Once{}
	instance = nil
	defer func() {
		once = sync.Once{}
		instance = nil
	}()

	m := GetInstance()
	if m.IsRunning() {
		t.Error("expected IsRunning() false for new manager")
	}
}

func TestStatusConstants(t *testing.T) {
	tests := []struct {
		status   Status
		expected string
	}{
		{StatusStopped, "stopped"},
		{StatusStarting, "starting"},
		{StatusRunning, "running"},
		{StatusError, "error"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			if string(tc.status) != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, string(tc.status))
			}
		})
	}
}

func TestCDPCookieStruct(t *testing.T) {
	c := CDPCookie{
		Name:     "session",
		Value:    "abc",
		Domain:   ".example.com",
		Path:     "/",
		Expires:  1700000000,
		HTTPOnly: true,
		Secure:   true,
	}

	if c.Name != "session" {
		t.Errorf("expected name 'session', got %q", c.Name)
	}
	if !c.HTTPOnly {
		t.Error("expected HTTPOnly true")
	}
	if !c.Secure {
		t.Error("expected Secure true")
	}
}
