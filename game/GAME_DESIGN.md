# Game Design Document: Isometric Social Room Game

## Project Overview
An isometric room decoration and social game built with Ebitengine in Go, inspired by Habbo Hotel. Players can decorate rooms, interact with furniture, customize their avatars, and socialize with other players in real-time.

**Technical Specifications:**
- Engine: Ebitengine (Go)
- Art Style: 64x64 pixel art, isometric perspective
- Platform: Desktop and Web (WebAssembly)
- Target Resolution: 640x480 (expandable)

## Core Game Pillars
1. **Social Interaction** - Real-time player interaction and communication
2. **Creative Expression** - Room decoration and avatar customization
3. **Functional Gameplay** - Interactive furniture and activity systems
4. **Community Building** - Visiting rooms, sharing creations, collaborative activities

## Game Systems

### 1. Room System
**Status:** Not yet implemented
**Priority:** Core

#### Components:
- Isometric grid-based room layout
- Room ownership and permissions
- Public/private room settings
- Room visiting mechanics
- Room templates/themes

#### Technical Considerations:
- Grid size (suggested: 16x16 tiles minimum)
- Tile dimensions in isometric view
- Z-ordering for proper depth rendering
- Room data serialization for saving/loading

### 2. Furniture System
**Status:** Conceptualized
**Priority:** Core

#### Functional Furniture Categories:
- **Cooking Appliances** (stoves, ovens, microwaves)
  - Ingredient consumption from inventory
  - Cooking animations (quick, 2-3 second sequences)
  - Recipe system for different meals
  - Meal quality/success rates
  
- **Seating** (chairs, sofas, benches)
  - Avatar sitting animations
  - Social interaction bonuses when seated together
  
- **Storage** (cabinets, chests, refrigerators)
  - Inventory expansion
  - Ingredient preservation (for cooking system)
  
- **Entertainment** (TVs, radios, game consoles)
  - Mini-games or passive entertainment
  - Group watching/listening experiences

#### Technical Requirements:
- Furniture placement validation
- Rotation states (4 or 8 directions)
- Interaction zones/hotspots
- State management (on/off, in-use, etc.)

### 3. Character Customization System
**Status:** Conceptualized
**Priority:** Core

#### Customizable Elements:
- **Head**
  - Hair styles (10-15 initial options)
  - Hair colors (color picker or preset palette)
  - Hats/headwear (separate layer)
  
- **Face**
  - Eye styles and colors
  - Facial features (nose, mouth variations)
  - Accessories (glasses, facial hair)
  
- **Body**
  - Skin tones (inclusive range)
  - Body types (if feasible with art constraints)
  
- **Clothing**
  - Shirts/tops (various styles)
  - Pants/bottoms/skirts
  - Shoes (multiple styles)
  - Outerwear (jackets, coats)
  
#### Implementation Notes:
- Modular sprite system with layers
- Z-order: back hair → body → clothes → front hair → accessories
- Animation sets for each customizable piece
- Wardrobe save/load system

### 4. Inventory & Item System
**Status:** Required for other systems
**Priority:** High

#### Components:
- Personal inventory (grid or list-based)
- Item categories (furniture, ingredients, clothes, consumables)
- Item stacking for consumables
- Trading system between players
- Currency system (for purchasing items)

#### Integration Points:
- Cooking system (ingredients)
- Furniture placement (owned furniture)
- Wardrobe system (owned clothing)
- Gift giving/trading

### 5. Social Interaction System
**Status:** Planning needed
**Priority:** Core

#### Features:
- **Chat System**
  - Text chat with chat bubbles
  - Emotes/reactions
  - Private messaging
  
- **Player Actions**
  - Wave, dance, sit animations
  - Gift giving (hand items to other players)
  - Friend system
  
- **Meal Sharing** (from cooking system)
  - Hand cooked meals to other players
  - Eating animations
  - Buffs/effects from meals

### 6. Animation System
**Status:** Foundation needed
**Priority:** High

#### Required Animations:
- **Character Movement**
  - 8-directional walk cycles
  - Idle animations
  - Sitting/standing transitions
  
- **Interaction Animations**
  - Item pickup/placement
  - Cooking sequences
  - Eating/drinking
  - Social gestures
  
- **Furniture Animations**
  - Stove flames/cooking effects
  - TV screens
  - Opening/closing doors and drawers

## Development Phases

### Phase 1: Foundation (Current)
- [x] Basic game loop with Ebitengine
- [ ] Isometric rendering system
- [ ] Basic player movement in isometric space
- [ ] Tile-based room rendering

### Phase 2: Core Room Mechanics
- [ ] Furniture placement system
- [ ] Basic furniture sprites (5-10 items)
- [ ] Room saving/loading
- [ ] Camera controls (zoom, pan)

### Phase 3: Character System
- [ ] Character customization UI
- [ ] Modular sprite system
- [ ] Basic animations (walk, idle)
- [ ] Character save/load

### Phase 4: Interaction Systems
- [ ] Furniture interactions
- [ ] Cooking system implementation
- [ ] Inventory management
- [ ] Item transfer between players

### Phase 5: Social Features
- [ ] Multiplayer networking
- [ ] Chat system
- [ ] Room visiting
- [ ] Friend system

### Phase 6: Polish & Expansion
- [ ] Additional furniture sets
- [ ] More cooking recipes
- [ ] Mini-games
- [ ] Achievements/progression

## System Interconnections

```
Character Customization <---> Inventory (clothing items)
         |
         v
    Animation System <---> Social Interactions
         |
         v
    Room System <---> Furniture System
         |                    |
         v                    v
    Multiplayer <---> Cooking System <---> Inventory (ingredients)
                              |
                              v
                      Item Transfer/Trading
```

## Technical Challenges & Considerations

1. **Isometric Depth Sorting**
   - Complex z-ordering for overlapping sprites
   - Performance optimization for many objects

2. **Modular Sprite System**
   - Efficient layering of character parts
   - Animation synchronization across layers

3. **Networking**
   - Real-time position updates
   - State synchronization for furniture/items
   - Chat message delivery

4. **Save System**
   - Room configurations
   - Character customization
   - Inventory persistence

5. **Performance**
   - Sprite batching for many furniture items
   - Efficient collision detection in isometric space
   - WebAssembly optimization

## Art Asset Requirements

### Priority 1 (MVP)
- [ ] Character base sprites (8 directions, walk/idle)
- [ ] 5-10 basic furniture items
- [ ] Isometric floor/wall tiles
- [ ] Basic UI elements

### Priority 2
- [ ] Character customization pieces (hair, clothes)
- [ ] Functional furniture sprites with states
- [ ] Food/ingredient sprites
- [ ] Cooking animations/effects

### Priority 3
- [ ] Decorative furniture variations
- [ ] Seasonal/themed content
- [ ] Special effects (particles, etc.)

## Open Questions & Decisions Needed

1. **Room Size**: Fixed size or expandable? Start with 16x16 and test?
2. **Art Style**: Exact pixel dimensions for characters? (32x64 suggested for isometric)
3. **Monetization**: Premium furniture? Cosmetics? Room slots?
4. **Recipes**: How complex should the cooking system be? Timing-based or ingredient-only?
5. **Persistence**: Local saves only or cloud/server-based?
6. **Social Features**: Public room directory? Featured rooms?

## Next Steps

1. Implement isometric rendering system
2. Create basic character sprite with movement
3. Design tile-based room system
4. Prototype furniture placement mechanics
5. Test cooking system concept with simple implementation

---

*This document is a living reference that will be updated as development progresses and new ideas are added.*