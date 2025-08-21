package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Test helper functions
func setupTest() {
	// Reset global state
	orderBook = OrderBook{
		BuyOrders:  make([]Order, 0),
		SellOrders: make([]Order, 0),
	}
	trades = make([]Trade, 0)
}

func TestPlaceOrderHandler_ValidBuyOrder(t *testing.T) {
	setupTest()

	req := PlaceOrderRequest{
		Side:     SideBuy,
		Price:    100.0,
		Quantity: 10,
	}

	jsonData, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/api/place-order", bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	placeOrderHandler(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", response.Code)
	}

	var result PlaceOrderResponse
	json.Unmarshal(response.Body.Bytes(), &result)

	if result.OrderID == "" {
		t.Error("Expected order ID to be generated")
	}

	// Since we now return all trades, for the first order there should be no trades
	if len(result.Trades) != 0 {
		t.Errorf("Expected no trades for first order, got %d", len(result.Trades))
	}
}

func TestPlaceOrderHandler_ValidSellOrder(t *testing.T) {
	setupTest()

	req := PlaceOrderRequest{
		Side:     SideSell,
		Price:    100.0,
		Quantity: 10,
	}

	jsonData, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/api/place-order", bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	placeOrderHandler(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", response.Code)
	}

	var result PlaceOrderResponse
	json.Unmarshal(response.Body.Bytes(), &result)

	if result.OrderID == "" {
		t.Error("Expected order ID to be generated")
	}
}

func TestPlaceOrderHandler_InvalidJSON(t *testing.T) {
	setupTest()

	request := httptest.NewRequest("POST", "/api/place-order", bytes.NewBufferString("invalid json"))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	placeOrderHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", response.Code)
	}
}

func TestPlaceOrderHandler_EmptyBody(t *testing.T) {
	setupTest()

	request := httptest.NewRequest("POST", "/api/place-order", bytes.NewBufferString(""))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	placeOrderHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", response.Code)
	}
}

func TestPlaceOrderHandler_InvalidSide(t *testing.T) {
	setupTest()

	req := PlaceOrderRequest{
		Side:     "invalid",
		Price:    100.0,
		Quantity: 10,
	}

	jsonData, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/api/place-order", bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	placeOrderHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", response.Code)
	}
}

func TestPlaceOrderHandler_ZeroPrice(t *testing.T) {
	setupTest()

	req := PlaceOrderRequest{
		Side:     SideBuy,
		Price:    0.0,
		Quantity: 10,
	}

	jsonData, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/api/place-order", bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	placeOrderHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", response.Code)
	}
}

func TestPlaceOrderHandler_NegativePrice(t *testing.T) {
	setupTest()

	req := PlaceOrderRequest{
		Side:     SideBuy,
		Price:    -10.0,
		Quantity: 10,
	}

	jsonData, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/api/place-order", bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	placeOrderHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", response.Code)
	}
}

func TestPlaceOrderHandler_ZeroQuantity(t *testing.T) {
	setupTest()

	req := PlaceOrderRequest{
		Side:     SideBuy,
		Price:    100.0,
		Quantity: 0,
	}

	jsonData, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/api/place-order", bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	placeOrderHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", response.Code)
	}
}

func TestPlaceOrderHandler_NegativeQuantity(t *testing.T) {
	setupTest()

	req := PlaceOrderRequest{
		Side:     SideBuy,
		Price:    100.0,
		Quantity: -10,
	}

	jsonData, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/api/place-order", bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	placeOrderHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", response.Code)
	}
}

func TestPlaceOrderHandler_ExcessivePrice(t *testing.T) {
	setupTest()

	req := PlaceOrderRequest{
		Side:     SideBuy,
		Price:    1000000000.0, // Over the limit
		Quantity: 10,
	}

	jsonData, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/api/place-order", bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	placeOrderHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", response.Code)
	}
}

func TestPlaceOrderHandler_ExcessiveQuantity(t *testing.T) {
	setupTest()

	req := PlaceOrderRequest{
		Side:     SideBuy,
		Price:    100.0,
		Quantity: 1000000000, // Over the limit
	}

	jsonData, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/api/place-order", bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	placeOrderHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", response.Code)
	}
}

