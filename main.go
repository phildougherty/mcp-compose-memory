package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "mcp-compose-memory/internal/database"
    "mcp-compose-memory/internal/handlers"
    "mcp-compose-memory/internal/knowledge"
    "net/http"
    "os"
    "os/signal"
    "strconv"
    "syscall"
    "time"

    "github.com/gorilla/mux"
    "github.com/spf13/cobra"
)

var (
    version = "0.7.0"
    host    string
    port    int
    dbURL   string
)

func main() {
    var rootCmd = &cobra.Command{
        Use:     "mcp-compose-memory",
        Short:   "MCP Memory Server with PostgreSQL backend",
        Version: version,
        RunE:    startServer,
    }

    rootCmd.Flags().StringVar(&host, "host", "0.0.0.0", "Host to bind to")
    rootCmd.Flags().IntVar(&port, "port", 3001, "Port to bind to")
    rootCmd.Flags().StringVar(&dbURL, "db-url", "", "Database connection URL")

    if err := rootCmd.Execute(); err != nil {
        log.Fatal(err)
    }
}

func startServer(cmd *cobra.Command, args []string) error {
    // Set database URL from environment if not provided
    if dbURL == "" {
        dbURL = os.Getenv("DATABASE_URL")
        if dbURL == "" {
            dbURL = "postgresql://postgres:password@localhost:5432/memory_graph?sslmode=disable"
        }
    }

    log.Printf("Starting MCP Memory Server v%s", version)
    log.Printf("Database URL: %s", dbURL)
    log.Printf("Binding to %s:%d", host, port)

    // Wait for database to be ready in production
    if os.Getenv("NODE_ENV") == "production" {
        log.Println("Waiting 5 seconds for database to be ready...")
        time.Sleep(5 * time.Second)
    }

    // Initialize database connection
    db, err := database.NewConnection(dbURL)
    if err != nil {
        return fmt.Errorf("failed to connect to database: %w", err)
    }
    defer db.Close()

    // Run migrations
    if err := database.RunMigrations(db); err != nil {
        return fmt.Errorf("failed to run migrations: %w", err)
    }

    // Create knowledge graph manager
    manager := knowledge.NewManager(db)

    // Create MCP handler
    mcpHandler := handlers.NewMCPHandler(manager)

    // Setup HTTP server
    router := mux.NewRouter()
    
    // Health check endpoint
    router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/plain")
        w.WriteHead(http.StatusOK)
        w.Write([]byte("OK"))
    }).Methods("GET")

    // MCP JSON-RPC endpoint
    router.HandleFunc("/", mcpHandler.HandleMCPRequest).Methods("POST", "OPTIONS")

    // Enable CORS
    router.Use(corsMiddleware)

    server := &http.Server{
        Addr:         fmt.Sprintf("%s:%d", host, port),
        Handler:      router,
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // Graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go func() {
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
        <-sigChan

        log.Println("Shutting down server...")
        shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer shutdownCancel()

        if err := server.Shutdown(shutdownCtx); err != nil {
            log.Printf("Server shutdown error: %v", err)
        }
        cancel()
    }()

    log.Printf("MCP Memory Server running on http://%s:%d", host, port)

    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        return fmt.Errorf("server failed: %w", err)
    }

    <-ctx.Done()
    log.Println("Server stopped")
    return nil
}

func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        if r.Method == "OPTIONS" {
            w.WriteHead(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}
