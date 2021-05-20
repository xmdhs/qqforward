package push

import (
	"testing"
)

func TestSplit(t *testing.T) {
	l := Split("123", 200)
	if l[0] != "123" {
		t.Fatal(l)
	}
}