func TestPlaceOrderHandler_WrongMethod(t *testing.T) {
	setupTest()

	request := httptest.NewRequest("GET", "/api/place-order", nil)
	response := httptest.NewRecorder()

	placeOrderHandler(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", response.Code)
	}
}

func TestProcessOrder_BuyOrderNoMatch(t *testing.T) {
	setupTest()

	order := Order{
		ID:        "test-buy",
		Side:      SideBuy,
		Price:     100.0,
		Quantity:  10,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}

	processOrder(order)

	if len(orderBook.BuyOrders) != 1 {
		t.Errorf("Expected 1 buy order in book, got %d", len(orderBook.BuyOrders))
	}

	if orderBook.BuyOrders[0].ID != "test-buy" {
		t.Error("Expected buy order to be in book")
	}
}

func TestProcessOrder_SellOrderNoMatch(t *testing.T) {
	setupTest()

	order := Order{
		ID:        "test-sell",
		Side:      SideSell,
		Price:     100.0,
		Quantity:  10,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}

	processOrder(order)

	if len(orderBook.SellOrders) != 1 {
		t.Errorf("Expected 1 sell order in book, got %d", len(orderBook.SellOrders))
	}

	if orderBook.SellOrders[0].ID != "test-sell" {
		t.Error("Expected sell order to be in book")
	}
}

func TestProcessOrder_BuyOrderMatchesSell(t *testing.T) {
	setupTest()

	// Add a sell order to the book
	sellOrder := Order{
		ID:        "sell-1",
		Side:      SideSell,
		Price:     100.0,
		Quantity:  10,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}
	orderBook.SellOrders = append(orderBook.SellOrders, sellOrder)

	// Place a buy order that should match
	buyOrder := Order{
		ID:        "buy-1",
		Side:      SideBuy,
		Price:     101.0,
		Quantity:  5,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}

	processOrder(buyOrder)

	// Should have no buy orders (fully matched)
	if len(orderBook.BuyOrders) != 0 {
		t.Errorf("Expected 0 buy orders, got %d", len(orderBook.BuyOrders))
	}

	// Should have 1 sell order with reduced quantity
	if len(orderBook.SellOrders) != 1 {
		t.Errorf("Expected 1 sell order, got %d", len(orderBook.SellOrders))
	}

	if orderBook.SellOrders[0].Quantity != 5 {
		t.Errorf("Expected sell order quantity to be 5, got %d", orderBook.SellOrders[0].Quantity)
	}

	// Should have 1 trade
	if len(trades) != 1 {
		t.Errorf("Expected 1 trade, got %d", len(trades))
	}

	if trades[0].Price != 100.0 {
		t.Errorf("Expected trade price to be 100.0, got %.2f", trades[0].Price)
	}
}

func TestProcessOrder_SellOrderMatchesBuy(t *testing.T) {
	setupTest()

	// Add a buy order to the book
	buyOrder := Order{
		ID:        "buy-1",
		Side:      SideBuy,
		Price:     101.0,
		Quantity:  10,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}
	orderBook.BuyOrders = append(orderBook.BuyOrders, buyOrder)

	// Place a sell order that should match
	sellOrder := Order{
		ID:        "sell-1",
		Side:      SideSell,
		Price:     100.0,
		Quantity:  5,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}

	processOrder(sellOrder)

	// Should have no sell orders (fully matched)
	if len(orderBook.SellOrders) != 0 {
		t.Errorf("Expected 0 sell orders, got %d", len(orderBook.SellOrders))
	}

	// Should have 1 buy order with reduced quantity
	if len(orderBook.BuyOrders) != 1 {
		t.Errorf("Expected 1 buy order, got %d", len(orderBook.BuyOrders))
	}

	if orderBook.BuyOrders[0].Quantity != 5 {
		t.Errorf("Expected buy order quantity to be 5, got %d", orderBook.BuyOrders[0].Quantity)
	}

	// Should have 1 trade
	if len(trades) != 1 {
		t.Errorf("Expected 1 trade, got %d", len(trades))
	}

	if trades[0].Price != 101.0 {
		t.Errorf("Expected trade price to be 101.0, got %.2f", trades[0].Price)
	}
}

