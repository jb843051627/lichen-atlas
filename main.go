package main

import (
	"flag"
	"log"
	"os"

	"github.com/jb843051627/lichen-atlas/internal/clock"
	"github.com/jb843051627/lichen-atlas/internal/handler"
	"github.com/jb843051627/lichen-atlas/internal/service"
	"github.com/jb843051627/lichen-atlas/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	dbPath := flag.String("db", "data/lichen-atlas.db", "SQLite database path")
	flag.Parse()
	if err := os.MkdirAll("data", 0o755); err != nil {
		log.Fatal(err)
	}
	db, err := store.Open(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := service.NewApplication(db, clock.Real{})
	server := handler.NewServer(app)
	log.Printf("lichen-atlas listening on %s", *addr)
	if err := server.ListenAndServe(*addr); err != nil {
		log.Fatal(err)
	}
}
