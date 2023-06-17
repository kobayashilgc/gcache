package gcache

import (
	"fmt"
	"testing"
	"time"
)

func TestDemo(t *testing.T) {
	gcache.Add("k1", "v1", time.Duration(2*time.Minute))
	v1 := gcache.Get("k1")
	fmt.Println(v1.(string))
}
