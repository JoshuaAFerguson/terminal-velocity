# Terminal Velocity - UI Prototypes (Escape Velocity Style)

Based on the classic Escape Velocity game interface, adapted for terminal/ASCII.

## 1. Main Space View (In-Flight)

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ TERMINAL VELOCITY v1.0          [Sol System]          Shields: ████████░░ 80%┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃    ╔═══════════════════════════════════════════════════════════════╗        ┃
┃    ║                                                               ║        ┃
┃    ║                          *                                    ║  ┏━━━━━━━━━━━━━┓
┃    ║                                                               ║  ┃   RADAR     ┃
┃    ║             *                    ⊕ Earth                      ║  ┃             ┃
┃    ║                                                               ║  ┃      *      ┃
┃    ║                                                               ║  ┃             ┃
┃    ║        *                                                      ║  ┃   ⊕    ◆    ┃
┃    ║                                 △                             ║  ┃        ▲    ┃
┃    ║                                You                            ║  ┃      *      ┃
┃    ║                                                               ║  ┃             ┃
┃    ║                                             ◆ Pirate          ║  ┗━━━━━━━━━━━━━┛
┃    ║           ⊕ Mars                                              ║
┃    ║                                                               ║  ┏━━━━━━━━━━━━━┓
┃    ║  *                                                            ║  ┃   STATUS    ┃
┃    ║                                                               ║  ┣━━━━━━━━━━━━━┫
┃    ║                                                       *       ║  ┃ Hull: ██████┃
┃    ║                                                               ║  ┃       100%  ┃
┃    ╚═══════════════════════════════════════════════════════════════╝  ┃ Fuel: ████░░┃
┃                                                                        ┃       67%   ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃ Speed: 340  ┃
┃  ┃ TARGET: Pirate Viper    ┃  ┃ CARGO: 15/50 tons                ┃  ┃ Credits:    ┃
┃  ┃ Distance: 2,340 km      ┃  ┃ ▪ Food (10t)  ▪ Electronics (5t) ┃  ┃  52,400 cr  ┃
┃  ┃ Shields: 45%            ┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┗━━━━━━━━━━━━━┛
┃  ┃ Attitude: Hostile       ┃                                                        ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━┛                                                        ┃
┃                                                                                     ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [L]and  [J]ump  [T]arget  [F]ire  [H]ail  [M]ap  [I]nfo  [ESC] Menu              ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

**Key:**
- `△` = Your ship
- `⊕` = Planet
- `◆` = Hostile ship
- `*` = Stars/background
- `▲` = Your ship on radar
- Main viewport shows top-down space view
- Radar shows relative positions
- Status panels show ship info
- Target info for selected ship
- Cargo manifest
- Bottom command bar


