// +build no

package push

import (
	"fmt"
	"testing"
)

func Test_httpGet(t *testing.T) {
	b, _, err := httpGet(`https://www.mcbbs.net/template/mcbbs/image/logo_sc.png`)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(b))
}
