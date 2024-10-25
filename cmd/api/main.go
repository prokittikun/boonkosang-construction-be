package main

import (
	"boonkosang/internal/adapters/postgres"
	"boonkosang/internal/adapters/rest"
	"boonkosang/internal/infrastructure/database"
	"boonkosang/internal/infrastructure/server"
	"boonkosang/internal/usecase"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println("Warning: No .env file found")
	}

	// Now that env vars are loaded, we can use getEnv
	fmt.Println("Boonkosang API", getEnv("DB_HOST", "beer"))

	// Create a new configuration
	dbConfig := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvAsInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", ""),
		DBName:   getEnv("DB_NAME", "general"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	db, err := database.NewSQLxDB(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.CloseSQLxDB(db)

	app := server.NewFiberServer()

	userRepo := postgres.NewUserRepository(db)
	jwtSecret := getEnv("JWT_SECRET", "your_default_secret")
	jwtExpiration := getEnvAsDuration("JWT_EXPIRATION", 15*time.Minute)
	userUseCase := usecase.NewUserUsecase(userRepo, jwtSecret, jwtExpiration)
	UserHandler := rest.NewUserHandler(userUseCase)
	UserHandler.UserRoutes(app)

	clientRepo := postgres.NewClientRepository(db)
	clientUseCase := usecase.NewClientUsecase(clientRepo)
	ClientHandler := rest.NewClientHandler(clientUseCase)
	ClientHandler.ClientRoutes(app)

	supplierRepo := postgres.NewSupplierRepository(db)
	supplierUseCase := usecase.NewSupplierUsecase(supplierRepo)
	SupplierHandler := rest.NewSupplierHandler(supplierUseCase)
	SupplierHandler.SupplierRoutes(app)

	projectRepo := postgres.NewProjectRepository(db)
	projectUseCase := usecase.NewProjectUsecase(projectRepo, clientRepo)
	ProjectHandler := rest.NewProjectHandler(projectUseCase)
	ProjectHandler.ProjectRoutes(app)

	materialRepo := postgres.NewMaterialRepository(db)
	materialUseCase := usecase.NewMaterialUsecase(materialRepo, supplierRepo)
	MaterialHandler := rest.NewMaterialHandler(materialUseCase)
	MaterialHandler.MaterialRoutes(app)

	port := getEnv("PORT", "8004")
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}
