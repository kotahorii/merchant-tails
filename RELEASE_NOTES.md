# Merchant Tails Release Notes

## Version 0.1.0 (Alpha) - 2025-08-12

### 🎉 Initial Alpha Release

Welcome to the first alpha release of Merchant Tails! This educational business simulation game teaches investment fundamentals through fantasy merchant trading.

### ✨ Core Features

#### Game Systems
- **Trading System**: Buy low, sell high mechanics with 6 item categories
- **Market Dynamics**: Supply and demand based pricing with realistic fluctuations
- **Banking System**: Simple savings with 2% annual interest rate
- **Weather System**: 3 weather types affecting market prices and demand
- **Season System**: 4 seasons with 30 days each, affecting item availability
- **Player Progression**: 4 merchant ranks from Apprentice to Master

#### User Interface
- **Main Menu**: New game, continue, tutorial, and settings options
- **Game Screen**: Integrated shop, market, inventory, and bank views
- **Localization**: Full support for English and Japanese languages
- **Save System**: Auto-save and manual save functionality

#### Tutorial
- **8-Step Tutorial**: Comprehensive introduction to game mechanics
- **Contextual Hints**: Help system for new players
- **Progressive Learning**: Gradual introduction of concepts

### 🎮 Gameplay

#### Item Categories
1. **Fruits** - Fast turnover, subject to spoilage
2. **Potions** - Medium risk and reward
3. **Weapons** - High value, slower sales
4. **Accessories** - Luxury items with volatile prices
5. **Spellbooks** - Specialized market segment
6. **Gems** - High risk, high reward investments

#### Difficulty Levels
- **Easy**: More forgiving prices, higher starting gold
- **Normal**: Balanced gameplay experience
- **Hard**: Challenging market conditions

### 🔧 Technical Details

#### Architecture
- **Frontend**: Godot 4.4.1 with GDScript
- **Backend**: Go 1.24 with Clean Architecture
- **Integration**: GDExtension for Go-Godot communication
- **Data**: Protocol Buffers for serialization

#### Performance
- Target 60 FPS gameplay
- Optimized market calculations
- Efficient inventory management
- Minimal memory footprint

### 📝 Known Issues

1. **GDExtension Loading**: Some systems may require manual library path configuration
2. **Save Migration**: No save compatibility between alpha versions
3. **Balancing**: Market prices may need adjustment based on player feedback
4. **Tutorial**: Some tutorial steps may not trigger correctly

### 🚫 Features NOT Included (By Design)

To maintain simplicity and focus on educational value, the following features were intentionally removed:

- AI merchant competitors
- Complex investment instruments
- Multiplayer functionality
- Advanced analytics and charts
- Loan and credit systems
- Real-time market events
- Crafting systems
- Combat mechanics

### 🛠️ Development Focus

This release prioritizes:
- **Simplicity**: Easy to understand for beginners
- **Education**: Teaching investment fundamentals
- **Stability**: Reliable core gameplay
- **Performance**: Smooth experience on modest hardware

### 📋 System Requirements

#### Minimum
- OS: Windows 10, macOS 10.15, Ubuntu 20.04
- RAM: 4GB
- Storage: 500MB
- Graphics: OpenGL 3.3 support

#### Recommended
- OS: Latest version of Windows/macOS/Linux
- RAM: 8GB
- Storage: 1GB
- Graphics: Dedicated GPU with OpenGL 4.0

### 🐛 Bug Reporting

Please report bugs through:
- GitHub Issues: [github.com/yourusername/merchant-tails/issues]
- Include: System specs, steps to reproduce, error messages

### 🙏 Acknowledgments

Special thanks to:
- Alpha testers for valuable feedback
- Open source community for tools and libraries
- Investment education advisors for gameplay consultation

### 📅 Roadmap

#### Next Release (v0.2.0)
- Balance adjustments based on player feedback
- Additional tutorial improvements
- Performance optimizations
- Bug fixes

#### Future Plans (v1.0.0)
- Achievement system
- Extended item catalog
- Seasonal events
- Platform-specific optimizations
- Steam release preparation

### 📜 License

Merchant Tails is released under [LICENSE TYPE]. See LICENSE file for details.

### 📞 Support

For support and questions:
- Documentation: `/docs` folder
- Community: Discord server (coming soon)
- Email: support@merchanttails.com (coming soon)

---

Thank you for trying Merchant Tails Alpha! Your feedback is invaluable in shaping the future of this educational gaming experience.

Happy Trading!
The Merchant Tails Team