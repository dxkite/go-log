package log

import (
	"fmt"
	"io"
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	w := io.MultiWriter(NewJsonWriter(os.Stdout))
	log := New(w, true)

	fmt.Println("\njson logger")
	if err := log.Output(1, "default", Linfo, "information\n"); err != nil {
		t.Error(err)
	}

	log.Println(Ldebug, Application("user"), "user info", "some", 1, "@", "11")
	log.Println(Application("user"), Lwarn, "user info", "some", 1, "@", "11")
	log.Println(Lerror, "user info", "some", 1, "@", "11")
	log.Warn("user info", "some", 1, "@", "11")
	log.Warn(Application("user"), "user info", "some", 1, "@", "11")
	fmt.Println("\ntext logger")
	Warn("user info", "some", 1, "@", "11")
	Error(Application("user"), "user info", "some", 1, "@", "11")
	_ = Output(1, "app", Linfo, "message\n")
}
