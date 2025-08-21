# Valhalla Order Book

A simple order book implementation in Go that follows standard exchange matching engine principles.

## Features

- **Price-Time Priority**: Orders are matched based on best price first, then oldest time
- **Limit Order Matching**: Incoming orders are matched against resting orders in the book
- **Trade Pricing**: Trades execute at the resting order's price (maker-taker model)
- **Partial Fills**: Orders can be partially filled and remain in the book
- **REST API**: Simple HTTP endpoints for placing orders and viewing the book

## Order Book Rules

1. **If the incoming order still has quantity, add it to the book as its limit price with newest time priority**
2. **Fill the oldest orders at the best price level first**
3. **Trade pricing = resting book order's price**

## API Endpoints

### Place Order
```
POST /api/place-order
Content-Type: application/json

{
  "side": "buy" | "sell",
  "price": 100.50,
  "quantity": 100
}
```

Response:
```json
{
  "order_id": "uuid",
  "trades": [
    {
      "id": "trade-uuid",
      "maker_id": "resting-order-id",
      "taker_id": "incoming-order-id",
      "price": 100.00,
      "quantity": 50,
      "created_at": "2024-01-01T12:00:00Z"
    }
  ]
}
```

**Note**: The `trades` field returns ALL executed trades in match order, not just the trades from the current order.

### Get All Orders
```
GET /api/orders
```

### Get All Trades
```
GET /api/trades
```

### Get Order Book
```
GET /api/orderbook
```

## Running the Server

```bash
go run main.go
```

The server will start on port 8080.

## Testing

Run the test script to see the order book in action:

```bash
go run test_orderbook.go
```

This will demonstrate:
1. Placing a sell order at $100
2. Placing a buy order at $95 (no match)
3. Placing a buy order at $105 (matches at $100)
4. Placing a sell order at $98 (matches with remaining buy order)
5. Viewing the final order book state

## Example Scenario

1. **Sell Order**: 50 units at $100 → Added to sell side of book
2. **Buy Order**: 30 units at $95 → Added to buy side of book (no match)
3. **Buy Order**: 20 units at $105 → Matches with sell order at $100
   - Trade: 20 units at $100 (resting order's price)
   - Sell order: 30 units remaining at $100
4. **Sell Order**: 15 units at $98 → Matches with buy order at $95
   - Trade: 15 units at $95 (resting order's price)
   - Buy order: 15 units remaining at $95

## Implementation Details

- **Order Book Structure**: Separate arrays for buy and sell orders, sorted by price and time
- **Matching Logic**: Incoming orders are matched against the opposite side of the book
- **Price Priority**: Best prices are matched first (highest for buys, lowest for sells)
- **Time Priority**: Within the same price level, oldest orders are matched first
- **Trade Execution**: Trades execute at the resting order's price (maker-taker model)
