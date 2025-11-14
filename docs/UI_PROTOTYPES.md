# Terminal Velocity - UI Prototypes (Escape Velocity Style)

Based on the classic Escape Velocity game interface, adapted for terminal/ASCII.

## 0. Login / Registration Screen

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃                                                                              ┃
┃                                                                              ┃
┃                         ████████╗███████╗██████╗ ███╗   ███╗               ┃
┃                         ╚══██╔══╝██╔════╝██╔══██╗████╗ ████║               ┃
┃                            ██║   █████╗  ██████╔╝██╔████╔██║               ┃
┃                            ██║   ██╔══╝  ██╔══██╗██║╚██╔╝██║               ┃
┃                            ██║   ███████╗██║  ██║██║ ╚═╝ ██║               ┃
┃                            ╚═╝   ╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝               ┃
┃                                                                              ┃
┃                        ██╗   ██╗███████╗██╗      ██████╗  ██████╗██╗████████╗██╗   ██╗
┃                        ██║   ██║██╔════╝██║     ██╔═══██╗██╔════╝██║╚══██╔══╝╚██╗ ██╔╝
┃                        ██║   ██║█████╗  ██║     ██║   ██║██║     ██║   ██║    ╚████╔╝
┃                        ╚██╗ ██╔╝██╔══╝  ██║     ██║   ██║██║     ██║   ██║     ╚██╔╝
┃                         ╚████╔╝ ███████╗███████╗╚██████╔╝╚██████╗██║   ██║      ██║
┃                          ╚═══╝  ╚══════╝╚══════╝ ╚═════╝  ╚═════╝╚═╝   ╚═╝      ╚═╝
┃                                                                              ┃
┃                           A Multiplayer Space Trading Game                  ┃
┃                                                                              ┃
┃                                                                              ┃
┃                  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓              ┃
┃                  ┃           LOGIN TO YOUR ACCOUNT           ┃              ┃
┃                  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫              ┃
┃                  ┃                                           ┃              ┃
┃                  ┃  Username: [___________________________]  ┃              ┃
┃                  ┃                                           ┃              ┃
┃                  ┃  Password: [***************************]  ┃              ┃
┃                  ┃                                           ┃              ┃
┃                  ┃           [ Login with Password ]         ┃              ┃
┃                  ┃           [ Login with SSH Key  ]         ┃              ┃
┃                  ┃                                           ┃              ┃
┃                  ┃  ─────────────── OR ───────────────────   ┃              ┃
┃                  ┃                                           ┃              ┃
┃                  ┃       [ Create New Account ]              ┃              ┃
┃                  ┃                                           ┃              ┃
┃                  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛              ┃
┃                                                                              ┃
┃                                                                              ┃
┃              Connect via SSH: ssh username@terminal-velocity.io:2222        ┃
┃                                                                              ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [Tab] Next Field  [Enter] Submit  [R]egister  [Q]uit                        ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```


## 1. Main Space View (In-Flight) - With Chat

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
┃ ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓ ┃
┃ ┃ CHAT [Global] ▼                                               [C] to expand ┃ ┃
┃ ┃ SpaceCadet: Anyone near Sol system?                                  3m ago ┃ ┃
┃ ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛ ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [L]and  [J]ump  [T]arget  [F]ire  [H]ail  [M]ap  [C]hat  [I]nfo  [ESC] Menu ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

**Main Space View - Chat Expanded:**

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
┃    ║        *                         △                            ║  ┃   ⊕    ◆    ┃
┃    ║                                 You                           ║  ┃        ▲    ┃
┃    ║                                             ◆ Pirate          ║  ┗━━━━━━━━━━━━━┛
┃    ║           ⊕ Mars                                              ║  ┏━━━━━━━━━━━━━┓
┃    ╚═══════════════════════════════════════════════════════════════╝  ┃   STATUS    ┃
┃                                                                        ┃ Hull: ██████┃
┃ ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃ Fuel: ████░░┃
┃ ┃ CHAT: [Global ▼] [System] [Faction] [DM]         [C] to collapse ┃  ┃ Speed: 340  ┃
┃ ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃ Credits:    ┃
┃ ┃                                                                  ┃  ┃  52,400 cr  ┃
┃ ┃ [SpaceCadet] Anyone near Sol system?                     3m ago ┃  ┗━━━━━━━━━━━━━┛
┃ ┃ [TraderJoe] Yeah I'm docked at Earth. Need anything?     2m ago ┃
┃ ┃ [SpaceCadet] Looking for escort to Alpha Centauri        2m ago ┃
┃ ┃ [PirateKing] I'll escort you... to your doom! Arr!       1m ago ┃
┃ ┃ [TraderJoe] Ignore him. I can escort for 5k credits      1m ago ┃
┃ ┃ [YOU] I'm at Earth too, what's the pirate situation?     now    ┃
┃ ┃                                                                  ┃
┃ ┃ Message: [_________________________________________________]    ┃
┃ ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [Tab] Switch Channel  [Enter] Send  [C]ollapse Chat  [ESC] Back to Flight   ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
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


## 8. Active Combat Screen

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ COMBAT ENGAGED!                  [Sol System]          Shields: ██████░░░░ 60%┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃    ╔═══════════════════════════════════════════════════════════════╗        ┃
┃    ║                    TACTICAL DISPLAY                           ║        ┃
┃    ║                                                               ║        ┃
┃    ║                                                               ║        ┃
┃    ║                          ◆                                    ║  ┏━━━━━━━━━━━━━┓
┃    ║                      Pirate Viper                             ║  ┃ YOUR SHIP   ┃
┃    ║                       [LOCKED]                                ║  ┣━━━━━━━━━━━━━┫
┃    ║                          ↓                                    ║  ┃ Corvette    ┃
┃    ║                   ~~~~ WEAPONS ~~~~                           ║  ┃             ┃
┃    ║                          ↓                                    ║  ┃ Hull: ██████┃
┃    ║                                                               ║  ┃       100%  ┃
┃    ║                          △                                    ║  ┃             ┃
┃    ║                      Your Ship                                ║  ┃ Shields:    ┃
┃    ║                                                               ║  ┃ ██████░░░░  ┃
┃    ║                                                               ║  ┃       60%   ┃
┃    ║        Distance: 1,850 km     Closing at 120 km/s            ║  ┃             ┃
┃    ║                                                               ║  ┃ Energy:     ┃
┃    ╚═══════════════════════════════════════════════════════════════╝  ┃ ████████░░  ┃
┃                                                                        ┃       80%   ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┗━━━━━━━━━━━━━┛
┃  ┃ ENEMY: Pirate Viper                                           ┃
┃  ┃ Hull: ████░░░░ 40%   Shields: ██░░░░░░ 25%   Weapons: Active ┃  ┏━━━━━━━━━━━━━┓
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃ WEAPONS     ┃
┃                                                                        ┣━━━━━━━━━━━━━┫
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃ 1. Laser    ┃
┃  ┃ COMBAT LOG:                                                    ┃  ┃    Cannon   ┃
┃  ┃ > Pirate Viper is hailing you: "Prepare to die!"              ┃  ┃    [READY]  ┃
┃  ┃ > You fire Laser Cannon - HIT for 45 damage!                  ┃  ┃             ┃
┃  ┃ > Pirate fires Pulse Laser - MISS!                            ┃  ┃ 2. Pulse    ┃
┃  ┃ > Your shields absorb 30 damage from Pulse Laser              ┃  ┃    Laser    ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃    [READY]  ┃
┃                                                                        ┃             ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃ 3. Missiles ┃
┃  ┃ YOUR TURN - Select Action:                                     ┃  ┃    [15/15]  ┃
┃  ┃                                                                ┃  ┗━━━━━━━━━━━━━┛
┃  ┃  [1] Fire Laser Cannon     [2] Fire Pulse Laser               ┃
┃  ┃  [3] Fire Missile          [E] Evasive Maneuvers              ┃
┃  ┃  [D] Defend (Boost Shields) [R] Retreat (Flee Combat)         ┃
┃  ┃                                                                ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [1-3] Fire Weapon  [E]vade  [D]efend  [R]etreat  [H]ail  [ESC] Menu        ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```


