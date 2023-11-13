package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type App struct {
	router http.Handler
	rDB    *redis.Client
}

func New() *App {
	app := &App{
		rDB: redis.NewClient(&redis.Options{}),
	}

	app.getRoutes()

	return app
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":3000",
		Handler: a.router,
	}

	err := a.rDB.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("Failed to connect to redis: %w", err)
	}

	defer func() {
		if err := a.rDB.Close(); err != nil {
			fmt.Println("Failed to close redis: %w", err)
		}
	}()

	errorChannel := make(chan error, 1)

	go func() {
		err = server.ListenAndServe()
		if err != nil {
			errorChannel <- fmt.Errorf("Failed to start the server: %w", err)
		}

		close(errorChannel)
	}()

	select {
	case err = <-errorChannel:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		return server.Shutdown(timeout)
	}
}
