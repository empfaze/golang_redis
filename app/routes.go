package app

import (
	"github.com/empfaze/golang_redis/handlers"
	"github.com/empfaze/golang_redis/repositories/order"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (a *App) getRoutes() {
	router := chi.NewRouter()

	router.Use(middleware.Logger)

	router.Route("/orders", a.getOrderRoutes)

	a.router = router
}

func (a *App) getOrderRoutes(router chi.Router) {
	handler := &handlers.Order{
		Repository: &order.RedisRepository{
			Client: a.rDB,
		},
	}

	router.Get("/", handler.GetAll)
	router.Get("/{id}", handler.GetById)
	router.Post("/", handler.Create)
	router.Put("/{id}", handler.UpdateByID)
	router.Delete("/{id}", handler.DeleteByID)
}