## 9. Quest Board

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ QUEST TERMINAL - Earth Station                              52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ ACTIVE QUESTS                                          [2/5 Active]   ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃ ▶ [MAIN] The Pirate Menace                               Chapter 1  ┃  ┃
┃  ┃   Progress: ████░░░░░░ 40%   Next: Investigate Sirius system       ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃   [SIDE] Trader's Gambit                                            ┃  ┃
┃  ┃   Progress: ██████░░░░ 60%   Next: Deliver goods to Mars           ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ QUEST: The Pirate Menace (Main Quest - Chapter 1)                   ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  A mysterious increase in pirate activity has been reported across  ┃  ┃
┃  ┃  human space. United Earth Intelligence suspects a larger           ┃  ┃
┃  ┃  organization is coordinating these attacks. Your mission is to     ┃  ┃
┃  ┃  investigate and neutralize the threat.                             ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  OBJECTIVES:                                                         ┃  ┃
┃  ┃  ✓ Speak with Admiral Chen at Earth Station                         ┃  ┃
┃  ┃  ✓ Eliminate 5 pirate ships in Sol system                           ┃  ┃
┃  ┃  ▪ Investigate pirate base in Sirius system           [IN PROGRESS] ┃  ┃
┃  ┃  ▪ Recover pirate communications logs                               ┃  ┃
┃  ┃  ▪ Return to Admiral Chen with findings                             ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  REWARDS:                                                            ┃  ┃
┃  ┃  • 50,000 credits                                                    ┃  ┃
┃  ┃  • Reputation: +20 United Earth                                     ┃  ┃
┃  ┃  • Unlock: Advanced Weapons Access                                  ┃  ┃
┃  ┃  • Unlock: Chapter 2 - "The Shadow Syndicate"                       ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ AVAILABLE QUESTS (3)                                                 ┃  ┃
┃  ┃ [SIDE] Lost Cargo                [FACTION] Pirate Hunters United    ┃  ┃
┃  ┃ [EXPLORATION] The Outer Reaches                                     ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [↑↓] Select Quest  [Enter] Details  [A]bandon Quest  [ESC] Back            ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```


## 10. Player Info / Status Screen

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ PILOT RECORD                                                52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ PILOT: SpaceCaptain              ┃  ┃ CURRENT SHIP: Corvette         ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃ "Starhawk"                     ┃  ┃
┃  ┃                                  ┃  ┃                                ┃  ┃
┃  ┃ Status: Competent Trader         ┃  ┃         ___                    ┃  ┃
┃  ┃                                  ┃  ┃        /   \___                ┃  ┃
┃  ┃ Combat Rating: ████████░░ 80/100 ┃  ┃       |  ▲  ====>              ┃  ┃
┃  ┃ Trading Rating: ██████░░░░ 65/100┃  ┃        \___/                   ┃  ┃
┃  ┃ Exploration: ████░░░░░░ 45/100   ┃  ┃                                ┃  ┃
┃  ┃                                  ┃  ┃ Hull: 100%  Shields: 80%       ┃  ┃
┃  ┃ Rank: Commander                  ┃  ┃ Cargo: 15/50 tons              ┃  ┃
┃  ┃ Account Age: 47 days             ┃  ┃ Fuel: 201/300 units            ┃  ┃
┃  ┃ Play Time: 142 hours             ┃  ┃                                ┃  ┃
┃  ┃                                  ┃  ┃ Value: 180,000 cr              ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ STATISTICS                                                           ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  Total Kills: 143                 Missions Completed: 47            ┃  ┃
┃  ┃  Pirate Kills: 89                 Quests Completed: 8               ┃  ┃
┃  ┃  Player Kills: 3                  Failed Missions: 2                ┃  ┃
┃  ┃  Deaths: 12                       Active Quests: 2                  ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  Systems Explored: 34/100         Jumps Made: 1,247                 ┃  ┃
┃  ┃  Planets Visited: 67              Total Distance: 847 LY            ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  Total Credits Earned: 2,340,500  Profit from Trading: 890,200     ┃  ┃
┃  ┃  Total Spent: 2,288,100           Biggest Trade: 45,000             ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ REPUTATION                                                           ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  United Earth:    ████████░░ +75  (Respected)                       ┃  ┃
┃  ┃  Confederation:   █████░░░░░ +45  (Friendly)                        ┃  ┃
┃  ┃  Free Traders:    ███████░░░ +68  (Honored)                         ┃  ┃
┃  ┃  Rebels:          ██░░░░░░░░ +15  (Neutral)                         ┃  ┃
┃  ┃  Pirates:         ░░░░░░░░░░ -85  (Hostile)                         ┃  ┃
┃  ┃  Corporation:     ████░░░░░░ +32  (Friendly)                        ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [A]chievements  [L]eaderboards  [S]hips Owned  [ESC] Back                  ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```


