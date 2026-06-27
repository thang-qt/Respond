package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"respond/internal/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/lib/pq"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	direction := flag.String("direction", "up", "Migration direction: up or down")
	steps := flag.Int("steps", 0, "Number of steps to run (0 = all)")
	reset := flag.Bool("reset", false, "Drop all database objects and re-run all up migrations")
	flag.Parse()

	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	m, err := migrate.New("file://migrations", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("migrate init: %v", err)
	}

	if *reset {
		if err := m.Drop(); err != nil {
			log.Fatalf("migrate reset drop failed: %v", err)
		}
		if err := dropPublicEnumTypes(cfg.DatabaseURL); err != nil {
			log.Fatalf("migrate reset enum cleanup failed: %v", err)
		}
		// Recreate migrate instance after Drop. Some drivers keep
		// state tied to the dropped schema_migrations table.
		if sourceErr, dbErr := m.Close(); sourceErr != nil || dbErr != nil {
			log.Printf("migrate reset close warning: source=%v db=%v", sourceErr, dbErr)
		}
		m, err = migrate.New("file://migrations", cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("migrate reset re-init failed: %v", err)
		}
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate reset up failed: %v", err)
		}
		return
	}

	switch *direction {
	case "up":
		if *steps > 0 {
			err = m.Steps(*steps)
		} else {
			err = m.Up()
		}
	case "down":
		if *steps > 0 {
			err = m.Steps(-*steps)
		} else {
			err = m.Down()
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown direction: %s\n", *direction)
		os.Exit(1)
	}

	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migrate %s failed: %v", *direction, err)
	}
}

func dropPublicEnumTypes(dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT t.typname
		FROM pg_type t
		JOIN pg_namespace n ON n.oid = t.typnamespace
		WHERE t.typtype = 'e'
		  AND n.nspname = 'public'
		ORDER BY t.typname
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, name := range names {
		query := fmt.Sprintf("DROP TYPE IF EXISTS %s.%s CASCADE", pq.QuoteIdentifier("public"), pq.QuoteIdentifier(name))
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}