## 2. Planetary Landing Screen

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ EARTH STATION                    United Earth                  52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃    ╔═══════════════════════════════════════════════════════════════╗        ┃
┃    ║                                                               ║        ┃
┃    ║            Welcome to Earth Station, Commander.               ║        ┃
┃    ║                                                               ║        ┃
┃    ║         [ASCII art of planet/station could go here]           ║        ┃
┃    ║                       _______________                         ║        ┃
┃    ║                      /               \                        ║        ┃
┃    ║                     /    ⊕  EARTH     \                       ║        ┃
┃    ║                    |  (Terran Alliance)|                      ║        ┃
┃    ║                     \     Pop: 8.2B    /                      ║        ┃
┃    ║                      \_____    _______/                       ║        ┃
┃    ║                        /   \__/   \                           ║        ┃
┃    ║                       /  Station   \                          ║        ┃
┃    ║                       \____________/                          ║        ┃
┃    ║                                                               ║        ┃
┃    ║                                                               ║        ┃
┃    ╚═══════════════════════════════════════════════════════════════╝        ┃
┃                                                                              ┃
┃    ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓    ┃
┃    ┃  AVAILABLE SERVICES:       ┃  ┃  SHIP STATUS:                   ┃    ┃
┃    ┃                            ┃  ┃                                 ┃    ┃
┃    ┃  [C] Commodity Exchange    ┃  ┃  Ship: Corvette "Starhawk"      ┃    ┃
┃    ┃  [O] Outfitters            ┃  ┃  Hull: 100%  Shields: 80%       ┃    ┃
┃    ┃  [S] Shipyard              ┃  ┃  Fuel: 67%   Cargo: 15/50t      ┃    ┃
┃    ┃  [M] Mission BBS           ┃  ┃                                 ┃    ┃
┃    ┃  [B] Bar & News            ┃  ┃  Current System: Sol            ┃    ┃
┃    ┃  [R] Refuel (1,200 cr)     ┃  ┃  Government: United Earth       ┃    ┃
┃    ┃  [H] Repairs (Free)        ┃  ┃  Tech Level: 9                  ┃    ┃
┃    ┃                            ┃  ┃                                 ┃    ┃
┃    ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛    ┃
┃                                                                              ┃
┃    ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓    ┃
┃    ┃ NEWS: Pirate activity reported in nearby systems...             ┃    ┃
┃    ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛    ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [T]akeoff  [Tab] Next Service  [ESC] Exit                                   ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```


## 3. Commodity Trading Screen

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


## 4. Shipyard Screen

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ SHIPYARD - Earth Station                                     52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ AVAILABLE SHIPS:           ┃  ┃ SHIP: LIGHTNING FIGHTER              ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                            ┃  ┃                                      ┃  ┃
┃  ┃   Shuttle        12,000 cr ┃  ┃         ___                          ┃  ┃
┃  ┃ ▶ Lightning      45,000 cr ┃  ┃        /   \___                      ┃  ┃
┃  ┃   Courier        75,000 cr ┃  ┃       |  △  ___>                     ┃  ┃
┃  ┃   Corvette      180,000 cr ┃  ┃        \___/                         ┃  ┃
┃  ┃   Destroyer     450,000 cr ┃  ┃                                      ┃  ┃
┃  ┃   Freighter     220,000 cr ┃  ┃  Class: Light Fighter                ┃  ┃
┃  ┃   Cruiser       780,000 cr ┃  ┃  Price: 45,000 cr                    ┃  ┃
┃  ┃   Battleship  1,500,000 cr ┃  ┃                                      ┃  ┃
┃  ┃                            ┃  ┃  Hull: ████░░ 80                     ┃  ┃
┃  ┃                            ┃  ┃  Shields: ███░░░ 60                  ┃  ┃
┃  ┃                            ┃  ┃  Speed: ████████ 450                 ┃  ┃
┃  ┃                            ┃  ┃  Accel: ███████░ 380                 ┃  ┃
┃  ┃                            ┃  ┃  Maneuver: ████████ 420              ┃  ┃
┃  ┃                            ┃  ┃                                      ┃  ┃
┃  ┃                            ┃  ┃  Cargo: 15 tons                      ┃  ┃
┃  ┃                            ┃  ┃  Fuel: 300 units                     ┃  ┃
┃  ┃                            ┃  ┃  Weapon Slots: 2                     ┃  ┃
┃  ┃                            ┃  ┃  Outfit Slots: 3                     ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ YOUR SHIP: Corvette "Starhawk"                     Trade-in: 126,000 ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃ Purchase Lightning for 45,000 cr?                                   ┃  ┃
┃  ┃ With trade-in credit: You will GAIN 81,000 cr                       ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃ [ Purchase ] [ Trade-In Purchase ] [ Cancel ]                       ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [↑↓] Select Ship  [Enter] Details  [P]urchase  [T]rade-In  [ESC] Back      ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```


