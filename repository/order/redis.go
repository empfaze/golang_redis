package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/empfaze/golang_microservices/models"
	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	Client *redis.Client
}

type GetAllPaginationParams struct {
	Size   uint64
	Offset uint64
}

type GetAllResponse struct {
	Orders []models.Order
	Cursor uint64
}

var ErrNotExist = errors.New("Order dows not exist")

func mapOrderIDToKey(id uint64) string {
	return fmt.Sprintf("order:%d", id)
}

func (r *RedisRepository) Create(ctx context.Context, order models.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("Failed to encode an order: %w", err)
	}

	key := mapOrderIDToKey(order.OrderID)

	txn := r.Client.TxPipeline()

	result := txn.SetNX(ctx, key, string(data), 0)
	if err := result.Err(); err != nil {
		txn.Discard()

		return fmt.Errorf("Failed to set: %w", err)
	}

	if err := txn.SAdd(ctx, "orders", key).Err(); err != nil {
		txn.Discard()

		return fmt.Errorf("Failed to add to orders set: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("Failed to exec: %w", err)
	}

	return nil
}

func (r *RedisRepository) FindByID(ctx context.Context, id uint64) (models.Order, error) {
	key := mapOrderIDToKey(id)

	result, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return models.Order{}, ErrNotExist
	} else if err != nil {
		return models.Order{}, fmt.Errorf("Get order: %w", err)
	}

	var order models.Order
	err = json.Unmarshal([]byte(result), &order)
	if err != nil {
		return models.Order{}, fmt.Errorf("Failed to unmarshal the result: %w", err)
	}

	return order, nil
}

func (r *RedisRepository) DeleteByID(ctx context.Context, id uint64) error {
	key := mapOrderIDToKey(id)

	txn := r.Client.TxPipeline()

	err := txn.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		txn.Discard()

		return ErrNotExist
	} else if err != nil {
		txn.Discard()

		return fmt.Errorf("Get order: %w", err)
	}

	if err := txn.SRem(ctx, "orders", key).Err(); err != nil {
		txn.Discard()

		return fmt.Errorf("Failed to remove from orders set: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("Failed to exec: %w", err)
	}

	return nil
}

func (r *RedisRepository) UpdateByID(ctx context.Context, order models.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("Failed to encode an order: %w", err)
	}

	key := mapOrderIDToKey(order.OrderID)

	err = r.Client.SetXX(ctx, key, string(data), 0).Err()
	if errors.Is(err, redis.Nil) {
		return ErrNotExist
	} else if err != nil {
		return fmt.Errorf("Get order: %w", err)
	}

	return nil
}

func (r *RedisRepository) GetAll(ctx context.Context, page GetAllPaginationParams) (GetAllResponse, error) {
	result := r.Client.SScan(ctx, "orders", page.Offset, "*", int64(page.Size))

	keys, cursor, err := result.Result()
	if err != nil {
		return GetAllResponse{}, fmt.Errorf("Failed to get order ids: %w", err)
	}

	if len(keys) == 0 {
		return GetAllResponse{
			Orders: []models.Order{},
		}, nil
	}

	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return GetAllResponse{}, fmt.Errorf("Failed to get orders: %w", err)
	}

	orders := make([]models.Order, len(xs))

	for index, value := range xs {
		value := value.(string)
		var order models.Order

		err := json.Unmarshal([]byte(value), &order)
		if err != nil {
			return GetAllResponse{}, fmt.Errorf("Failed to decode order json: %w", err)
		}

		orders[index] = order
	}

	return GetAllResponse{
		Orders: orders,
		Cursor: cursor,
	}, nil
}
