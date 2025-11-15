# Terminal Velocity - Economy Balance

## Overview
This document tracks the economic balance of Terminal Velocity, including trade routes, pricing, and profitability analysis.

## Current Economic Parameters

### Starting Conditions
- **Starting Credits**: 10,000 cr (default from schema)
- **Starting Ship**: Shuttle (25,000 cr value)
  - Cargo Space: 20 units
  - Fuel: 100 units
  - Hull: 100 HP

### Commodity Base Prices

#### Food Category (Essential)
| Commodity | Base Price | Tech Level | Notes |
|-----------|------------|------------|-------|
| Food | 50 | 1 | Stable demand |
| Water | 30 | 1 | Essential |
| Textiles | 60 | 2 | Low tech |
| Livestock | 120 | 1 | Moderate value |

#### Electronics Category (High Tech)
| Commodity | Base Price | Tech Level | Notes |
|-----------|------------|------------|-------|
| Electronics | 200 | 4 | Good margins |
| Computers | 350 | 5 | High value |
| Robotics | 600 | 7 | Very high value |
| AI Cores | 1,200 | 9 | Premium tech |

#### Weapons Category (Restricted)
| Commodity | Base Price | Tech Level | Illegal In |
|-----------|------------|------------|------------|
| Weapons | 500 | 3 | pacifist_union |
| Ammunition | 180 | 2 | pacifist_union |
| Explosives | 400 | 3 | pacifist_union |
| Military Hardware | 900 | 6 | pacifist_union, independent |

#### Medical Category (High Demand)
| Commodity | Base Price | Tech Level | Notes |
|-----------|------------|------------|-------|
| Medicine | 150 | 5 | Steady demand |
| Medical Equipment | 380 | 6 | Professional use |
| Vaccines | 280 | 7 | Crisis pricing |
| Bio-Organs | 2,500 | 8 | Premium medical |

#### Luxury Goods (High Margin)
| Commodity | Base Price | Tech Level | Notes |
|-----------|------------|------------|-------|
| Luxury Goods | 400 | 6 | Wealthy markets |
| Jewelry | 800 | 4 | Stable value |
| Art | 1,000 | 5 | Cultural items |
| Exotic Animals | 1,500 | 3 | Rare creatures |

#### Industrial (Bulk)
| Commodity | Base Price | Tech Level | Notes |
|-----------|------------|------------|-------|
| Machinery | 250 | 4 | Industrial demand |
| Construction Materials | 90 | 2 | High volume |
| Power Cells | 320 | 5 | Energy storage |
| Industrial Chemicals | 150 | 4 | Manufacturing |

#### Raw Materials / Ore (Mining)
| Commodity | Base Price | Tech Level | Notes |
|-----------|------------|------------|-------|
| Metal Ore | 80 | 2 | Common resource |
| Precious Metals | 450 | 3 | High value |
| Crystals | 350 | 5 | Tech applications |
| Radioactive Materials | 600 | 6 | Restricted, dangerous |

#### Contraband (High Risk/High Reward)
| Commodity | Base Price | Tech Level | Illegal In | Risk Level |
|-----------|------------|------------|------------|------------|
| Narcotics | 800 | 4 | federation, republic, corporate | High |
| Slaves | 1,500 | 1 | Most systems | Very High |
| Stolen Goods | 600 | 1 | federation, republic, corporate | High |
| Alien Artifacts | 3,000 | 1 | federation | Moderate |
| Military Intel | 5,000 | 7 | Most systems | Extreme |

## Pricing Mechanics

### Tech Level Modifiers

#### For Selling (High tech planets pay less for low tech goods)
- Planet tech > commodity tech: `1.0 - (diff * 0.1)`
- Planet tech < commodity tech: `1.0 + (abs(diff) * 0.15)`

#### For Buying (High tech planets sell high tech goods cheaper)
- Planet tech >= commodity tech: `1.0 - (diff * 0.05)`
- Planet tech < commodity tech: `1.0 + (abs(diff) * 0.2)`

### Supply/Demand Modifiers
- **No stock, high demand**: 2.0 + (demand * 0.1)
- **No demand, high stock**: 0.3
- **Oversupply** (stock > demand): 1.0 - min(ratio-1.0, 1.0) * 0.4 (floor 30%)
- **Undersupply** (demand > stock): 1.0 + (1.0-ratio) * 0.8 (cap 250%)