## 5. Outfitter Screen

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ OUTFITTER - Earth Station                                    52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ WEAPONS                        ┃  ┃ OUTFIT: SHIELD BOOSTER MK2      ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃   Laser Cannon      8,500 cr   ┃  ┃                                 ┃  ┃
┃  ┃   Pulse Laser      12,000 cr   ┃  ┃  "Advanced shield regeneration  ┃  ┃
┃  ┃   Blaze Cannon     18,000 cr   ┃  ┃   matrix increases your shield  ┃  ┃
┃  ┃   Proton Turret    45,000 cr   ┃  ┃   strength by 50 points."       ┃  ┃
┃  ┃   Missile Launcher 25,000 cr   ┃  ┃                                 ┃  ┃
┃  ┃                                ┃  ┃  Price: 35,000 cr               ┃  ┃
┃  ┃ DEFENSE                        ┃  ┃  Mass: 5 tons                   ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃                                 ┃  ┃
┃  ┃   Shield Booster     15,000 cr ┃  ┃  Effect: +50 shields            ┃  ┃
┃  ┃ ▶ Shield Boost MK2   35,000 cr ┃  ┃          +10 regen/sec          ┃  ┃
┃  ┃   Armor Plating      22,000 cr ┃  ┃                                 ┃  ┃
┃  ┃   Point Defense       8,500 cr ┃  ┃  Slots Used: 1                  ┃  ┃
┃  ┃                                ┃  ┃  Tech Level Required: 7         ┃  ┃
┃  ┃ SYSTEMS                        ┃  ┃                                 ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃  [ Purchase (Qty: 1) ]          ┃  ┃
┃  ┃   Cargo Pod         10,000 cr  ┃  ┃  [ Install & Purchase ]         ┃  ┃
┃  ┃   Fuel Tank          7,500 cr  ┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃  ┃   Engine Upgrade    50,000 cr  ┃                                       ┃
┃  ┃   Scanner Array     12,000 cr  ┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃ YOUR SHIP: Corvette            ┃  ┃
┃                                       ┃                                 ┃  ┃
┃                                       ┃ Outfit Slots: 4/6 used          ┃  ┃
┃                                       ┃ Weapon Slots: 2/4 used          ┃  ┃
┃                                       ┃                                 ┃  ┃
┃                                       ┃ Installed:                      ┃  ┃
┃                                       ┃  ▪ Laser Cannon (2x)            ┃  ┃
┃                                       ┃  ▪ Shield Booster               ┃  ┃
┃                                       ┃  ▪ Cargo Pod                    ┃  ┃
┃                                       ┃  ▪ Fuel Tank                    ┃  ┃
┃                                       ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [↑↓] Select  [Tab] Category  [B]uy  [S]ell  [I]nstall  [U]ninstall  [ESC]  ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```


## 6. Mission Board

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ MISSION BBS - Earth Station                                  52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ AVAILABLE MISSIONS                                                   ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃ ▶ [DELIVERY] Rush Shipment to Mars                                  ┃  ┃
┃  ┃   Reward: 8,500 cr   Deadline: 3 days   Cargo: 15 tons              ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃   [BOUNTY] Eliminate Pirate Lord Zaxon                              ┃  ┃
┃  ┃   Reward: 45,000 cr   Deadline: None   Difficulty: ████████         ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃   [ESCORT] Protect Convoy to Alpha Centauri                         ┃  ┃
┃  ┃   Reward: 22,000 cr   Deadline: 7 days   Ships: 3                   ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃   [DELIVERY] Medical Supplies Needed                                ┃  ┃
┃  ┃   Reward: 12,000 cr   Deadline: 2 days   Cargo: 8 tons  [URGENT]    ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃   [COMBAT] Clear Pirate Nest                                        ┃  ┃
┃  ┃   Reward: 35,000 cr   Deadline: None   Difficulty: ██████░░         ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ MISSION DETAILS: Rush Shipment to Mars                              ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  "A shipment of industrial components needs to reach Mars Colony    ┃  ┃
┃  ┃   before the next construction cycle begins. Time is of the         ┃  ┃
┃  ┃   essence! Deliver 15 tons of components within 3 days."            ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  Employer: Mars Construction Guild                                  ┃  ┃
┃  ┃  Destination: Mars - Olympus Mons Spaceport                         ┃  ┃
┃  ┃  Payment: 8,500 credits                                             ┃  ┃
┃  ┃  Cargo Space Required: 15 tons                                      ┃  ┃
┃  ┃  Time Limit: 3 days (72 hours)                                      ┃  ┃
┃  ┃  Reputation: None required                                          ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  [ Accept Mission ]  [ Decline ]                                    ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [↑↓] Select  [Enter] View Details  [A]ccept  [D]ecline  [ESC] Back         ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```