## 11. Faction Management Screen

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ FACTION MANAGEMENT                                          52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ YOUR FACTION: Star Traders Guild                           [Leader]  ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  "We trade in the stars, and the stars trade with us."              ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  Founded: 23 days ago                Members: 47                    ┃  ┃
┃  ┃  Faction Level: 12                   Active Members: 18             ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  Treasury: 2,450,000 credits         Territory: 3 systems           ┃  ┃
┃  ┃  Passive Income: +12,000 cr/day      Tax Rate: 5%                   ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ MEMBERS (Online: 18/47)          ┃  ┃ FACTION STATS                  ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃  ┃                                  ┃  ┃                                ┃  ┃
┃  ┃ ▶ SpaceCaptain    [Leader] ●    ┃  ┃ Total Kills: 2,340             ┃  ┃
┃  ┃   TraderJoe       [Officer] ●   ┃  ┃ Trade Volume: 45M cr           ┃  ┃
┃  ┃   CargoPilot      [Member] ●    ┃  ┃ Systems Explored: 89           ┃  ┃
┃  ┃   SpaceMerc       [Member] ●    ┃  ┃ Faction Wars Won: 12           ┃  ┃
┃  ┃   MiningBoss      [Member] ○    ┃  ┃                                ┃  ┃
┃  ┃   QuickShip       [Recruit] ●   ┃  ┃ Alliances: 2                   ┃  ┃
┃  ┃   PirateBane      [Member] ●    ┃  ┃ Enemies: 1                     ┃  ┃
┃  ┃   ... (40 more)                 ┃  ┃                                ┃  ┃
┃  ┃                                  ┃  ┃ Faction Rank: #7 (Global)      ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ CONTROLLED TERRITORY                                                 ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  ⊙ Sirius System         Income: +5,000 cr/day    Defense: ████░░   ┃  ┃
┃  ┃  ⊙ Barnard's Star        Income: +4,200 cr/day    Defense: ██████   ┃  ┃
┃  ┃  ⊙ Wolf 359              Income: +2,800 cr/day    Defense: ███░░░   ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ ACTIONS:                                                             ┃  ┃
┃  ┃ [I]nvite Member  [K]ick Member  [P]romote  [D]emote  [T]erritory    ┃  ┃
┃  ┃ [W]ar Declaration  [A]lliance  [S]ettings  [L]eave Faction          ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [↑↓] Select Member  [Tab] Switch Panel  [Enter] Details  [ESC] Back        ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```


## 12. Leaderboards Screen

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ LEADERBOARDS                                                52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ [Credits ▼]  [Combat]  [Trading]  [Exploration]          Season 3    ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ TOP PILOTS BY NET WORTH                                              ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃ RANK  PILOT            FACTION              NET WORTH      SHIP      ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  🥇 1  TradeKing        Merchant Guild      45,890,250 cr  Cruiser  ┃  ┃
┃  ┃  🥈 2  PirateLord       Crimson Raiders     42,156,800 cr  Battlesh ┃  ┃
┃  ┃  🥉 3  SpaceBaron       Free Traders        38,920,100 cr  Cruiser  ┃  ┃
┃  ┃     4  CreditCollector  Star Traders Guild  35,445,670 cr  Corvette ┃  ┃
┃  ┃     5  WealthSeeker     Independent         32,108,900 cr  Freighter┃  ┃
┃  ┃     6  MoneyMaker       Merchant Guild      28,567,300 cr  Corvette ┃  ┃
┃  ┃     7  RichTrader       Star Traders Guild  25,890,400 cr  Corvette ┃  ┃
┃  ┃     8  CargoMogul       Free Traders        23,445,200 cr  Freighter┃  ┃
┃  ┃     9  ProfitHunter     Independent         21,234,500 cr  Courier  ┃  ┃
┃  ┃    10  TradeMaster      Merchant Guild      19,876,300 cr  Corvette ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃    ...                                                               ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃   ▶ 47  SpaceCaptain    Star Traders Guild   8,234,500 cr  Corvette ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃    ...                                                               ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃   247  LastPlace        Independent            12,450 cr  Shuttle   ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ YOUR RANK: #47 of 247 active pilots                                 ┃  ┃
┃  ┃ Next Rank: Gain 720,000 cr to reach #46                             ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [Tab] Switch Category  [F]ilter by Faction  [S]eason History  [ESC] Back  ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```


