package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/empfaze/golang_microservices/models"
	"github.com/empfaze/golang_microservices/repository/order"
	"github.com/google/uuid"
)

type Order struct {
	Repository *order.RedisRepository
}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CustomerID uuid.UUID         `json:"customer_id"`
		LineItems  []models.LineItem `json:"line_items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	now := time.Now().UTC()

	order := models.Order{
		OrderID:    rand.Uint64(),
		CustomerID: body.CustomerID,
		LineItems:  body.LineItems,
		CreatedAt:  &now,
	}

	err := o.Repository.Create(r.Context(), order)
	if err != nil {
		fmt.Println("Failed to create an order: %w", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	result, err := json.Marshal(order)
	if err != nil {
		fmt.Println("Failed marshall an order: %w", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Write(result)
	w.WriteHeader(http.StatusCreated)
}

func (o *Order) GetAll(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get all orders")
}

func (o *Order) GetById(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get an order by ID")
}

func (o *Order) UpdateByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Update an order by ID")
}

func (o *Order) DeleteByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete an order by ID")
}
