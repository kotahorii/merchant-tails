package ai

import (
	"sync"
)

// RelationshipType represents the type of relationship between merchants
type RelationshipType int

const (
	RelationshipNeutral RelationshipType = iota
	RelationshipFriendly
	RelationshipRival
	RelationshipAllied
)

// MarketInformation represents information shared between merchants
type MarketInformation struct {
	ItemID      string
	PriceChange float64
	Source      string
	Timestamp   int64
	Reliability float64 // 0.0 to 1.0
}

// InformationPropagation represents how information spreads
type InformationPropagation struct {
	TargetID    string
	Information *MarketInformation
	Delay       int // Ticks/turns before information arrives
	Reliability float64
}

// MerchantRelationship represents a relationship between two merchants
type MerchantRelationship struct {
	MerchantA    string
	MerchantB    string
	Type         RelationshipType
	Strength     float64 // 0.0 to 1.0
	TrustLevel   float64 // 0.0 to 1.0
	TradeHistory []TradeRecord
}

// MerchantNetwork manages relationships and information flow
type MerchantNetwork struct {
	merchants     map[string]*AIMerchant
	relationships map[string]map[string]*MerchantRelationship // [merchantA][merchantB]
	information   map[string][]*MarketInformation
	mu            sync.RWMutex
}

// NewMerchantNetwork creates a new merchant network
func NewMerchantNetwork() *MerchantNetwork {
	return &MerchantNetwork{
		merchants:     make(map[string]*AIMerchant),
		relationships: make(map[string]map[string]*MerchantRelationship),
		information:   make(map[string][]*MarketInformation),
	}
}

// AddMerchant adds a merchant to the network
func (n *MerchantNetwork) AddMerchant(merchant *AIMerchant) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.merchants[merchant.ID()] = merchant

	// Initialize relationship map for this merchant
	if _, exists := n.relationships[merchant.ID()]; !exists {
		n.relationships[merchant.ID()] = make(map[string]*MerchantRelationship)
	}
}

// RemoveMerchant removes a merchant from the network
func (n *MerchantNetwork) RemoveMerchant(merchantID string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	delete(n.merchants, merchantID)
	delete(n.relationships, merchantID)

	// Remove relationships with this merchant
	for _, relations := range n.relationships {
		delete(relations, merchantID)
	}
}

// AddRelationship creates or updates a relationship between two merchants
func (n *MerchantNetwork) AddRelationship(merchantA, merchantB string, relType RelationshipType) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Ensure both merchants exist
	if _, existsA := n.merchants[merchantA]; !existsA {
		return
	}
	if _, existsB := n.merchants[merchantB]; !existsB {
		return
	}

	// Create relationship
	relationship := &MerchantRelationship{
		MerchantA:    merchantA,
		MerchantB:    merchantB,
		Type:         relType,
		Strength:     0.5, // Start with neutral strength
		TrustLevel:   0.5, // Start with neutral trust
		TradeHistory: make([]TradeRecord, 0),
	}

	// Add bidirectional relationship
	if n.relationships[merchantA] == nil {
		n.relationships[merchantA] = make(map[string]*MerchantRelationship)
	}
	if n.relationships[merchantB] == nil {
		n.relationships[merchantB] = make(map[string]*MerchantRelationship)
	}

	n.relationships[merchantA][merchantB] = relationship
	n.relationships[merchantB][merchantA] = relationship
}

// GetRelationship gets the relationship between two merchants
func (n *MerchantNetwork) GetRelationship(merchantA, merchantB string) *MerchantRelationship {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if relations, exists := n.relationships[merchantA]; exists {
		return relations[merchantB]
	}
	return nil
}

