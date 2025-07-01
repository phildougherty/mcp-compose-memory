package database

import (
    "database/sql"
    "fmt"
    "log"
    "time"

    _ "github.com/lib/pq"
)

func NewConnection(dbURL string) (*sql.DB, error) {
    log.Printf("Database connection string: %s", dbURL)

    // Wait for database to be available
    if err := waitForDatabase(dbURL, 60); err != nil {
        return nil, fmt.Errorf("database connection timeout: %w", err)
    }

    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    // Configure connection pool
    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(time.Hour)

    // Test connection
    if err := db.Ping(); err != nil {
        db.Close()
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    log.Println("Database connection established successfully")
    return db, nil
}

func waitForDatabase(dbURL string, maxRetries int) error {
    log.Println("Waiting for database to become available...")

    for i := 0; i < maxRetries; i++ {
        db, err := sql.Open("postgres", dbURL)
        if err != nil {
            log.Printf("Database connection attempt %d/%d failed: %v", i+1, maxRetries, err)
            time.Sleep(3 * time.Second)
            continue
        }

        err = db.Ping()
        db.Close()

        if err == nil {
            log.Println("Database connection established successfully")
            return nil
        }

        log.Printf("Database ping attempt %d/%d failed: %v", i+1, maxRetries, err)

        if i == maxRetries-1 {
            return fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
        }

        time.Sleep(3 * time.Second)
    }

    return fmt.Errorf("database connection timeout")
}
