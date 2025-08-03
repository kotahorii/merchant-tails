# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This repository contains "Merchant Tales (マーチャントテイル ～商人物語～)", a Unity-based business simulation game where players learn investment concepts through fantasy merchant gameplay. The game teaches financial literacy without explicitly mentioning "stock trading" by having players manage a fantasy shop with 6 different item types that represent various investment concepts.

## Development Phase

**Current Status**: Documentation and planning phase - Unity project not yet created
**Target**: 18-month development cycle with planned Steam and Nintendo Switch release

## Project Architecture

The game follows a **3-layer MVC + Observer Pattern** architecture:

### Core Systems

- **GameManager**: Central state management (MainMenu, Tutorial, Shopping, StoreManagement, MarketView, Paused)
- **TimeManager**: Handles seasons (Spring/Summer/Autumn/Winter) and daily phases (Morning/Afternoon/Evening/Night)
- **MarketSystem**: Price calculations with seasonal and event-based modifiers for 6 item types
- **InventorySystem**: Dual inventory (shop front-of-house vs market speculation)
- **EventSystem**: Scheduled and random events that affect market prices
- **SaveSystem**: JSON-based persistence with auto-save functionality

### Item Types & Investment Concepts

| Item                     | Investment Concept   | Characteristics             |
| ------------------------ | -------------------- | --------------------------- |
| くだもの (Fruit)         | Short-term trading   | Expires quickly             |
| ポーション (Potion)      | Growth stocks        | Event-driven price changes  |
| 武器 (Weapon)            | Blue-chip stocks     | Stable prices, low turnover |
| アクセサリー (Accessory) | Speculative stocks   | Trend-driven volatility     |
| 魔法書 (Magic Book)      | Bonds                | High-value, stable          |
| 宝石 (Gem)               | High-risk investment | Unpredictable, high returns |

## Development Commands

### Unity Project Setup (Phase 1 - Week 1-2)

```bash
# Unity 6.1 LTS (6000.1.14f1) project creation required
# Folder structure: Scripts/, Assets/, Prefabs/, Scenes/
```

### Testing Strategy

Based on the design document, implement:

- Unit tests for MarketSystem price calculations
- Integration tests for complete day flow cycles
- Save/load integrity testing
- Performance testing for update frequency optimization

## Key Technical Decisions

### Data Management

- **Persistence**: JSON + PlayerPrefs for save data
- **State Management**: Finite State Machine pattern
- **Updates**: Separated into frame-rate, fixed-rate, and slow update cycles for performance

### Progression System

4-tier merchant ranking system that unlocks features:

- 見習い (Apprentice): ~1,000G - Basic features only
- 一人前 (Skilled): ~5,000G - Price forecasting unlocked
- ベテラン (Veteran): ~10,000G - Advanced analytics
- マスター (Master): 10,000G+ - All features

### Event-Driven Architecture

Uses EventBus pattern for system communication:

- PriceChangedEvent
- SeasonChangedEvent
- PhaseChangedEvent
- Custom game events

## Development Workflow

### Current Priority (Based on TODO List)

1. **Phase 1**: Unity environment setup and basic architecture
2. **Phase 2**: Core systems (Time, Market, Inventory, Events)
3. **Phase 3**: Gameplay features and progression
4. **Phase 4**: UI/UX systems
5. **Phase 5**: Data persistence and optimization
6. **Phase 6**: Testing and debugging

### Team Structure

- **Programmer**: Unity C# implementation, system architecture
- **Designer/Artist**: UI/UX, character art, backgrounds, audio

## Localization

Support for Japanese and English languages through LocalizationManager system.

## Important Constraints

### Educational Focus

The game must teach investment concepts without explicitly using financial terms. All mechanics should feel natural within the fantasy merchant context.

### Performance Targets

- Target platforms: Steam (PC) and Nintendo Switch
- Must handle complex market calculations while maintaining smooth gameplay
- Memory optimization required for extended play sessions

## Success Metrics

- Break-even: 1,000 copies sold
- Success target: 5,000 copies in first year
- Educational effectiveness measured through player progression and understanding

## Documentation References

- **PRD**: `docs/prd.md` - Complete product requirements
- **Design Document**: `docs/design-doc.md` - Technical architecture details
- **Development TODO**: `docs/development-todo.md` - Phase-by-phase implementation plan

## Development Best Practices

- PUSH したあとは必ず CI が成功しているかを確認すること
