# Economy & Trading Screens

This document covers all economy and trading-related UI screens in Terminal Velocity.

## Overview

**Screens**: 5
- Trading Screen
- Trading Enhanced Screen
- Cargo Screen
- Marketplace Screen
- Trade Routes Screen

**Purpose**: Handle all economic activities including commodity trading, cargo management, market analysis, and trade route planning.

**Source Files**:
- `internal/tui/trading.go` - Basic commodity trading interface
- `internal/tui/trading_enhanced.go` - Advanced trading with market analysis
- `internal/tui/cargo.go` - Cargo hold management
- `internal/tui/marketplace.go` - Multi-commodity marketplace view
- `internal/tui/traderoutes.go` - Trade route planning and profit calculations

---

## Trading Screen

### Source File
`internal/tui/trading.go`

### Purpose
Basic commodity exchange interface for buying and selling goods at planets. Simple, focused interface for quick trades.

### ASCII Prototype

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ COMMODITY EXCHANGE - Earth Station                           52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ COMMODITY          BUY PRICE   SELL PRICE   STOCK   YOUR CARGO       ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃ ▶ Food               45 cr       52 cr      High       10 tons       ┃  ┃
┃  ┃   Water              28 cr       35 cr      Med         0 tons       ┃  ┃
┃  ┃   Textiles          110 cr      125 cr      Low         0 tons       ┃  ┃
┃  ┃   Electronics       380 cr      425 cr      Med         5 tons       ┃  ┃
┃  ┃   Computers         890 cr      950 cr      Low         0 tons       ┃  ┃
┃  ┃   Weapons         1,200 cr    1,350 cr      Med         0 tons       ┃  ┃
┃  ┃   Medical Sup.      450 cr      490 cr      High        0 tons       ┃  ┃
┃  ┃   Luxury Goods    2,100 cr    2,300 cr      Low         0 tons       ┃  ┃
┃  ┃   Industrial       180 cr      205 cr      High        0 tons       ┃  ┃
┃  ┃   Minerals          95 cr      110 cr      High        0 tons       ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ SELECTED: Food                                                       ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃ Buy Price:  52 cr/ton             Sell Price: 45 cr/ton             ┃  ┃
┃  ┃ In Cargo:   10 tons               Available Space: 35 tons          ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃ Quantity: [____] tons                                                ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃ [ Buy ]  [ Sell ]  [ Max Buy ]  [ Sell All ]                        ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ TIP: Food is cheap here! Best profit selling at mining colonies.    ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [↑↓] Select  [B]uy  [S]ell  [M]ax Buy  [A]ll Sell  [ESC] Back              ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### Components
- **Commodity List**: All 15 tradeable commodities with prices and availability
- **Selection Panel**: Details of currently selected commodity
- **Quantity Input**: Field for specifying trade amount
- **Action Buttons**: Buy, sell, max buy, sell all operations
- **Trading Tip Box**: Contextual advice for profitable trades

### Key Bindings
- `↑`/`↓` or `J`/`K` - Navigate commodity list
- `B` - Buy commodity (prompted for quantity)
- `S` - Sell commodity (prompted for quantity)
- `M` - Max buy (fill cargo with selected commodity)
- `A` - Sell all of selected commodity
- `ESC` - Return to landing screen

### State Management

**Model Structure** (`tradingModel`):
```go
type tradingModel struct {
    planet           *models.Planet
    commodities      []*models.Commodity
    selectedIndex    int
    quantity         int
    playerShip       *models.Ship
    playerCredits    int
    inputMode        bool  // True when entering quantity
    width            int
    height           int
}
```

**Messages**:
- `tea.KeyMsg` - Keyboard input
- `commoditiesLoadedMsg` - Market data loaded
- `tradCompleteMsg` - Buy/sell transaction completed
- `tradeErrorMsg` - Insufficient funds or cargo space

### Data Flow
1. Load planet market data from `MarketRepository`
2. Display commodity list with current prices
3. User selects commodity and action (buy/sell)
4. Validate transaction (credits, cargo space)
5. Update ship cargo and player credits
6. Refresh market display
7. Record transaction in trade history