func TestProcessOrder_PartialMatch(t *testing.T) {
	setupTest()

	// Add a sell order to the book
	sellOrder := Order{
		ID:        "sell-1",
		Side:      SideSell,
		Price:     100.0,
		Quantity:  5,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}
	orderBook.SellOrders = append(orderBook.SellOrders, sellOrder)

	// Place a buy order with larger quantity
	buyOrder := Order{
		ID:        "buy-1",
		Side:      SideBuy,
		Price:     101.0,
		Quantity:  10,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}

	processOrder(buyOrder)

	// Should have 1 buy order with remaining quantity
	if len(orderBook.BuyOrders) != 1 {
		t.Errorf("Expected 1 buy order, got %d", len(orderBook.BuyOrders))
	}

	if orderBook.BuyOrders[0].Quantity != 5 {
		t.Errorf("Expected buy order quantity to be 5, got %d", orderBook.BuyOrders[0].Quantity)
	}

	// Should have no sell orders (fully matched)
	if len(orderBook.SellOrders) != 0 {
		t.Errorf("Expected 0 sell orders, got %d", len(orderBook.SellOrders))
	}

	// Should have 1 trade
	if len(trades) != 1 {
		t.Errorf("Expected 1 trade, got %d", len(trades))
	}
}

func TestProcessOrder_MultipleMatches(t *testing.T) {
	setupTest()

	// Add multiple sell orders at different prices
	sellOrder1 := Order{
		ID:        "sell-1",
		Side:      SideSell,
		Price:     99.0,
		Quantity:  5,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}
	sellOrder2 := Order{
		ID:        "sell-2",
		Side:      SideSell,
		Price:     100.0,
		Quantity:  5,
		Status:    OrderStatusPending,
		CreatedAt: time.Now().Add(time.Millisecond),
	}
	orderBook.SellOrders = append(orderBook.SellOrders, sellOrder1, sellOrder2)

	// Place a buy order that should match both
	buyOrder := Order{
		ID:        "buy-1",
		Side:      SideBuy,
		Price:     101.0,
		Quantity:  8,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}

	processOrder(buyOrder)

	// Should have 0 buy orders (fully consumed)
	if len(orderBook.BuyOrders) != 0 {
		t.Errorf("Expected 0 buy orders, got %d", len(orderBook.BuyOrders))
	}

	// Should have 1 sell order with remaining quantity
	if len(orderBook.SellOrders) != 1 {
		t.Errorf("Expected 1 sell order, got %d", len(orderBook.SellOrders))
	}

	if orderBook.SellOrders[0].Quantity != 2 {
		t.Errorf("Expected sell order quantity to be 2, got %d", orderBook.SellOrders[0].Quantity)
	}

	// Should have 2 trades
	if len(trades) != 2 {
		t.Errorf("Expected 2 trades, got %d", len(trades))
	}

	// First trade should be at $99.00 (better price)
	if trades[0].Price != 99.0 {
		t.Errorf("Expected first trade price to be 99.0, got %.2f", trades[0].Price)
	}

	// Second trade should be at $100.00
	if trades[1].Price != 100.0 {
		t.Errorf("Expected second trade price to be 100.0, got %.2f", trades[1].Price)
	}
}

