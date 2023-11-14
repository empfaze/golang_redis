package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/empfaze/golang_redis/models"
	"github.com/empfaze/golang_redis/repositories/order"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Order struct {
	Repository *order.RedisRepository
}

const COMPLETED_STATUS = "completed"
const SHIPPED_STATUS = "shipped"

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
	cursorString := r.URL.Query().Get("cursor")
	if cursorString == "" {
		cursorString = "0"
	}

	cursor, err := strconv.ParseUint(cursorString, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	const size = 50
	result, err := o.Repository.GetAll(r.Context(), order.GetAllPaginationParams{
		Offset: cursor,
		Size:   size,
	})
	if err != nil {
		fmt.Println("Failed to find all orders: %w", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	var response struct {
		Items []models.Order `json:"items"`
		Next  uint64         `json:"next,omitempty"`
	}
	response.Items = result.Orders
	response.Next = result.Cursor

	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Failed to marshall response: %w", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Write(data)
}

func (o *Order) GetById(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	orderID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	result, err := o.Repository.FindByID(r.Context(), orderID)
	if errors.Is(err, order.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)

		return
	} else if err != nil {
		fmt.Println("Failed to find order by id: %w", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		fmt.Println("Failed to marshall an order: %w", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}

func (o *Order) UpdateByID(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	id := chi.URLParam(r, "id")

	orderID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := o.Repository.FindByID(r.Context(), orderID)
	if errors.Is(err, order.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)

		return
	} else if err != nil {
		fmt.Println("failed to find by id:", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	now := time.Now().UTC()

	switch body.Status {
	case SHIPPED_STATUS:
		if result.ShippedAt != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}
		result.ShippedAt = &now
	case COMPLETED_STATUS:
		if result.CompletedAt != nil || result.ShippedAt == nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}
		result.CompletedAt = &now
	default:
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	err = o.Repository.UpdateByID(r.Context(), result)
	if err != nil {
		fmt.Println("failed to insert:", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		fmt.Println("failed to marshal:", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}

func (o *Order) DeleteByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	orderID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = o.Repository.DeleteByID(r.Context(), orderID)
	if errors.Is(err, order.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)

		return
	} else if err != nil {
		fmt.Println("failed to find by id:", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
}