### Trading Mechanics

**Commodities** (15 total):
- Food, Water, Textiles, Electronics, Computers
- Weapons, Medical Supplies, Luxury Goods, Industrial Equipment
- Minerals, Metals, Fuel, Radioactives, Illegal Goods, Narcotics

**Pricing Factors**:
- **Tech Level**: Higher tech systems produce/demand different goods
- **Supply/Demand**: Prices fluctuate based on stock levels
- **Government**: Some governments restrict certain commodities
- **Random Variance**: ±10% price fluctuation for market dynamics

**Stock Levels**:
- High: >100 units available
- Medium: 50-100 units
- Low: <50 units
- Out of Stock: 0 units (cannot buy)

**Profit Calculation**:
```
Profit = (Sell Price - Buy Price) * Quantity
Profit Margin = ((Sell Price - Buy Price) / Buy Price) * 100%
```

### Trade Validation
- **Buy**: Check player has sufficient credits
- **Buy**: Check ship has cargo space
- **Sell**: Check player has commodity in cargo
- **Illegal Goods**: Check if legal in current system

### Related Screens
- **Landing Screen** - Return with `ESC`
- **Cargo Screen** - View full cargo manifest
- **Trading Enhanced** - Advanced market analysis
- **Trade Routes** - Plan profitable routes

---

## Trading Enhanced Screen

### Source File
`internal/tui/trading_enhanced.go`

### Purpose
Advanced trading interface with market trends, price history, profit calculators, and multi-commodity comparison.

### ASCII Prototype

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ ADVANCED TRADING - Earth Station                             52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ COMMODITIES                    ┃  ┃ FOOD - MARKET ANALYSIS          ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                ┃  ┃                                 ┃  ┃
┃  ┃ ▶ Food          45cr  ▲ High  ┃  ┃ Current Price: 45 cr/ton        ┃  ┃
┃  ┃   Water         28cr  ━ Med   ┃  ┃ Average Price: 42 cr/ton        ┃  ┃
┃  ┃   Textiles     110cr  ▼ Low   ┃  ┃ Price Trend: ▲ Rising +7%       ┃  ┃
┃  ┃   Electronics  380cr  ▲ Med   ┃  ┃                                 ┃  ┃
┃  ┃   Computers    890cr  ━ Low   ┃  ┃ Supply: High (234 tons)         ┃  ┃
┃  ┃   Weapons    1,200cr  ▲ Med   ┃  ┃ Demand: Medium                  ┃  ┃
┃  ┃   Medical      450cr  ▼ High  ┃  ┃                                 ┃  ┃
┃  ┃   Luxury     2,100cr  ━ Low   ┃  ┃ Best Profit At:                 ┃  ┃
┃  ┃   Industrial   180cr  ▲ High  ┃  ┃  1. Mars Colony    +28 cr/ton   ┃  ┃
┃  ┃   Minerals      95cr  ━ High  ┃  ┃  2. Wolf 359       +24 cr/ton   ┃  ┃
┃  ┃                                ┃  ┃  3. Sirius Station +19 cr/ton   ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ PRICE HISTORY (7 DAYS)                                               ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  52cr ┤                                                          ●   ┃  ┃
┃  ┃  50cr ┤                                                    ●   ●     ┃  ┃
┃  ┃  48cr ┤                                          ●   ●               ┃  ┃
┃  ┃  46cr ┤                                    ●                         ┃  ┃
┃  ┃  44cr ┤                          ●   ●                               ┃  ┃
┃  ┃  42cr ┤                    ●                                         ┃  ┃
┃  ┃  40cr ┤              ●                                               ┃  ┃
┃  ┃       └──────────────────────────────────────────────────────────    ┃  ┃
┃  ┃         Mon  Tue  Wed  Thu  Fri  Sat  Sun                           ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ PROFIT CALCULATOR                                                    ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  Buy: 35 tons @ 45 cr/ton = 1,575 cr                                ┃  ┃
┃  ┃  Sell at Mars @ 73 cr/ton = 2,555 cr                                ┃  ┃
┃  ┃  Profit: 980 cr (+62%)                                              ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  Quantity: [35] tons    [ Calculate ]  [ Execute Trade ]            ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [↑↓] Select  [Tab] Switch Panel  [T]rade  [R]outes  [H]istory  [ESC] Back  ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### Components
- **Commodity List**: Enhanced with trend indicators and stock levels
- **Market Analysis Panel**: Current prices, trends, best sell locations
- **Price History Chart**: 7-day price graph for selected commodity
- **Profit Calculator**: Calculate potential profit for planned trades
- **Trend Indicators**: ▲ rising, ▼ falling, ━ stable

