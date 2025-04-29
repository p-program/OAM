package analyze

import (
	"fmt"
	"testing"
)

func TestGofileGPT(t *testing.T) {
	fmt.Println(GofileGPT("test.txt"))
	// 0 分
}

func TestGofileMeta(t *testing.T) {
	fmt.Println(GofileMeta("test.txt"))
	// 0.5 分

}
