package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/dukerupert/faa-aircraft-search/internal/database"
	"github.com/dukerupert/faa-aircraft-search/internal/migration"
)

func main() {
	var (
		action   = flag.String("action", "", "Action to perform: import, clear, count")
		filePath = flag.String("file", "aircraft_data.xlsx", "Path to Excel file for import")
	)
	flag.Parse()

	if *action == "" {
		fmt.Println("Usage:")
		fmt.Println("  go run cmd/migrate/main.go -action=import [-file=path/to/file.xlsx]")
		fmt.Println("  go run cmd/migrate/main.go -action=clear")
		fmt.Println("  go run cmd/migrate/main.go -action=count")
		os.Exit(1)
	}

	ctx := context.Background()

	// Initialize database connection
	pool, err := database.InitDatabase(ctx)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.Close(pool)

	switch *action {
	case "import":
		err = migration.MigrateFromExcel(ctx, pool, *filePath)
		if err != nil {
			log.Fatal("Migration failed:", err)
		}
		fmt.Println("Migration completed successfully!")

	case "clear":
		err = migration.ClearData(ctx, pool)
		if err != nil {
			log.Fatal("Failed to clear data:", err)
		}
		fmt.Println("Data cleared successfully!")

	case "count":
		count, err := migration.GetRecordCount(ctx, pool)
		if err != nil {
			log.Fatal("Failed to get record count:", err)
		}
		fmt.Printf("Aircraft records in database: %d\n", count)

	default:
		log.Fatal("Unknown action:", *action)
	}
}