### Key Bindings
- `↑`/`↓` or `J`/`K` - Navigate commodity list
- `Tab` - Switch between panels (commodities, analysis, calculator)
- `T` - Execute trade
- `R` - Open trade routes planner
- `H` - View full price history
- `ESC` - Return to landing screen

### State Management

**Model Structure** (`tradingEnhancedModel`):
```go
type tradingEnhancedModel struct {
    tradingModel              // Embed basic trading model
    priceHistory     map[string][]float64  // 7 days of prices per commodity
    trends           map[string]string     // "rising", "falling", "stable"
    profitCalc       *ProfitCalculation
    focusedPanel     int  // 0=commodities, 1=analysis, 2=calculator
}

type ProfitCalculation struct {
    Commodity      string
    BuyQuantity    int
    BuyPrice       float64
    SellLocation   string
    SellPrice      float64
    Profit         float64
    ProfitPercent  float64
}
```

**Messages**:
- All messages from basic trading model
- `priceHistoryLoadedMsg` - Historical data loaded
- `trendsCalculatedMsg` - Price trends analyzed
- `profitCalculatedMsg` - Profit calculation complete

### Data Flow
1. Load market data and price history
2. Calculate price trends (7-day average vs current)
3. Identify best sell locations for each commodity
4. Display enhanced analytics
5. User can calculate potential profits
6. Execute trades with full market knowledge

### Market Analysis Features

**Price Trends**:
- **Rising (▲)**: Price >5% above 7-day average
- **Falling (▼)**: Price >5% below 7-day average
- **Stable (━)**: Price within ±5% of average

**Best Profit Calculation**:
- Query all known systems for commodity prices
- Calculate profit margin: (sell_price - buy_price) / buy_price
- Factor in distance (fuel costs)
- Rank by net profit potential

**Price History**:
- Store last 7 days of prices per commodity
- Display as ASCII line chart
- Identify patterns (seasonal, event-driven, etc.)

### Advanced Features
- **Multi-commodity comparison**: Compare profitability across all commodities
- **Route suggestions**: Recommend profitable trade routes
- **Market predictions**: Basic trend forecasting
- **Transaction history**: Track your past trades

### Related Screens
- **Trading Screen** - Simpler interface option
- **Trade Routes** - Full route planning
- **Cargo Screen** - Manage inventory

---

## Cargo Screen

### Source File
`internal/tui/cargo.go`

### Purpose
Cargo hold management - view inventory, jettison items, and monitor cargo capacity.

### ASCII Prototype

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ CARGO HOLD - Corvette "Starhawk"                             52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ CARGO CAPACITY: 15/50 tons (30%)                 [████░░░░░░░░░░░░]  ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ COMMODITY           QUANTITY    VALUE/TON    TOTAL VALUE              ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃ ▶ Food              10 tons      52 cr        520 cr                ┃  ┃
┃  ┃   Electronics        5 tons     425 cr      2,125 cr               ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  [Empty cargo space]                                                 ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  Total Cargo Value: 2,645 credits                                   ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ SELECTED: Food (10 tons)                                             ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  Acquired at: Earth Station                                          ┃  ┃
┃  ┃  Purchase Price: 45 cr/ton                                           ┃  ┃
┃  ┃  Current Value: 52 cr/ton (here)                                     ┃  ┃
┃  ┃  Potential Profit: +70 cr (+15.6%) if sold here                     ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  Jettison Quantity: [____] tons                                      ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  [ Jettison Selected ]  [ Jettison All ]                            ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  ⚠️ Warning: Jettisoned cargo is permanently lost!                   ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [↑↓] Select  [J]ettison  [A]ll Jettison  [S]ort  [ESC] Back                ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### Components
- **Capacity Bar**: Visual representation of cargo hold usage
- **Cargo List**: All commodities currently in hold
- **Detail Panel**: Information about selected cargo item
- **Jettison Controls**: Remove items from cargo hold
- **Value Summary**: Total cargo worth

