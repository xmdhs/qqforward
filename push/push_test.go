// +build no

package push

import (
	"strings"
	"testing"
)

func TestPush(t *testing.T) {
	b, c, err := Downloadimg(`https://attachment.mcbbs.net/forum/202011/12/192731zhnz9hvrctn7zsc1.jpg`, 3)
	if err != nil {
		t.Fatal(err)
	}
	l := strings.Split(c, "/")
	if len(l) < 2 {
		t.Fail()
	}
	by, c, err := PostFile("img."+l[1], b, "test", "@mcbbsimg")
	if err != nil {
		t.Fatal(err)
	}
	Push(by.Bytes(), c, 3)
}
