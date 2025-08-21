package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Side string

const (
	SideBuy  Side = "buy"
	SideSell Side = "sell"
)

type OrderStatus string

const (
	OrderStatusPending         OrderStatus = "pending"
	OrderStatusFilled          OrderStatus = "filled"
	OrderStatusPartiallyFilled OrderStatus = "partially_filled"
	OrderStatusCancelled       OrderStatus = "cancelled"
)

// Order represents an order structure
type Order struct {
	ID        string      `json:"id"`
	Side      Side        `json:"side"`
	Quantity  int         `json:"quantity"`
	Price     float64     `json:"price"`
	Status    OrderStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
}

type Trade struct {
	ID        string    `json:"id"`
	MakerID   string    `json:"maker_id"`
	TakerID   string    `json:"taker_id"`
	Price     float64   `json:"price"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
}

// OrderBook represents the order book with separate buy and sell sides
type OrderBook struct {
	BuyOrders  []Order `json:"buy_orders"`
	SellOrders []Order `json:"sell_orders"`
}

// PlaceOrderRequest represents the request body for placing an order
type PlaceOrderRequest struct {
	Side     Side    `json:"side"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

// PlaceOrderResponse represents the response for placing an order
type PlaceOrderResponse struct {
	OrderID string  `json:"order_id"`
	Trades  []Trade `json:"trades,omitempty"`
}

var orderBook OrderBook
var trades []Trade

func main() {
	// Initialize order book and trades
	orderBook = OrderBook{
		BuyOrders:  make([]Order, 0),
		SellOrders: make([]Order, 0),
	}
	trades = make([]Trade, 0)

	// Define routes
	http.HandleFunc("/api/place-order", placeOrderHandler)
	http.HandleFunc("/api/orders", getOrdersHandler)
	http.HandleFunc("/api/trades", getTradesHandler)
	http.HandleFunc("/api/orderbook", getOrderBookHandler)

	// Start server
	fmt.Println("Server starting on port 8080...")
	fmt.Println("API endpoints:")
	fmt.Println("  POST http://localhost:8080/api/place-order - Place buy/sell order")
	fmt.Println("  GET  http://localhost:8080/api/orders - View all orders")
	fmt.Println("  GET  http://localhost:8080/api/trades - View all trades")
	fmt.Println("  GET  http://localhost:8080/api/orderbook - View order book")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func placeOrderHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow POST method
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Method not allowed",
			"details": "Only POST method is supported for this endpoint",
		})
		return
	}

	// Parse request body
	var req PlaceOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorMessage := "Invalid JSON format in request body"
		if strings.Contains(err.Error(), "unexpected end of JSON input") {
			errorMessage = "Request body is empty or incomplete"
		} else if strings.Contains(err.Error(), "invalid character") {
			errorMessage = "Request body contains invalid JSON syntax"
		} else if strings.Contains(err.Error(), "cannot unmarshal") {
			errorMessage = "Request body contains invalid data types"
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   errorMessage,
			"details": err.Error(),
		})
		return
	}

	// Validate request with detailed error messages
	var validationErrors []string

	// Validate quantity
	if req.Quantity <= 0 {
		validationErrors = append(validationErrors, "quantity must be a positive number (received: "+fmt.Sprintf("%d", req.Quantity)+")")
	} else if req.Quantity > 999999999 {
		validationErrors = append(validationErrors, "quantity is too high (maximum allowed: 999,999,999)")
	}

	// Validate price
	if req.Price <= 0 {
		validationErrors = append(validationErrors, "price must be a positive number (received: "+fmt.Sprintf("%.2f", req.Price)+")")
	} else if req.Price > 999999999.99 {
		validationErrors = append(validationErrors, "price is too high (maximum allowed: 999,999,999.99)")
	}

	// Validate side
	if req.Side == "" {
		validationErrors = append(validationErrors, "side is required and cannot be empty")
	} else if req.Side != SideBuy && req.Side != SideSell {
		validationErrors = append(validationErrors, "side must be either 'buy' or 'sell' (received: '"+string(req.Side)+"')")
	}

	// Return all validation errors if any exist
	if len(validationErrors) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Validation failed",
			"details": validationErrors,
		})
		return
	}

	// Create new order
	order := Order{
		ID:        generateOrderID(),
		Side:      req.Side,
		Quantity:  req.Quantity,
		Price:     req.Price,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}

	// Process the order through the order book
	processOrder(order)

	// Return all trades in match order
	response := PlaceOrderResponse{
		OrderID: order.ID,
		Trades:  trades,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// generateOrderID creates a simple order ID
func generateOrderID() string {
	return uuid.New().String()
}

// generateTradeID creates a simple trade ID
func generateTradeID() string {
	return uuid.New().String()
}

// processOrder processes an incoming order through the order book
func processOrder(order Order) {
	if order.Side == SideBuy {
		// Try to match buy order against sell orders
		remainingOrder, _ := matchBuyOrder(order)

		// If there's remaining quantity, add to buy side of order book
		if remainingOrder.Quantity > 0 {
			addToOrderBook(remainingOrder)
		}
	} else {
		// Try to match sell order against buy orders
		remainingOrder, _ := matchSellOrder(order)

		// If there's remaining quantity, add to sell side of order book
		if remainingOrder.Quantity > 0 {
			addToOrderBook(remainingOrder)
		}
	}
}

// matchBuyOrder matches a buy order against existing sell orders
func matchBuyOrder(buyOrder Order) (Order, []Trade) {
	var executedTrades []Trade
	remainingOrder := buyOrder

	// Sort sell orders by price (lowest first) and then by time (oldest first)
	sort.Slice(orderBook.SellOrders, func(i, j int) bool {
		if orderBook.SellOrders[i].Price != orderBook.SellOrders[j].Price {
			return orderBook.SellOrders[i].Price < orderBook.SellOrders[j].Price
		}
		return orderBook.SellOrders[i].CreatedAt.Before(orderBook.SellOrders[j].CreatedAt)
	})

	// Try to match against sell orders
	for i := 0; i < len(orderBook.SellOrders) && remainingOrder.Quantity > 0; {
		sellOrder := orderBook.SellOrders[i]

		// Check if prices can match (buy price >= sell price)
		if remainingOrder.Price >= sellOrder.Price {
			// Execute trade
			tradeQuantity := min(remainingOrder.Quantity, sellOrder.Quantity)
			trade := Trade{
				ID:        generateTradeID(),
				MakerID:   sellOrder.ID,      // Resting order (sell)
				TakerID:   remainingOrder.ID, // Incoming order (buy)
				Price:     sellOrder.Price,   // Trade at resting order's price
				Quantity:  tradeQuantity,
				CreatedAt: time.Now(),
			}

			executedTrades = append(executedTrades, trade)
			trades = append(trades, trade)

			// Update quantities
			remainingOrder.Quantity -= tradeQuantity
			orderBook.SellOrders[i].Quantity -= tradeQuantity

			// Update order status
			if orderBook.SellOrders[i].Quantity == 0 {
				orderBook.SellOrders[i].Status = OrderStatusFilled
				// Remove filled order
				orderBook.SellOrders = append(orderBook.SellOrders[:i], orderBook.SellOrders[i+1:]...)
				// Don't increment i since we removed an element
			} else {
				orderBook.SellOrders[i].Status = OrderStatusPartiallyFilled
				i++ // Move to next order
			}

			// Update remaining order status
			if remainingOrder.Quantity == 0 {
				remainingOrder.Status = OrderStatusFilled
			} else {
				remainingOrder.Status = OrderStatusPartiallyFilled
			}
		} else {
			// No more matches possible
			break
		}
	}

	return remainingOrder, executedTrades
}

// matchSellOrder matches a sell order against existing buy orders
func matchSellOrder(sellOrder Order) (Order, []Trade) {
	var executedTrades []Trade
	remainingOrder := sellOrder

	// Sort buy orders by price (highest first) and then by time (oldest first)
	sort.Slice(orderBook.BuyOrders, func(i, j int) bool {
		if orderBook.BuyOrders[i].Price != orderBook.BuyOrders[j].Price {
			return orderBook.BuyOrders[i].Price > orderBook.BuyOrders[j].Price
		}
		return orderBook.BuyOrders[i].CreatedAt.Before(orderBook.BuyOrders[j].CreatedAt)
	})

	// Try to match against buy orders
	for i := 0; i < len(orderBook.BuyOrders) && remainingOrder.Quantity > 0; {
		buyOrder := orderBook.BuyOrders[i]

		// Check if prices can match (sell price <= buy price)
		if remainingOrder.Price <= buyOrder.Price {
			// Execute trade
			tradeQuantity := min(remainingOrder.Quantity, buyOrder.Quantity)
			trade := Trade{
				ID:        generateTradeID(),
				MakerID:   buyOrder.ID,       // Resting order (buy)
				TakerID:   remainingOrder.ID, // Incoming order (sell)
				Price:     buyOrder.Price,    // Trade at resting order's price
				Quantity:  tradeQuantity,
				CreatedAt: time.Now(),
			}

			executedTrades = append(executedTrades, trade)
			trades = append(trades, trade)

			// Update quantities
			remainingOrder.Quantity -= tradeQuantity
			orderBook.BuyOrders[i].Quantity -= tradeQuantity

			// Update order status
			if orderBook.BuyOrders[i].Quantity == 0 {
				orderBook.BuyOrders[i].Status = OrderStatusFilled
				// Remove filled order
				orderBook.BuyOrders = append(orderBook.BuyOrders[:i], orderBook.BuyOrders[i+1:]...)
				// Don't increment i since we removed an element
			} else {
				orderBook.BuyOrders[i].Status = OrderStatusPartiallyFilled
				i++ // Move to next order
			}

			// Update remaining order status
			if remainingOrder.Quantity == 0 {
				remainingOrder.Status = OrderStatusFilled
			} else {
				remainingOrder.Status = OrderStatusPartiallyFilled
			}
		} else {
			// No more matches possible
			break
		}
	}

	return remainingOrder, executedTrades
}

// addToOrderBook adds an order to the appropriate side of the order book
func addToOrderBook(order Order) {
	if order.Side == SideBuy {
		orderBook.BuyOrders = append(orderBook.BuyOrders, order)
		// Sort buy orders by price (highest first) and then by time (oldest first)
		sort.Slice(orderBook.BuyOrders, func(i, j int) bool {
			if orderBook.BuyOrders[i].Price != orderBook.BuyOrders[j].Price {
				return orderBook.BuyOrders[i].Price > orderBook.BuyOrders[j].Price
			}
			return orderBook.BuyOrders[i].CreatedAt.Before(orderBook.BuyOrders[j].CreatedAt)
		})
	} else {
		orderBook.SellOrders = append(orderBook.SellOrders, order)
		// Sort sell orders by price (lowest first) and then by time (oldest first)
		sort.Slice(orderBook.SellOrders, func(i, j int) bool {
			if orderBook.SellOrders[i].Price != orderBook.SellOrders[j].Price {
				return orderBook.SellOrders[i].Price < orderBook.SellOrders[j].Price
			}
			return orderBook.SellOrders[i].CreatedAt.Before(orderBook.SellOrders[j].CreatedAt)
		})
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getAllOrders returns all orders in the order book
func getAllOrders() []Order {
	var allOrders []Order
	allOrders = append(allOrders, orderBook.BuyOrders...)
	allOrders = append(allOrders, orderBook.SellOrders...)
	return allOrders
}

// getOrdersHandler returns all orders in the system
func getOrdersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	allOrders := getAllOrders()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"orders": allOrders,
		"count":  len(allOrders),
	})
}

// getTradesHandler returns all trades in the system
func getTradesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"trades": trades,
		"count":  len(trades),
	})
}

// getOrderBookHandler returns the current order book
func getOrderBookHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"orderbook":  orderBook,
		"buy_count":  len(orderBook.BuyOrders),
		"sell_count": len(orderBook.SellOrders),
	})
}
