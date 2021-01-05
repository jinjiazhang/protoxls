package main

import (
	"fmt"
	"os"
)

func main() {
	file := "scheme.proto"
	if len(os.Args) >= 2 {
		file = os.Args[1]
	}

	scheme, err := ParseScheme(file)
	if err != nil {
		fmt.Println("parse scheme fail, err:", err)
		return
	}

	for _, meta := range scheme.GetMessageTypes() {
		LoadXlsStore(meta)
	}
}
