package ai

// PersonalityType represents different AI merchant personalities
type PersonalityType int

const (
	PersonalityAggressive PersonalityType = iota
	PersonalityConservative
	PersonalityBalanced
	PersonalityOpportunistic
)

// MerchantPersonality defines the behavior traits of an AI merchant
type MerchantPersonality interface {
	Type() PersonalityType
	RiskTolerance() float64         // 0.0 (risk-averse) to 1.0 (risk-seeking)
	TradingFrequency() float64      // Multiplier for how often they trade
	ProfitMarginTarget() float64    // Target profit margin
	CompetitivenessFactor() float64 // How much they react to competition
	PatienceFactor() float64        // How long they wait for good deals
}

// AggressivePersonality represents an aggressive trading style
type AggressivePersonality struct{}

// NewAggressivePersonality creates an aggressive personality
func NewAggressivePersonality() MerchantPersonality {
	return &AggressivePersonality{}
}

func (p *AggressivePersonality) Type() PersonalityType {
	return PersonalityAggressive
}

func (p *AggressivePersonality) RiskTolerance() float64 {
	return 0.8
}

func (p *AggressivePersonality) TradingFrequency() float64 {
	return 1.5
}

func (p *AggressivePersonality) ProfitMarginTarget() float64 {
	return 0.3
}

func (p *AggressivePersonality) CompetitivenessFactor() float64 {
	return 1.2
}

func (p *AggressivePersonality) PatienceFactor() float64 {
	return 0.5
}

// ConservativePersonality represents a conservative trading style
type ConservativePersonality struct{}

// NewConservativePersonality creates a conservative personality
func NewConservativePersonality() MerchantPersonality {
	return &ConservativePersonality{}
}

func (p *ConservativePersonality) Type() PersonalityType {
	return PersonalityConservative
}

func (p *ConservativePersonality) RiskTolerance() float64 {
	return 0.2
}

func (p *ConservativePersonality) TradingFrequency() float64 {
	return 0.7
}

func (p *ConservativePersonality) ProfitMarginTarget() float64 {
	return 0.5
}

func (p *ConservativePersonality) CompetitivenessFactor() float64 {
	return 0.8
}

func (p *ConservativePersonality) PatienceFactor() float64 {
	return 1.5
}

// BalancedPersonality represents a balanced trading style
type BalancedPersonality struct{}

// NewBalancedPersonality creates a balanced personality
func NewBalancedPersonality() MerchantPersonality {
	return &BalancedPersonality{}
}

func (p *BalancedPersonality) Type() PersonalityType {
	return PersonalityBalanced
}

func (p *BalancedPersonality) RiskTolerance() float64 {
	return 0.5
}

func (p *BalancedPersonality) TradingFrequency() float64 {
	return 1.0
}

func (p *BalancedPersonality) ProfitMarginTarget() float64 {
	return 0.4
}

func (p *BalancedPersonality) CompetitivenessFactor() float64 {
	return 1.0
}

func (p *BalancedPersonality) PatienceFactor() float64 {
	return 1.0
}

// OpportunisticPersonality represents an opportunistic trading style
type OpportunisticPersonality struct{}

// NewOpportunisticPersonality creates an opportunistic personality
func NewOpportunisticPersonality() MerchantPersonality {
	return &OpportunisticPersonality{}
}

func (p *OpportunisticPersonality) Type() PersonalityType {
	return PersonalityOpportunistic
}

func (p *OpportunisticPersonality) RiskTolerance() float64 {
	return 0.6
}

func (p *OpportunisticPersonality) TradingFrequency() float64 {
	return 1.3
}

func (p *OpportunisticPersonality) ProfitMarginTarget() float64 {
	return 0.35
}

func (p *OpportunisticPersonality) CompetitivenessFactor() float64 {
	return 1.1
}

func (p *OpportunisticPersonality) PatienceFactor() float64 {
	return 0.8
}

// GetPersonalityName returns the string name of a personality type
func GetPersonalityName(personality PersonalityType) string {
	switch personality {
	case PersonalityAggressive:
		return "Aggressive"
	case PersonalityConservative:
		return "Conservative"
	case PersonalityBalanced:
		return "Balanced"
	case PersonalityOpportunistic:
		return "Opportunistic"
	default:
		return "Unknown"
	}
}
