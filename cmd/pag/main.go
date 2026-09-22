package main

import (
	"fmt"
	"github.com/Ploos-AS/Ploos-Application-Gateway/internal/version"
)

func main() {
	fmt.Printf("PAG — Ploos Application Gateway %s\n", version.Version)
}