### Key Bindings
- `↑`/`↓` or `J`/`K` - Navigate cargo list
- `J` - Jettison selected commodity (prompted for quantity)
- `A` - Jettison all of selected commodity
- `S` - Sort cargo (by value, quantity, name)
- `ESC` - Return to previous screen

### State Management

**Model Structure** (`cargoModel`):
```go
type cargoModel struct {
    ship           *models.Ship
    cargoItems     []*CargoItem
    selectedIndex  int
    jettisonQty    int
    sortBy         string  // "name", "quantity", "value"
    width          int
    height         int
}

type CargoItem struct {
    Commodity      *models.Commodity
    Quantity       int
    PurchasePrice  float64
    PurchaseLocation string
    CurrentValue   float64
}
```

**Messages**:
- `tea.KeyMsg` - Keyboard input
- `cargoLoadedMsg` - Cargo manifest loaded
- `jettisonCompleteMsg` - Items jettisoned
- `cargoSortedMsg` - Cargo list re-sorted

### Data Flow
1. Load ship cargo from database
2. Display cargo items with current market values
3. Calculate potential profit/loss for each item
4. User can jettison items to free space
5. Update ship cargo in database
6. Refresh cargo display

### Cargo Management Features

**Capacity Tracking**:
- Current cargo weight vs maximum capacity
- Visual progress bar
- Warning when approaching capacity
- Cannot pick up cargo when full

**Value Tracking**:
- Purchase price recorded when bought
- Current market value displayed
- Profit/loss calculation
- Total cargo value summary

**Sorting Options**:
- By name (alphabetical)
- By quantity (most/least)
- By value (highest/lowest)
- By profit potential (best gains)

**Jettison Mechanics**:
- Permanently removes items from cargo
- Frees cargo space immediately
- No refund or recovery
- Confirmation required for safety
- Useful when cargo space needed urgently

### Related Screens
- **Trading Screen** - Sell cargo commodities
- **Space View** - Return to flight
- **Ship Management** - View cargo capacity stats

---

## Marketplace Screen

### Source File
`internal/tui/marketplace.go`

### Purpose
Comprehensive market overview showing all commodities, multiple systems, and comparative pricing.

