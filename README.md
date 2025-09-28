# TCG Engine

A Digital Trading Card Game engine written in Go - a personal learning project focused on systems design, state management, and clean architecture principles.

## 🎯 Project Overview

This is a backend-focused TCG engine that implements core card game mechanics similar to games like Magic: The Gathering or Hearthstone. The project serves as a learning playground for exploring complex state management, game rule systems, and architectural patterns in Go.

## 🚀 Features Implemented

### Core Game Systems
- **Multi-effect card resolution** - Cards can have multiple sequential effects
- **Flexible targeting system** - Support for player, creature, and auto-targeting
- **Zone management** - Cards move between deck, hand, board, and graveyard
- **Owner/Controller tracking** - Proper handling of card ownership vs control
- **Turn-based gameplay** - Energy/mana system with automatic ramping
- **Effect resolution** - Damage, healing, card draw, and stat buffs

### Architecture Highlights
- **Clean separation of concerns** - Distinct packages for game logic, cards, players
- **Comprehensive validation** - Multi-layer validation for game actions
- **Extensive test coverage** - TDD approach with deterministic testing
- **Extensible effect system** - Function map pattern for easy effect additions

## 🛠️ Tech Stack

- **Language:** Go 1.25.1
- **Testing:** testify framework
- **Web Framework:** chi router (prepared for future API endpoints)

## 📚 Learning Goals

This project is designed to practice:
- Complex state machine design
- Rule-based system modeling
- Test-driven development with Go
- Clean architecture principles
- System design for extensibility
- Managing intricate game state interactions

## 🎮 Game Concepts

The engine supports:
- **Creatures:** Permanent cards with attack/health that can battle
- **Spells:** One-time effects that resolve and go to graveyard
- **Effects:** Modular actions like damage, heal, draw cards, buff stats
- **Targeting:** Flexible targeting system with validation
- **Resources:** Energy system that increases each turn

## 🧪 Testing

Run tests with:
```bash
go test ./... -v
```

Tests include deterministic seed-based testing to ensure reproducible game states.

## 📝 Project Status

This is an active learning project. Current focus areas:
- [x] Core game state management
- [x] Multi-effect spell resolution
- [x] Zone movement and card lifecycle
- [x] Effect targeting and validation
- [ ] Combat system
- [ ] REST API implementation
- [ ] Game persistence
- [ ] Simple AI opponent

## 🤝 Note

This is a personal learning project focused on backend systems and game logic. There is intentionally no UI/frontend component - all interaction is designed to happen via API endpoints (future) and tests (current).
