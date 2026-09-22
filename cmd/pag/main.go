package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Ploos-AS/Ploos-Application-Gateway/internal/config"
	"github.com/Ploos-AS/Ploos-Application-Gateway/internal/gateway"
	"github.com/Ploos-AS/Ploos-Application-Gateway/internal/version"
)

func main() {
	configPath := flag.String("config", "", "path to PAG JSON configuration")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.Version)
		return
	}
	if *configPath == "" {
		fmt.Fprintf(os.Stderr, "PAG — Ploos Application Gateway %s\n", version.Version)
		fmt.Fprintln(os.Stderr, "usage: pag -config /etc/pag/pag.json")
		os.Exit(2)
	}

	cfg, err := config.Load(*configPath)
	if err != nil { log.Fatalf("configuration rejected: %v", err) }

	registry, err := gateway.DiscoverAll(context.Background(), cfg.Gateways)
	if err != nil { log.Fatalf("gateway discovery failed (fail closed): %v", err) }

	fmt.Printf("PAG %s: configuration valid; %d gateway(s) healthy\n", version.Version, len(registry.IDs()))
	for _, id := range registry.IDs() { fmt.Printf("gateway: %s\n", id) }
}
