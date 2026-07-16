package logze

import (
	"bytes"
	"strings"
	"testing"
)

// TestFatalFlushesDiode verifies that Fatal* methods write their message
// synchronously past the diode writer before exiting, so it is never lost.
func TestFatalFlushesDiode(t *testing.T) {
	origExit := osExit
	defer func() { osExit = origExit }()

	tests := []struct {
		name string
		call func(l Logger)
		want string
	}{
		{name: "Fatal", call: func(l Logger) { l.Fatal("fatal plain") }, want: "fatal plain"},
		{name: "Fatalf", call: func(l Logger) { l.Fatalf("fatal formatted %d", 42) }, want: "fatal formatted 42"},
		{name: "Fatalln", call: func(l Logger) { l.Fatalln("fatal line") }, want: "fatal line"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			exitCode := -1
			osExit = func(code int) { exitCode = code }

			var b bytes.Buffer
			l := New(NewConfig(&b)) // diode enabled by default

			tc.call(l)

			if exitCode != 1 {
				t.Errorf("expected exit code 1, got %d", exitCode)
			}
			// The fatal entry is written synchronously past the diode,
			// so it must already be in the buffer.
			if !strings.Contains(b.String(), tc.want) {
				t.Errorf("expected %q written before exit, got %q", tc.want, b.String())
			}
		})
	}
}
