package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/glebarez/go-sqlite"
)

// gen-db builds an independent ops-dump.db from manifest/sql/ops-dump.sql using the
// pure-Go glebarez/modernc sqlite driver (no CGO). The resulting database is used
// as the source for `gf gen dao` code generation.
//
// Usage:
//
//	go run ./hack/gen-db [sqlPath] [dbPath]
func main() {
	sqlPath := "manifest/sql/ops-dump.sql"
	dbPath := "ops-dump.db"
	if len(os.Args) > 1 {
		sqlPath = os.Args[1]
	}
	if len(os.Args) > 2 {
		dbPath = os.Args[2]
	}

	_ = os.Remove(dbPath)

	raw, err := os.ReadFile(sqlPath)
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	for _, stmt := range strings.Split(string(raw), ";") {
		s := strings.TrimSpace(stmt)
		if s == "" {
			continue
		}
		if _, err := db.Exec(s); err != nil {
			panic(fmt.Sprintf("exec failed: %v\nsql: %s", err, s))
		}
	}
	fmt.Printf("generated %s from %s\n", dbPath, sqlPath)
}
