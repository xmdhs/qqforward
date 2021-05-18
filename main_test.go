package main

import (
	"fmt"
	"testing"
)

func Test_cqcode(t *testing.T) {
	l := cqcode(`test[CQ:face,id=123]测试[CQ:face,id=123]aaa[CQ:face,id=123]啥`)
	fmt.Println(l)
	s := `test[CQ:face,id=123]测试[CQ:face,id=123]aaa[CQ:face,id=123]啥`
	fmt.Println(s[20:26])
}
