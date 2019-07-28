package main

import (
	"testing"
)

func TestCoregoAPIAnalyzer_GetAPIList(t *testing.T) {
	new(CoregoAPIAnalyzer).GetAPIList("corego/service/xyt/api")
}