### Buy/Sell Spread
- Planet buy price (what they pay you): 60-80% of sell price
- This creates a natural margin for traders

## Profitable Trade Routes

### Route 1: Basic Food Loop (Low Risk, Low Reward)
**Start**: Low tech agricultural world (Tech 1-2)
- **Buy**: Food (30-50 cr), Water (20-30 cr)
- **Sell at**: High tech industrial world (Tech 6+)
- **Expected Profit**: 20-30 cr per unit
- **Return Trip**: Buy Electronics (150-200 cr), sell at ag world
- **Round Trip Profit**: 50-80 cr per unit
- **Risk**: Very Low
- **Capital Needed**: 1,000-2,000 cr for 20 units

### Route 2: Tech Equipment (Medium Risk, Good Reward)
**Start**: High tech world (Tech 7+)
- **Buy**: Computers (280-320 cr), Electronics (160-180 cr)
- **Sell at**: Mid tech world (Tech 3-5)
- **Expected Profit**: 100-150 cr per unit
- **Return Trip**: Buy Industrial goods, Machinery
- **Round Trip Profit**: 150-250 cr per unit
- **Risk**: Low
- **Capital Needed**: 5,000-8,000 cr

### Route 3: Luxury Goods (High Margin)
**Start**: Wealthy high tech world
- **Buy**: Art (800-900 cr), Jewelry (650-750 cr)
- **Sell at**: Wealthy independent systems
- **Expected Profit**: 200-400 cr per unit
- **Risk**: Low (but requires capital)
- **Capital Needed**: 15,000+ cr

### Route 4: Contraband (High Risk, High Reward)
**Start**: Criminal/pirate bases
- **Buy**: Narcotics (600-700 cr)
- **Sell at**: Systems where legal or unpoliced
- **Expected Profit**: 300-600 cr per unit (50-100% margin)
- **Risk**: Very High (confiscation, fines, criminal record)
- **Capital Needed**: 10,000+ cr
- **Notes**: Requires finding markets where legal

### Route 5: Medical Emergency (Event-Based)
**Trigger**: Disease outbreak, war zone
- **Buy**: Medicine (120-150 cr), Vaccines (220-260 cr)
- **Sell at**: Crisis zone
- **Expected Profit**: 200-500 cr per unit
- **Risk**: Medium (dangerous areas)
- **Capital Needed**: 5,000+ cr

## Economic Balance Recommendations

### 1. Starting Resources
**Current**:
- 10,000 cr starting capital
- Shuttle ship (20 cargo)

**Recommended**:
- Keep 10,000 cr (allows 1-2 basic trade runs)
- Shuttle is appropriate starting ship
- First profitable loop should net 500-1,000 cr

### 2. Price Adjustments

#### Increase Contraband Risk/Reward
- Contraband should offer 80-150% profit margins
- But add confiscation mechanics in future
- Increase base prices by 20%:
  - Narcotics: 800 → 1,000
  - Slaves: 1,500 → 2,000
  - Stolen Goods: 600 → 800
  - Military Intel: 5,000 → 7,500

#### Adjust Tech Level Multipliers
Current modifiers are good, but can be tuned:
- **Selling to high tech**: Keep current (0.1 per level decrease)
- **Buying from high tech**: Increase benefit to 0.07 per level (from 0.05)
- This makes high-tech hubs better sources

### 3. Starting Ship Cargo Optimization
With 20 cargo space:
- Basic food run: 20 units @ 50cr = 1,000 cr investment, ~500 cr profit
- This is reasonable for first run
- After 3-4 runs, can upgrade to Courier (40 cargo)

### 4. Trade Route Profitability Targets

| Experience Level | Profit per Trip | Trip Investment | ROI |
|-----------------|-----------------|-----------------|-----|
| Beginner (Shuttle) | 500-1,000 cr | 1,000-2,000 cr | 50-100% |
| Intermediate (Courier) | 2,000-5,000 cr | 5,000-10,000 cr | 40-80% |
| Advanced (Hauler) | 15,000-30,000 cr | 30,000-50,000 cr | 50-60% |
| Expert (Contraband) | 20,000-50,000 cr | 50,000-100,000 cr | 40-100% |

### 5. Risk vs Reward Balance