### ASCII Prototype

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ GALACTIC MARKETPLACE                                         52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ COMMODITY PRICES ACROSS SYSTEMS                                       ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃ Commodity     Earth    Mars    Alpha C  Sirius   Barnard   Wolf359 ┃  ┃
┃  ┃ ──────────────────────────────────────────────────────────────────  ┃  ┃
┃  ┃ ▶Food          45cr    73cr     68cr     52cr     78cr      81cr   ┃  ┃
┃  ┃   Water        28cr    55cr     48cr     31cr     62cr      58cr   ┃  ┃
┃  ┃   Textiles    110cr    98cr    145cr    122cr    105cr     128cr   ┃  ┃
┃  ┃   Electronics 380cr   425cr    398cr    445cr    412cr     390cr   ┃  ┃
┃  ┃   Computers   890cr   950cr    875cr  1,020cr    945cr     915cr   ┃  ┃
┃  ┃   Weapons   1,200cr 1,350cr  1,280cr  1,450cr  1,320cr   1,380cr   ┃  ┃
┃  ┃   Medical     450cr   490cr    468cr    512cr    485cr     475cr   ┃  ┃
┃  ┃   Luxury    2,100cr 2,300cr  2,250cr  2,480cr  2,350cr   2,290cr   ┃  ┃
┃  ┃   Industrial  180cr   205cr    192cr    218cr    198cr     210cr   ┃  ┃
┃  ┃   Minerals     95cr   110cr    102cr    118cr    108cr     115cr   ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃ Legend: Green=Best Buy  Red=Best Sell  Yellow=Your Location        ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ BEST TRADE OPPORTUNITIES                                             ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  1. Food: Buy at Earth (45cr) → Sell at Wolf 359 (81cr) = +80%     ┃  ┃
┃  ┃     Distance: 6.2 LY, Fuel Cost: 62 units, Net Profit: 1,260cr/35t ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  2. Textiles: Buy at Mars (98cr) → Sell at Alpha C (145cr) = +48%  ┃  ┃
┃  ┃     Distance: 3.8 LY, Fuel Cost: 38 units, Net Profit: 1,645cr/35t ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  3. Computers: Buy at Alpha C (875cr) → Sell at Sirius (1,020cr)   ┃  ┃
┃  ┃     Distance: 5.1 LY, Fuel Cost: 51 units, Net Profit: 5,075cr/35t ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ FILTERS: [All Commodities ▼]  [All Systems ▼]  [Sort: Profit ▼]    ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [↑↓] Select  [Tab] Switch View  [F]ilter  [R]efresh  [ESC] Back            ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### Components
- **Price Matrix**: Multi-system commodity price comparison
- **Trade Opportunities**: Ranked list of best profit routes
- **Filter Controls**: Narrow down display by commodity or system
- **Color Coding**: Visual indicators for best buy/sell locations

### Key Bindings
- `↑`/`↓` - Navigate commodity list or opportunities
- `Tab` - Switch between price matrix and opportunities view
- `F` - Open filter options
- `R` - Refresh market data
- `Enter` - View details of selected trade opportunity
- `ESC` - Close marketplace

### State Management

**Model Structure** (`marketplaceModel`):
```go
type marketplaceModel struct {
    systems          []*models.StarSystem
    commodities      []*models.Commodity
    priceMatrix      map[string]map[string]float64  // [commodity][system]price
    opportunities    []*TradeOpportunity
    selectedIndex    int
    viewMode         string  // "matrix" or "opportunities"
    filters          *MarketFilters
    width            int
    height           int
}

type TradeOpportunity struct {
    Commodity      string
    BuySystem      string
    BuyPrice       float64
    SellSystem     string
    SellPrice      float64
    ProfitMargin   float64
    Distance       float64
    FuelCost       int
    NetProfit      float64
}

type MarketFilters struct {
    CommodityType  string  // "all", "food", "tech", etc.
    MinProfit      float64
    MaxDistance    float64
    OnlyVisited    bool
}
```

**Messages**:
- `tea.KeyMsg` - Keyboard input
- `marketDataLoadedMsg` - Price matrix loaded
- `opportunitiesCalculatedMsg` - Trade opportunities ranked
- `marketRefreshedMsg` - Data refreshed from server

### Data Flow
1. Query market prices from all known systems
2. Build price comparison matrix
3. Calculate all possible trade routes
4. Rank by profit margin and net profit
5. Display top opportunities
6. Update on commodity/system selection
7. Refresh periodically (markets change)

### Market Analysis

**Price Matrix**:
- Shows prices for all commodities across multiple systems
- Color-coded: green (cheap), red (expensive), yellow (current location)
- Sortable by system or commodity
- Quick visual comparison

**Trade Opportunities**:
- All possible buy-low, sell-high combinations
- Sorted by profit potential
- Factors in distance (fuel costs)
- Net profit = gross profit - fuel costs
- Updates in real-time as markets change

**Filtering**:
- By commodity type (food, tech, industrial, etc.)
- By minimum profit threshold
- By maximum distance willing to travel
- By only visited systems (known markets)

### Related Screens
- **Trading Screen** - Execute trades at current location
- **Trade Routes** - Plan multi-hop trading routes
- **Navigation** - Jump to profitable systems

---

## Trade Routes Screen

