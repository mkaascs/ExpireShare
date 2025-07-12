package main

import (
	"expire-share/internal/config"
	"fmt"
)

func main() {
	cfg := config.MustLoad()
	fmt.Println(cfg)
}