## 13. Settings Screen

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃ SETTINGS                                                    52,400 credits ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ [Display ▼]  [Gameplay]  [Audio]  [Controls]  [Privacy]  [Account]  ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃ DISPLAY SETTINGS                                                     ┃  ┃
┃  ┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  Color Scheme:                                                       ┃  ┃
┃  ┃  ◉ Classic Green       ○ Blue Plasma      ○ Amber Terminal          ┃  ┃
┃  ┃  ○ White on Black      ○ Cyberpunk Neon                             ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  ─────────────────────────────────────────────────────────────────   ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  UI Scale:  [────●──────] 100%                                       ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  Animation Speed:  [───────●───] Normal                             ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  ─────────────────────────────────────────────────────────────────   ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  ☑ Show FPS Counter                                                 ┃  ┃
┃  ┃  ☑ Show Combat Damage Numbers                                       ┃  ┃
┃  ┃  ☑ Animate Space Objects                                            ┃  ┃
┃  ┃  ☐ Reduce Visual Effects (Performance Mode)                         ┃  ┃
┃  ┃  ☑ Show Player Names in Space                                       ┃  ┃
┃  ┃  ☐ Show Chat Timestamps                                             ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  ─────────────────────────────────────────────────────────────────   ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  HUD Opacity:  [──────────●] 100%                                    ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┃  Chat Window Position:                                               ┃  ┃
┃  ┃  ◉ Bottom       ○ Top         ○ Floating                            ┃  ┃
┃  ┃                                                                      ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┃  ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓  ┃
┃  ┃                   [ Save Settings ]  [ Reset to Default ]            ┃  ┃
┃  ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛  ┃
┃                                                                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ [Tab] Switch Category  [Space] Toggle Option  [S]ave  [R]eset  [ESC] Back  ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```