### Source File
`internal/tui/traderoutes.go`

### Purpose
Multi-hop trade route planning with profit optimization and route visualization.

### ASCII Prototype

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ TRADE ROUTE PLANNER                                          52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ ROUTE: Earth → Mars → Alpha Centauri → Earth                         ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  LEG 1: Earth → Mars (0.8 LY)                                        ┃  ┃
┃  ┃  ────────────────────────────────────────────────────────────────    ┃  ┃
┃  ┃  Buy:  Food (35t @ 45cr) = 1,575 cr                                  ┃  ┃
┃  ┃  Sell: Food (35t @ 73cr) = 2,555 cr                                  ┃  ┃
┃  ┃  Profit: +980 cr    Fuel: 8 units                                    ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  LEG 2: Mars → Alpha Centauri (3.2 LY)                               ┃  ┃
┃  ┃  ────────────────────────────────────────────────────────────────    ┃  ┃
┃  ┃  Buy:  Textiles (35t @ 98cr) = 3,430 cr                              ┃  ┃
┃  ┃  Sell: Textiles (35t @ 145cr) = 5,075 cr                             ┃  ┃
┃  ┃  Profit: +1,645 cr    Fuel: 32 units                                 ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  LEG 3: Alpha Centauri → Earth (3.2 LY)                              ┃  ┃
┃  ┃  ────────────────────────────────────────────────────────────────    ┃  ┃
┃  ┃  Buy:  Computers (15t @ 875cr) = 13,125 cr                           ┃  ┃
┃  ┃  Sell: Computers (15t @ 950cr) = 14,250 cr                           ┃  ┃
┃  ┃  Profit: +1,125 cr    Fuel: 32 units                                 ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ ROUTE SUMMARY                                                        ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  Total Distance: 7.2 light years                                     ┃  ┃
┃  ┃  Total Fuel Cost: 72 units (720 cr @ 10cr/unit)                     ┃  ┃
┃  ┃  Time Required: ~45 minutes                                          ┃  ┃
┃  ┃  Initial Investment: 1,575 cr                                        ┃  ┃
┃  ┃  Total Revenue: 21,880 cr                                            ┃  ┃
┃  ┃  Total Profit: 3,750 cr                                              ┃  ┃
┃  ┃  Net Profit (after fuel): 3,030 cr (+192%)                           ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  Risk Level: Low     Jumps: 3     Pirate Chance: 30%                ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  [ Start Route ]  [ Optimize ]  [ Save Route ]  [ Clear ]           ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [A]dd Stop  [D]elete Stop  [O]ptimize  [S]ave  [L]oad Route  [ESC] Cancel  ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

### Components
- **Route Path**: List of systems in planned route
- **Leg Details**: Buy/sell actions for each route segment
- **Route Summary**: Total profit, fuel, time, risk
- **Action Buttons**: Start, optimize, save route
- **Route Builder**: Add/remove stops

### Key Bindings
- `A` - Add system to route
- `D` - Delete selected system from route
- `O` - Auto-optimize route for maximum profit
- `S` - Save route for later use
- `L` - Load saved route
- `Enter` - Start executing route
- `ESC` - Cancel route planning

### State Management

**Model Structure** (`tradeRoutesModel`):
```go
type tradeRoutesModel struct {
    route           *TradeRoute
    savedRoutes     []*TradeRoute
    selectedLeg     int
    optimizing      bool
    width           int
    height          int
}

type TradeRoute struct {
    Name          string
    Legs          []*RouteLeg
    TotalDistance float64
    TotalFuel     int
    TotalProfit   float64
    NetProfit     float64
    EstimatedTime int  // minutes
    RiskLevel     string
}

type RouteLeg struct {
    FromSystem   string
    ToSystem     string
    Distance     float64
    FuelCost     int
    BuyCommodity string
    BuyQuantity  int
    BuyPrice     float64
    SellPrice    float64
    LegProfit    float64
}
```