## 7. System Map / Navigation

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ NAVIGATION MAP                    [Sol System]               52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃    ╔═══════════════════════════════════════════════════════════════╗        ┃
┃    ║                    NEARBY SYSTEMS                             ║        ┃
┃    ║                                                               ║        ┃
┃    ║                                                               ║        ┃
┃    ║         Barnard's Star                   Proxima Centauri    ║        ┃
┃    ║              ◉ (8.2 ly)                      ◉ (4.5 ly)       ║        ┃
┃    ║                  ╲                            ╱               ║        ┃
┃    ║                   ╲                          ╱                ║        ┃
┃    ║                    ╲                        ╱                 ║        ┃
┃    ║                     ╲                      ╱                  ║        ┃
┃    ║                      ╲                    ╱                   ║        ┃
┃    ║                       ╲                  ╱                    ║        ┃
┃    ║                        ╲                ╱                     ║        ┃
┃    ║                         ╲              ╱                      ║        ┃
┃    ║                       ⊕══════⊙════════◉                       ║        ┃
┃    ║                      Earth  SOL  Alpha Centauri               ║        ┃
┃    ║                      (YOU)   ▲    (3.2 ly)                    ║        ┃
┃    ║                                                               ║        ┃
┃    ║              Sirius                                           ║        ┃
┃    ║               ◉ (6.8 ly)                                      ║        ┃
┃    ║                                                               ║        ┃
┃    ╚═══════════════════════════════════════════════════════════════╝        ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ JUMP DESTINATIONS:             ┃  ┃ SELECTED: Alpha Centauri        ┃  ┃
┃  ┃                                ┃  ┃                                 ┃  ┃
┃  ┃ ▶ Alpha Centauri   (3.2 ly)   ┃  ┃ Distance: 3.2 light years       ┃  ┃
┃  ┃   Proxima Centauri (4.5 ly)   ┃  ┃ Fuel Required: 32 units         ┃  ┃
┃  ┃   Sirius           (6.8 ly)   ┃  ┃ Your Fuel: 201/300 units        ┃  ┃
┃  ┃   Barnard's Star   (8.2 ly)   ┃  ┃                                 ┃  ┃
┃  ┃                                ┃  ┃ Government: Confederation       ┃  ┃
┃  ┃                                ┃  ┃ Tech Level: 8                   ┃  ┃
┃  ┃                                ┃  ┃ Population: 2.4 billion         ┃  ┃
┃  ┃                                ┃  ┃                                 ┃  ┃
┃  ┃                                ┃  ┃ Services:                       ┃  ┃
┃  ┃                                ┃  ┃ ✓ Shipyard  ✓ Outfitter         ┃  ┃
┃  ┃                                ┃  ┃ ✓ Missions  ✓ Refuel            ┃  ┃
┃  ┃                                ┃  ┃                                 ┃  ┃
┃  ┃                                ┃  ┃ [ Engage Hyperdrive ]           ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [↑↓] Select System  [Enter] Jump  [I]nfo  [ESC] Back to Space              ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

## Design Notes

### Key Differences from Current Menu-Based UI:

1. **Main Space View** - Active 2D top-down space where you fly your ship, not just menus
2. **HUD Elements** - Real-time information displayed while flying
3. **Radar/Scanner** - Shows relative positions of objects
4. **Contextual Screens** - Each activity (trading, outfitting, etc.) has dedicated screen
5. **Visual Ship Representation** - ASCII art for ships and objects
6. **Status Bars** - Visual progress bars for shields, hull, fuel
7. **Quick Commands** - Single-key commands for common actions

### Suggested Implementation Approach:

1. Start with the main space view - this is the "home" screen
2. Each planet landing brings up the landing screen with services
3. Services transition to dedicated screens (trading, shipyard, etc.)
4. Keep the escape key as "go back" throughout
5. Use arrow keys for navigation in lists
6. Single letters for quick actions

### Technical Considerations:

- Use box-drawing characters (┏━┓ etc.) for borders
- Progress bars with █░ characters
- Clear visual hierarchy with spacing
- Consistent header/footer layout
- Current system/credits always visible in header
- Command hints always in footer

Would you like me to refine any of these prototypes or create additional screens (combat, quests, etc.)?
