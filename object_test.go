package mkdoc

import (
	"encoding/json"
	"fmt"
	"testing"
)

func Test_createRootObject(t *testing.T) {
	root, ref, err := createRootObject("[]string")
	if err != nil {
		fmt.Println(err)
		return
	}
	b, _ := json.MarshalIndent(root, "", "\t")
	fmt.Println(string(b))
	b, _ = json.MarshalIndent(ref, "", "\t")
	fmt.Println(string(b))
}
