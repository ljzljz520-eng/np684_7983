package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"ticketdesk/internal/batch"
	"ticketdesk/internal/httpapi"
	"ticketdesk/internal/store"
)

type config struct {
	database string
	address  string
}

func parseConfig() config {
	base := os.Getenv("TICKETDESK_DATA")
	if base == "" {
		base = ".ticketdesk"
	}
	database := filepath.Join(base, "ticketdesk.db")
	configValue := config{}
	flag.StringVar(&configValue.database, "db", database, "database file")
	flag.StringVar(&configValue.address, "addr", ":8080", "listen address")
	flag.Parse()
	return configValue
}

func run(configValue config) error {
	db, err := store.Open(configValue.database)
	if err != nil {
		return err
	}
	defer db.Close()
	service := batch.NewService(db)
	server := httpapi.New(service)
	log.Printf("ticketdesk listening on %s", configValue.address)
	return http.ListenAndServe(configValue.address, server.Handler())
}

func main() {
	if err := run(parseConfig()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