// PropagateInformation spreads market information through the network
func (n *MerchantNetwork) PropagateInformation(info *MarketInformation) []*InformationPropagation {
	n.mu.Lock()
	defer n.mu.Unlock()

	propagations := make([]*InformationPropagation, 0)

	// Get source merchant
	sourceMerchant, exists := n.merchants[info.Source]
	if !exists {
		return propagations
	}

	// Spread to connected merchants based on relationships
	if relations, exists := n.relationships[info.Source]; exists {
		for targetID, relationship := range relations {
			// Calculate propagation parameters based on relationship
			delay, reliability := n.calculatePropagationParams(relationship, sourceMerchant)

			propagation := &InformationPropagation{
				TargetID:    targetID,
				Information: info,
				Delay:       delay,
				Reliability: reliability * info.Reliability, // Compound reliability
			}

			propagations = append(propagations, propagation)

			// Store information for the target
			if n.information[targetID] == nil {
				n.information[targetID] = make([]*MarketInformation, 0)
			}
			n.information[targetID] = append(n.information[targetID], info)
		}
	}

	return propagations
}

// calculatePropagationParams determines how quickly and reliably information spreads
func (n *MerchantNetwork) calculatePropagationParams(relationship *MerchantRelationship, _ *AIMerchant) (int, float64) {
	baseDelay := 3 // Base delay in ticks
	baseReliability := 0.7

	switch relationship.Type {
	case RelationshipAllied:
		// Allies share information quickly and reliably
		return 1, 0.95 * relationship.TrustLevel

	case RelationshipFriendly:
		// Friends share information reasonably quickly
		delay := baseDelay - int(relationship.Strength*2)
		if delay < 1 {
			delay = 1
		}
		return delay, baseReliability + 0.2*relationship.TrustLevel

	case RelationshipRival:
		// Rivals share information slowly and unreliably
		return baseDelay * 2, 0.3 * relationship.TrustLevel

	default: // Neutral
		return baseDelay, baseReliability * relationship.TrustLevel
	}
}

// GetMerchantInformation gets all information available to a merchant
func (n *MerchantNetwork) GetMerchantInformation(merchantID string) []*MarketInformation {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.information[merchantID]
}

// UpdateRelationshipStrength updates the strength of a relationship based on interactions
func (n *MerchantNetwork) UpdateRelationshipStrength(merchantA, merchantB string, delta float64) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Get relationship without calling GetRelationship to avoid deadlock
	var relationship *MerchantRelationship
	if relations, exists := n.relationships[merchantA]; exists {
		relationship = relations[merchantB]
	}

	if relationship != nil {
		relationship.Strength += delta

		// Clamp between 0 and 1
		if relationship.Strength > 1.0 {
			relationship.Strength = 1.0
		} else if relationship.Strength < 0.0 {
			relationship.Strength = 0.0
		}

		// Adjust trust based on strength changes
		if delta > 0 {
			relationship.TrustLevel += delta * 0.5
		} else {
			relationship.TrustLevel += delta * 0.3 // Trust decreases slower
		}

		// Clamp trust
		if relationship.TrustLevel > 1.0 {
			relationship.TrustLevel = 1.0
		} else if relationship.TrustLevel < 0.0 {
			relationship.TrustLevel = 0.0
		}
	}
}

// GetNetworkClusters identifies merchant clusters/alliances
func (n *MerchantNetwork) GetNetworkClusters() [][]string {
	n.mu.RLock()
	defer n.mu.RUnlock()

	visited := make(map[string]bool)
	clusters := make([][]string, 0)

	for merchantID := range n.merchants {
		if !visited[merchantID] {
			cluster := n.findCluster(merchantID, visited)
			if len(cluster) > 1 { // Only include actual clusters
				clusters = append(clusters, cluster)
			}
		}
	}

	return clusters
}

// findCluster performs DFS to find connected merchants
func (n *MerchantNetwork) findCluster(startID string, visited map[string]bool) []string {
	cluster := make([]string, 0)
	stack := []string{startID}

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if visited[current] {
			continue
		}

		visited[current] = true
		cluster = append(cluster, current)

		// Add connected merchants with strong relationships
		if relations, exists := n.relationships[current]; exists {
			for neighborID, relationship := range relations {
				if !visited[neighborID] &&
					(relationship.Type == RelationshipAllied ||
						(relationship.Type == RelationshipFriendly && relationship.Strength > 0.7)) {
					stack = append(stack, neighborID)
				}
			}
		}
	}

	return cluster
}
