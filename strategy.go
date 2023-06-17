package gcache

import (
	"fmt"
	"time"
)

type GCMode uint8

const (
	Never GCMode = iota
	Auto
	Lazy
)

type Strategy struct {
	Mode GCMode
	Life time.Duration
}

type IStrategy interface {
	checkValid() error
}

func (st *Strategy) checkValid() error {
	b := st.Mode == Never && st.Life == 0 || st.Mode != Never && st.Life > 0
	if !b {
		return fmt.Errorf("strategy is invalid")
	}
	return nil
}