**Messages**:
- `tea.KeyMsg` - Keyboard input
- `routeCalculatedMsg` - Route metrics computed
- `routeOptimizedMsg` - Route automatically optimized
- `routeSavedMsg` - Route saved to database
- `routeLoadedMsg` - Saved route loaded

### Data Flow
1. User selects destination systems
2. Calculate best commodities to trade at each leg
3. Compute total distance, fuel, time
4. Calculate cumulative profit
5. Display route details
6. Optional: Auto-optimize for best profit
7. Execute route or save for later

### Route Planning Features

**Manual Route Building**:
- Add systems one by one
- App suggests best commodity for each leg
- Visual feedback on profitability
- Can rearrange order of stops

**Auto-Optimization**:
- Finds most profitable commodity at each leg
- Reorders stops to maximize profit
- Considers cargo capacity constraints
- Minimizes fuel costs
- Avoids dangerous systems (if desired)

**Route Metrics**:
- **Total Distance**: Sum of all jumps
- **Fuel Cost**: Distance * 10 units/LY * fuel price
- **Time Estimate**: Based on average trade time
- **Profit**: Sum of all leg profits minus costs
- **Risk Level**: Based on pirate activity in systems

**Saved Routes**:
- Save favorite trading routes
- Name and description
- Quick-load for repeated runs
- Share routes with faction members

**Risk Assessment**:
- Low: Safe systems, minimal pirate activity
- Medium: Some pirate presence
- High: Known pirate territory
- Extreme: Active war zones or hostile factions

### Route Execution

When starting a route:
1. Navigate to first system
2. Auto-suggest buying correct commodity and quantity
3. Navigate to next system
4. Auto-suggest selling
5. Repeat for all legs
6. Track progress through route
7. Update profit in real-time

### Related Screens
- **Navigation** - Jump to route systems
- **Trading** - Execute buy/sell at each stop
- **Marketplace** - View all market data

---

## Implementation Notes

### Database Integration
Economy screens interact with these repositories:
- `database.MarketRepository` - Commodity prices and availability
- `database.ShipRepository` - Cargo capacity and current cargo
- `database.PlayerRepository` - Credits and transaction history
- `database.SystemRepository` - System locations and tech levels

### Economic Model

**Supply & Demand**:
- Prices fluctuate based on market activity
- High demand → higher sell prices
- High supply → lower buy prices
- Player trades affect market (large volumes)

**Tech Level Impact**:
- Low tech (1-3): Food, water, textiles, minerals
- Medium tech (4-6): Electronics, industrial, weapons
- High tech (7-9): Computers, luxury goods, advanced tech
- Tech level affects production and demand

**Government Impact**:
- Free Market: No restrictions, best prices
- Corporate: Higher prices, controlled markets
- Pirate: Illegal goods available, risky
- Military: Weapons restricted, defense equipment available

**Dynamic Events**:
- Famines increase food prices
- Wars increase weapons demand
- Plagues affect medical supplies
- Tech booms affect computer prices

### Price Calculation

Base price calculation:
```go
basePrice := commodity.BasePrice
techModifier := calculateTechModifier(system.TechLevel, commodity)
supplyModifier := calculateSupplyDemand(commodity, system)
randomVariance := rand.Float64() * 0.1 - 0.05  // ±5%

finalPrice := basePrice * techModifier * supplyModifier * (1 + randomVariance)
```

### Transaction Validation

All trades validated for:
- Sufficient player credits (buy)
- Sufficient cargo space (buy)
- Commodity in cargo (sell)
- Legal status of commodity in system
- Market has stock available (buy)

### Performance Considerations

**Caching**:
- Cache market data for 5 minutes
- Refresh on explicit user request
- Background updates for real-time changes

**Optimization**:
- Lazy-load price history (only when viewed)
- Limit trade route calculations to reachable systems
- Index commodity prices by system for fast lookups

### Testing

Test files:
- `internal/tui/trading_test.go` - Trading screen tests
- `internal/tui/cargo_test.go` - Cargo management tests
- Unit tests for profit calculations
- Integration tests for full trade flows

---

**Last Updated**: 2025-11-15
**Document Version**: 1.0.0