**Legal Trade**:
- ROI: 40-80%
- Risk: Very Low
- Time: Moderate

**Contraband**:
- ROI: 80-150%
- Risk: High (confiscation, criminal status)
- Time: High (finding legal markets)

**Recommendation**: Contraband should offer 2x the profit but with significant risk

## Ship Progression Economics

### Entry Level (0-50,000 cr)
- **Shuttle**: 25,000 cr, 20 cargo
- **Courier**: 50,000 cr, 40 cargo
- **Target**: Basic trade runs, 500-2,000 cr per trip

### Mid Level (50,000-300,000 cr)
- **Hauler**: 150,000 cr, 100 cargo
- **Interceptor**: 75,000 cr, 10 cargo (combat)
- **Target**: Bulk trading, 5,000-15,000 cr per trip

### Advanced Level (300,000-1,000,000 cr)
- **Bulk Freighter**: 300,000 cr, 200 cargo
- **Gunship**: 250,000 cr, 50 cargo (armed trading)
- **Target**: Major trade runs, 20,000-50,000 cr per trip

### Expert Level (1,000,000+ cr)
- **Destroyer**: 750,000 cr (combat/escort)
- **Cruiser**: 1,500,000 cr (fleet command)
- **Target**: Rare goods, contraband, missions

## Fuel Economics
- Jump cost: 10 fuel per system
- Fuel price: ~50 cr per unit (varies by location)
- Cost per jump: ~500 cr
- This should be factored into profit calculations

## Market Dynamics

### Initial Stock/Demand Generation
- **Stock**: Base 100 + (tech_diff * 20) ± 30%
- **Demand**: (population / 1M) * category_multiplier * tech_factor ± 40%

### Category Demand Multipliers
- Food: 150 (essential)
- Medical: 100 (steady)
- Electronics: 80
- Industrial: 70
- Ore: 60
- Weapons: 50 (niche)
- Luxuries: 30 (low volume)
- Contraband: 20 (limited)

### Market Recovery
- Stock regenerates at 5% per hour toward target
- Demand normalizes at 5% per hour toward target
- Random events: 5% chance per hour (supply shocks, demand surges)

## Testing Scenarios

### Scenario 1: Beginner Trade Loop
1. Start with 10,000 cr, Shuttle (20 cargo)
2. Buy Food at agricultural world (50 cr/unit) = 1,000 cr for 20 units
3. Travel to high-tech world (500 cr fuel)
4. Sell Food (tech bonus ~2.0x) = ~2,000 cr revenue
5. Net profit: ~500 cr (50% ROI after fuel)
6. **Result**: Profitable but slow progression

### Scenario 2: Tech Trading
1. Player has 20,000 cr, Courier (40 cargo)
2. Buy Electronics at high-tech hub (180 cr/unit) = 7,200 cr
3. Sell at mid-tech world (tech penalty ~1.5x) = ~10,800 cr
4. Net profit: ~2,600 cr (36% ROI after fuel)
5. **Result**: Better profit, requires more capital

### Scenario 3: Contraband Run
1. Player has 50,000 cr, Gunship (50 cargo, armed)
2. Buy Narcotics at pirate base (700 cr/unit) = 35,000 cr
3. Sell at independent system where legal = ~70,000 cr
4. Net profit: ~34,000 cr (97% ROI after fuel/risk)
5. **Result**: Very profitable but risky

## Conclusion

The current economy is reasonably balanced for the following player progression:

1. **Hours 0-2**: Starting runs with Shuttle, learning mechanics, 500-1,000 cr/trip
2. **Hours 2-5**: Upgraded to Courier, better routes, 2,000-5,000 cr/trip
3. **Hours 5-10**: Hauler or armed ship, bulk trading or contraband, 10,000+ cr/trip
4. **Hours 10+**: Advanced ships, rare goods, missions, varied gameplay

**Key Adjustments Needed**:
1. Increase contraband base prices by 20-25%
2. Improve high-tech buy price modifier to 0.07
3. Ensure starting ship comes with some cargo for first trade
4. Add fuel cost awareness to trading UI

**Balance Target**: Player should be able to:
- Make first profit within 10 minutes
- Upgrade ship within 1-2 hours
- Access mid-tier ships within 3-5 hours
- Feel progression throughout 10+ hours of gameplay
