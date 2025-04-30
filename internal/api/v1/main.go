package v1

//go:generate swag init --parseDependency --parseInternal --outputTypes json

// This exists only to make swaggo happy.
// @title Web Cache API
// @version 1.0
// @description API for managing web cache keys
// @BasePath /api/v1
func main() {
	SetupRoutes(nil)
}