func TestProcessOrder_PriceTimePriority(t *testing.T) {
	setupTest()

	// Add sell orders with same price but different times
	sellOrder1 := Order{
		ID:        "sell-1",
		Side:      SideSell,
		Price:     100.0,
		Quantity:  5,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}
	sellOrder2 := Order{
		ID:        "sell-2",
		Side:      SideSell,
		Price:     100.0,
		Quantity:  5,
		Status:    OrderStatusPending,
		CreatedAt: time.Now().Add(time.Millisecond), // Later time
	}
	orderBook.SellOrders = append(orderBook.SellOrders, sellOrder1, sellOrder2)

	// Place a buy order that should match both
	buyOrder := Order{
		ID:        "buy-1",
		Side:      SideBuy,
		Price:     101.0,
		Quantity:  8,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}

	processOrder(buyOrder)

	// Should have 2 trades
	if len(trades) != 2 {
		t.Errorf("Expected 2 trades, got %d", len(trades))
	}

	// First trade should be with sell-1 (earlier time)
	if trades[0].MakerID != "sell-1" {
		t.Errorf("Expected first trade to be with sell-1, got %s", trades[0].MakerID)
	}

	// Second trade should be with sell-2 (later time)
	if trades[1].MakerID != "sell-2" {
		t.Errorf("Expected second trade to be with sell-2, got %s", trades[1].MakerID)
	}
}

func TestProcessOrder_NoMatchDueToPrice(t *testing.T) {
	setupTest()

	// Add a sell order at $100
	sellOrder := Order{
		ID:        "sell-1",
		Side:      SideSell,
		Price:     100.0,
		Quantity:  10,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}
	orderBook.SellOrders = append(orderBook.SellOrders, sellOrder)

	// Place a buy order at $99 (should not match)
	buyOrder := Order{
		ID:        "buy-1",
		Side:      SideBuy,
		Price:     99.0,
		Quantity:  5,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}

	processOrder(buyOrder)

	// Should have 1 buy order in book
	if len(orderBook.BuyOrders) != 1 {
		t.Errorf("Expected 1 buy order, got %d", len(orderBook.BuyOrders))
	}

	// Should have 1 sell order in book
	if len(orderBook.SellOrders) != 1 {
		t.Errorf("Expected 1 sell order, got %d", len(orderBook.SellOrders))
	}

	// Should have no trades
	if len(trades) != 0 {
		t.Errorf("Expected 0 trades, got %d", len(trades))
	}
}

