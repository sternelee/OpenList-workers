package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/OpenListTeam/OpenList-workers/workers/auth"
	"github.com/OpenListTeam/OpenList-workers/workers/db"
	"github.com/OpenListTeam/OpenList-workers/workers/drivers"
	_ "github.com/OpenListTeam/OpenList-workers/workers/drivers/virtual" // 注册虚拟驱动
	"github.com/OpenListTeam/OpenList-workers/workers/handlers"
	"github.com/OpenListTeam/OpenList-workers/workers/routes"
	"github.com/syumai/workers"
	"github.com/syumai/workers/cloudflare"
)

func main() {
	// Get D1 database binding
	d1 := cloudflare.NewD1Database("DB")
	dbClient := db.NewD1Client(d1)

	// Create repositories using the database connection
	repos := db.NewRepositories(dbClient.DB)

	// Create driver service
	driverService := drivers.NewDriverService(repos)

	// Initialize driver service
	ctx := context.Background()
	if err := driverService.Initialize(ctx); err != nil {
		// Log error but continue - workers environment may not have storage configured yet
		workers.Logger().Printf("Warning: Failed to initialize driver service: %v", err)
	}

	// Get JWT secret from environment or use default
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtAuth := auth.NewJWTAuth(jwtSecret)

	// Create auth handler using repositories
	authHandler := handlers.NewAuthHandler(repos)

	// Create HTTP mux with full route setup
	mux := routes.SetupRoutes(repos, driverService)

	// Add ping endpoint for health checks
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "pong", "service": "openlist-workers", "drivers": ` + 
			getDriverCountJSON(driverService) + `}`))
	})

	// CORS middleware
	corsHandler := corsMiddleware(mux)

	// Serve the application
	workers.Serve(corsHandler)
}

// CORS middleware to handle cross-origin requests
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Helper function to get driver count as JSON string
func getDriverCountJSON(driverService *drivers.DriverService) string {
	driverNames := drivers.GetDriverNames()
	return `{"available": ` + string(rune(len(driverNames))) + `, "types": ["Virtual"]}`
}
