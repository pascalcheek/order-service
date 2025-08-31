package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"order-service/internal/service"

	"github.com/gorilla/mux"
)

type APIHandler struct {
	orderService *service.OrderService
}

func NewAPIHandler(orderService *service.OrderService) *APIHandler {
	return &APIHandler{
		orderService: orderService,
	}
}

func (h *APIHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderUID := vars["order_uid"]

	if orderUID == "" {
		http.Error(w, "Order UID is required", http.StatusBadRequest)
		return
	}

	order, err := h.orderService.GetOrder(orderUID)
	if err != nil {
		log.Printf("Error getting order %s: %v", orderUID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if order == nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		log.Printf("Error encoding order to JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *APIHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
