// Demo HTTP server for galick testing
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// HealthResponse is the response for the health endpoint
type HealthResponse struct {
	Status  string    `json:"status"`
	Version string    `json:"version"`
	Time    time.Time `json:"time"`
}

// UserResponse is the response for the users endpoint
type UserResponse struct {
	Users []User `json:"users"`
	Count int    `json:"count"`
}

// User represents a user in the system
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}

// ProductResponse is the response for the products endpoint
type ProductResponse struct {
	Products []Product `json:"products"`
	Count    int       `json:"count"`
}

// Product represents a product in the system
type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	InStock     bool    `json:"in_stock"`
}

// OrderResponse is the response for creating an order
type OrderResponse struct {
	OrderID     int       `json:"order_id"`
	UserID      int       `json:"user_id"`
	ProductIDs  []int     `json:"product_ids"`
	TotalAmount float64   `json:"total_amount"`
	CreatedAt   time.Time `json:"created_at"`
}

// SimulateLatency adds random latency to the response
func simulateLatency() {
	// Randomly sleep between 10-200ms to simulate realistic API latency
	time.Sleep(time.Duration(10+rand.Intn(190)) * time.Millisecond)
}

// Occasionally fail the request
func shouldFail() bool {
	// 5% chance of failure
	return rand.Float32() < 0.05
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Health endpoint
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		simulateLatency()

		if shouldFail() {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"error": "Internal server error"}`)
			return
		}

		response := HealthResponse{
			Status:  "OK",
			Version: "1.0.0",
			Time:    time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Users endpoint
	http.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		simulateLatency()

		if shouldFail() {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"error": "Internal server error"}`)
			return
		}

		users := []User{
			{ID: 1, Name: "Alice Smith", Email: "alice@example.com", IsActive: true},
			{ID: 2, Name: "Bob Johnson", Email: "bob@example.com", IsActive: true},
			{ID: 3, Name: "Charlie Brown", Email: "charlie@example.com", IsActive: false},
			{ID: 4, Name: "Diana Prince", Email: "diana@example.com", IsActive: true},
			{ID: 5, Name: "Ethan Hunt", Email: "ethan@example.com", IsActive: true},
		}

		response := UserResponse{
			Users: users,
			Count: len(users),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Products endpoint
	http.HandleFunc("/api/products", func(w http.ResponseWriter, r *http.Request) {
		simulateLatency()

		if shouldFail() {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"error": "Internal server error"}`)
			return
		}

		products := []Product{
			{ID: 1, Name: "Laptop", Price: 999.99, Description: "High-performance laptop", InStock: true},
			{ID: 2, Name: "Smartphone", Price: 699.99, Description: "Latest smartphone model", InStock: true},
			{ID: 3, Name: "Headphones", Price: 149.99, Description: "Noise-cancelling headphones", InStock: false},
			{ID: 4, Name: "Tablet", Price: 349.99, Description: "10-inch tablet", InStock: true},
			{ID: 5, Name: "Smartwatch", Price: 249.99, Description: "Fitness tracking smartwatch", InStock: true},
		}

		response := ProductResponse{
			Products: products,
			Count:    len(products),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Orders endpoint (POST only)
	http.HandleFunc("/api/orders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprint(w, `{"error": "Method not allowed"}`)
			return
		}

		simulateLatency()

		if shouldFail() {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"error": "Internal server error"}`)
			return
		}

		response := OrderResponse{
			OrderID:     rand.Intn(1000) + 1,
			UserID:      rand.Intn(5) + 1,
			ProductIDs:  []int{rand.Intn(5) + 1, rand.Intn(5) + 1},
			TotalAmount: float64(rand.Intn(100000)) / 100.0,
			CreatedAt:   time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	port := 8080
	log.Printf("Starting demo server on port %d...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
