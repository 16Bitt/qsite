package main

import (
	"github.com/16bitt/qsite/pkg/qsite"
)

func main() {
	err := qsite.BootstrapDefault()
	if err != nil {
		panic(err)
	}
}
