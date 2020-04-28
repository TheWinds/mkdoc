package mkdoc

import (
	"math/rand"
	"time"
)

const (
	Version = "0.8"
)

func init() {
	rand.Seed(time.Now().Unix())
}