func TestGetOrdersHandler(t *testing.T) {
	setupTest()

	// Add some orders to the book
	order1 := Order{
		ID:        "order-1",
		Side:      SideBuy,
		Price:     100.0,
		Quantity:  10,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}
	order2 := Order{
		ID:        "order-2",
		Side:      SideSell,
		Price:     101.0,
		Quantity:  5,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}
	orderBook.BuyOrders = append(orderBook.BuyOrders, order1)
	orderBook.SellOrders = append(orderBook.SellOrders, order2)

	request := httptest.NewRequest("GET", "/api/orders", nil)
	response := httptest.NewRecorder()

	getOrdersHandler(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", response.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &result)

	if result["count"].(float64) != 2 {
		t.Errorf("Expected 2 orders, got %.0f", result["count"].(float64))
	}
}

func TestGetTradesHandler(t *testing.T) {
	setupTest()

	// Add some trades
	trade1 := Trade{
		ID:        "trade-1",
		MakerID:   "maker-1",
		TakerID:   "taker-1",
		Price:     100.0,
		Quantity:  5,
		CreatedAt: time.Now(),
	}
	trade2 := Trade{
		ID:        "trade-2",
		MakerID:   "maker-2",
		TakerID:   "taker-2",
		Price:     101.0,
		Quantity:  3,
		CreatedAt: time.Now(),
	}
	trades = append(trades, trade1, trade2)

	request := httptest.NewRequest("GET", "/api/trades", nil)
	response := httptest.NewRecorder()

	getTradesHandler(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", response.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &result)

	if result["count"].(float64) != 2 {
		t.Errorf("Expected 2 trades, got %.0f", result["count"].(float64))
	}
}

func TestGetOrderBookHandler(t *testing.T) {
	setupTest()

	// Add some orders to the book
	order1 := Order{
		ID:        "order-1",
		Side:      SideBuy,
		Price:     100.0,
		Quantity:  10,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}
	order2 := Order{
		ID:        "order-2",
		Side:      SideSell,
		Price:     101.0,
		Quantity:  5,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}
	orderBook.BuyOrders = append(orderBook.BuyOrders, order1)
	orderBook.SellOrders = append(orderBook.SellOrders, order2)

	request := httptest.NewRequest("GET", "/api/orderbook", nil)
	response := httptest.NewRecorder()

	getOrderBookHandler(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", response.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &result)

	if result["buy_count"].(float64) != 1 {
		t.Errorf("Expected 1 buy order, got %.0f", result["buy_count"].(float64))
	}

	if result["sell_count"].(float64) != 1 {
		t.Errorf("Expected 1 sell order, got %.0f", result["sell_count"].(float64))
	}
}

func TestMinFunction(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{5, 10, 5},
		{10, 5, 5},
		{5, 5, 5},
		{0, 5, 0},
		{-5, 5, -5},
		{5, -5, -5},
	}

	for _, test := range tests {
		result := min(test.a, test.b)
		if result != test.expected {
			t.Errorf("min(%d, %d) = %d, expected %d", test.a, test.b, result, test.expected)
		}
	}
}

func TestGenerateOrderID(t *testing.T) {
	id1 := generateOrderID()
	id2 := generateOrderID()

	if id1 == "" {
		t.Error("Expected non-empty order ID")
	}

	if id2 == "" {
		t.Error("Expected non-empty order ID")
	}

	if id1 == id2 {
		t.Error("Expected different order IDs")
	}
}

func TestGenerateTradeID(t *testing.T) {
	id1 := generateTradeID()
	id2 := generateTradeID()

	if id1 == "" {
		t.Error("Expected non-empty trade ID")
	}

	if id2 == "" {
		t.Error("Expected non-empty trade ID")
	}

	if id1 == id2 {
		t.Error("Expected different trade IDs")
	}
}

func TestAddToOrderBook_BuyOrders(t *testing.T) {
	setupTest()

	// Add buy orders in random order
	order1 := Order{
		ID:        "buy-1",
		Side:      SideBuy,
		Price:     100.0,
		Quantity:  10,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}
	order2 := Order{
		ID:        "buy-2",
		Side:      SideBuy,
		Price:     101.0, // Higher price
		Quantity:  5,
		Status:    OrderStatusPending,
		CreatedAt: time.Now().Add(time.Millisecond),
	}

	addToOrderBook(order1)
	addToOrderBook(order2)

	// Should be sorted by price (highest first)
	if len(orderBook.BuyOrders) != 2 {
		t.Errorf("Expected 2 buy orders, got %d", len(orderBook.BuyOrders))
	}

	if orderBook.BuyOrders[0].Price != 101.0 {
		t.Errorf("Expected first buy order price to be 101.0, got %.2f", orderBook.BuyOrders[0].Price)
	}

	if orderBook.BuyOrders[1].Price != 100.0 {
		t.Errorf("Expected second buy order price to be 100.0, got %.2f", orderBook.BuyOrders[1].Price)
	}
}

func TestAddToOrderBook_SellOrders(t *testing.T) {
	setupTest()

	// Add sell orders in random order
	order1 := Order{
		ID:        "sell-1",
		Side:      SideSell,
		Price:     101.0,
		Quantity:  10,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}
	order2 := Order{
		ID:        "sell-2",
		Side:      SideSell,
		Price:     100.0, // Lower price
		Quantity:  5,
		Status:    OrderStatusPending,
		CreatedAt: time.Now().Add(time.Millisecond),
	}

	addToOrderBook(order1)
	addToOrderBook(order2)

	// Should be sorted by price (lowest first)
	if len(orderBook.SellOrders) != 2 {
		t.Errorf("Expected 2 sell orders, got %d", len(orderBook.SellOrders))
	}

	if orderBook.SellOrders[0].Price != 100.0 {
		t.Errorf("Expected first sell order price to be 100.0, got %.2f", orderBook.SellOrders[0].Price)
	}

	if orderBook.SellOrders[1].Price != 101.0 {
		t.Errorf("Expected second sell order price to be 101.0, got %.2f", orderBook.SellOrders[1].Price)
	}
}

func TestGetAllOrders(t *testing.T) {
	setupTest()

	// Add orders to both sides
	order1 := Order{
		ID:        "buy-1",
		Side:      SideBuy,
		Price:     100.0,
		Quantity:  10,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}
	order2 := Order{
		ID:        "sell-1",
		Side:      SideSell,
		Price:     101.0,
		Quantity:  5,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}
	orderBook.BuyOrders = append(orderBook.BuyOrders, order1)
	orderBook.SellOrders = append(orderBook.SellOrders, order2)

	allOrders := getAllOrders()

	if len(allOrders) != 2 {
		t.Errorf("Expected 2 orders, got %d", len(allOrders))
	}
}

// Additional edge case tests

func TestProcessOrder_ExactPriceMatch(t *testing.T) {
	setupTest()

	// Add a sell order at exactly $100
	sellOrder := Order{
		ID:        "sell-1",
		Side:      SideSell,
		Price:     100.0,
		Quantity:  10,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}
	orderBook.SellOrders = append(orderBook.SellOrders, sellOrder)

	// Place a buy order at exactly $100 (should match)
	buyOrder := Order{
		ID:        "buy-1",
		Side:      SideBuy,
		Price:     100.0,
		Quantity:  5,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}

	processOrder(buyOrder)

	// Should have 1 trade
	if len(trades) != 1 {
		t.Errorf("Expected 1 trade, got %d", len(trades))
	}

	if trades[0].Price != 100.0 {
		t.Errorf("Expected trade price to be 100.0, got %.2f", trades[0].Price)
	}
}

func TestProcessOrder_ExactQuantityMatch(t *testing.T) {
	setupTest()

	// Add a sell order with quantity 10
	sellOrder := Order{
		ID:        "sell-1",
		Side:      SideSell,
		Price:     100.0,
		Quantity:  10,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}
	orderBook.SellOrders = append(orderBook.SellOrders, sellOrder)

	// Place a buy order with exactly the same quantity
	buyOrder := Order{
		ID:        "buy-1",
		Side:      SideBuy,
		Price:     101.0,
		Quantity:  10,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}

	processOrder(buyOrder)

	// Should have no orders in book (both fully matched)
	if len(orderBook.BuyOrders) != 0 {
		t.Errorf("Expected 0 buy orders, got %d", len(orderBook.BuyOrders))
	}

	if len(orderBook.SellOrders) != 0 {
		t.Errorf("Expected 0 sell orders, got %d", len(orderBook.SellOrders))
	}

	// Should have 1 trade
	if len(trades) != 1 {
		t.Errorf("Expected 1 trade, got %d", len(trades))
	}

	if trades[0].Quantity != 10 {
		t.Errorf("Expected trade quantity to be 10, got %d", trades[0].Quantity)
	}
}

func TestProcessOrder_EmptyOrderBook(t *testing.T) {
	setupTest()

	// Place orders when order book is empty
	buyOrder := Order{
		ID:        "buy-1",
		Side:      SideBuy,
		Price:     100.0,
		Quantity:  10,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}

	processOrder(buyOrder)

	if len(orderBook.BuyOrders) != 1 {
		t.Errorf("Expected 1 buy order, got %d", len(orderBook.BuyOrders))
	}

	if len(trades) != 0 {
		t.Errorf("Expected 0 trades, got %d", len(trades))
	}
}

func TestProcessOrder_OrderStatusUpdates(t *testing.T) {
	setupTest()

	// Add a sell order
	sellOrder := Order{
		ID:        "sell-1",
		Side:      SideSell,
		Price:     100.0,
		Quantity:  10,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}
	orderBook.SellOrders = append(orderBook.SellOrders, sellOrder)

	// Place a buy order that partially matches
	buyOrder := Order{
		ID:        "buy-1",
		Side:      SideBuy,
		Price:     101.0,
		Quantity:  5,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}

	processOrder(buyOrder)

	// Sell order should be partially filled
	if orderBook.SellOrders[0].Status != OrderStatusPartiallyFilled {
		t.Errorf("Expected sell order status to be partially_filled, got %s", orderBook.SellOrders[0].Status)
	}

	// Buy order should be fully filled (no remaining quantity)
	if len(orderBook.BuyOrders) != 0 {
		t.Errorf("Expected 0 buy orders, got %d", len(orderBook.BuyOrders))
	}
}

func TestProcessOrder_MakerTakerIdentification(t *testing.T) {
	setupTest()

	// Add a sell order (will be maker)
	sellOrder := Order{
		ID:        "sell-1",
		Side:      SideSell,
		Price:     100.0,
		Quantity:  10,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}
	orderBook.SellOrders = append(orderBook.SellOrders, sellOrder)

	// Place a buy order (will be taker)
	buyOrder := Order{
		ID:        "buy-1",
		Side:      SideBuy,
		Price:     101.0,
		Quantity:  5,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}

	processOrder(buyOrder)

	// Should have 1 trade
	if len(trades) != 1 {
		t.Errorf("Expected 1 trade, got %d", len(trades))
	}

	// Verify maker and taker
	if trades[0].MakerID != "sell-1" {
		t.Errorf("Expected maker ID to be sell-1, got %s", trades[0].MakerID)
	}

	if trades[0].TakerID != "buy-1" {
		t.Errorf("Expected taker ID to be buy-1, got %s", trades[0].TakerID)
	}
}

func TestProcessOrder_TradePriceIsMakerPrice(t *testing.T) {
	setupTest()

	// Add a sell order at $100 (maker)
	sellOrder := Order{
		ID:        "sell-1",
		Side:      SideSell,
		Price:     100.0,
		Quantity:  10,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}
	orderBook.SellOrders = append(orderBook.SellOrders, sellOrder)

	// Place a buy order at $101 (taker)
	buyOrder := Order{
		ID:        "buy-1",
		Side:      SideBuy,
		Price:     101.0,
		Quantity:  5,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}

	processOrder(buyOrder)

	// Trade should execute at maker's price ($100)
	if len(trades) != 1 {
		t.Errorf("Expected 1 trade, got %d", len(trades))
	}

	if trades[0].Price != 100.0 {
		t.Errorf("Expected trade price to be 100.0 (maker's price), got %.2f", trades[0].Price)
	}
}

func TestProcessOrder_ConcurrentOrders(t *testing.T) {
	setupTest()

	// Add multiple orders quickly to test sorting
	orders := []Order{
		{ID: "buy-1", Side: SideBuy, Price: 100.0, Quantity: 10, Status: OrderStatusPending, CreatedAt: time.Now()},
		{ID: "buy-2", Side: SideBuy, Price: 101.0, Quantity: 5, Status: OrderStatusPending, CreatedAt: time.Now()},
		{ID: "buy-3", Side: SideBuy, Price: 100.0, Quantity: 3, Status: OrderStatusPending, CreatedAt: time.Now()},
		{ID: "sell-1", Side: SideSell, Price: 103.0, Quantity: 8, Status: OrderStatusPending, CreatedAt: time.Now()}, // Higher price, no match
		{ID: "sell-2", Side: SideSell, Price: 102.0, Quantity: 6, Status: OrderStatusPending, CreatedAt: time.Now()}, // Higher price, no match
	}

	for _, order := range orders {
		processOrder(order)
	}

	// Verify buy orders are sorted by price (highest first)
	if len(orderBook.BuyOrders) != 3 {
		t.Errorf("Expected 3 buy orders, got %d", len(orderBook.BuyOrders))
	}

	if orderBook.BuyOrders[0].Price != 101.0 {
		t.Errorf("Expected first buy order price to be 101.0, got %.2f", orderBook.BuyOrders[0].Price)
	}

	// Verify sell orders are sorted by price (lowest first)
	if len(orderBook.SellOrders) != 2 {
		t.Errorf("Expected 2 sell orders, got %d", len(orderBook.SellOrders))
	}

	if orderBook.SellOrders[0].Price != 102.0 {
		t.Errorf("Expected first sell order price to be 102.0, got %.2f", orderBook.SellOrders[0].Price)
	}
}

func TestProcessOrder_EdgeCasePrices(t *testing.T) {
	setupTest()

	// Test with very small price difference
	sellOrder := Order{
		ID:        "sell-1",
		Side:      SideSell,
		Price:     100.0001,
		Quantity:  10,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}
	orderBook.SellOrders = append(orderBook.SellOrders, sellOrder)

	buyOrder := Order{
		ID:        "buy-1",
		Side:      SideBuy,
		Price:     100.0002,
		Quantity:  5,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}

	processOrder(buyOrder)

	// Should match due to small price difference
	if len(trades) != 1 {
		t.Errorf("Expected 1 trade, got %d", len(trades))
	}
}

func TestProcessOrder_EdgeCaseQuantities(t *testing.T) {
	setupTest()

	// Test with very small quantities
	sellOrder := Order{
		ID:        "sell-1",
		Side:      SideSell,
		Price:     100.0,
		Quantity:  1,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}
	orderBook.SellOrders = append(orderBook.SellOrders, sellOrder)

	buyOrder := Order{
		ID:        "buy-1",
		Side:      SideBuy,
		Price:     101.0,
		Quantity:  1,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}

	processOrder(buyOrder)

	// Should match with quantity 1
	if len(trades) != 1 {
		t.Errorf("Expected 1 trade, got %d", len(trades))
	}

	if trades[0].Quantity != 1 {
		t.Errorf("Expected trade quantity to be 1, got %d", trades[0].Quantity)
	}
}

func TestProcessOrder_HandlerReturnsCorrectTrades(t *testing.T) {
	setupTest()

	// Add a sell order to the book
	sellOrder := Order{
		ID:        "sell-1",
		Side:      SideSell,
		Price:     100.0,
		Quantity:  10,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}
	orderBook.SellOrders = append(orderBook.SellOrders, sellOrder)

	// Place a buy order via handler
	req := PlaceOrderRequest{
		Side:     SideBuy,
		Price:    101.0,
		Quantity: 5,
	}

	jsonData, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/api/place-order", bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	placeOrderHandler(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", response.Code)
	}

	var result PlaceOrderResponse
	json.Unmarshal(response.Body.Bytes(), &result)

	// Should return all trades (including the one we just created)
	if len(result.Trades) != 1 {
		t.Errorf("Expected 1 trade in response, got %d", len(result.Trades))
	}

	if result.Trades[0].Price != 100.0 {
		t.Errorf("Expected trade price to be 100.0, got %.2f", result.Trades[0].Price)
	}
}

