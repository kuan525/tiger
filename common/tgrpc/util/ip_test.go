package util

import (
	"fmt"
	"testing"
)

func TestExternaIP(t *testing.T) {
	s := ExternaIP()
	fmt.Println(s)
}