func TestProcessOrder_HandlerReturnsTradesInOrder(t *testing.T) {
	setupTest()

	// Add multiple sell orders at different prices
	sellOrder1 := Order{
		ID:        "sell-1",
		Side:      SideSell,
		Price:     99.0,
		Quantity:  5,
		Status:    OrderStatusPending,
		CreatedAt: time.Now(),
	}
	sellOrder2 := Order{
		ID:        "sell-2",
		Side:      SideSell,
		Price:     100.0,
		Quantity:  5,
		Status:    OrderStatusPending,
		CreatedAt: time.Now().Add(time.Millisecond),
	}
	orderBook.SellOrders = append(orderBook.SellOrders, sellOrder1, sellOrder2)

	// Place a buy order that matches both
	req := PlaceOrderRequest{
		Side:     SideBuy,
		Price:    101.0,
		Quantity: 8,
	}

	jsonData, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/api/place-order", bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	placeOrderHandler(response, request)

	if response.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", response.Code)
	}

	var result PlaceOrderResponse
	json.Unmarshal(response.Body.Bytes(), &result)

	// Should return all trades in match order (including the 2 we just created)
	if len(result.Trades) != 2 {
		t.Errorf("Expected 2 trades in response, got %d", len(result.Trades))
	}

	// First trade should be at $99.00 (better price)
	if result.Trades[0].Price != 99.0 {
		t.Errorf("Expected first trade price to be 99.0, got %.2f", result.Trades[0].Price)
	}

	// Second trade should be at $100.00
	if result.Trades[1].Price != 100.0 {
		t.Errorf("Expected second trade price to be 100.0, got %.2f", result.Trades[1].Price)
	}
}
