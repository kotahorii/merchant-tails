package ecs

// Standard component types used in the game
const (
	ComponentTypeTransform ComponentType = iota
	ComponentTypeRender
	ComponentTypePhysics
	ComponentTypeInventory
	ComponentTypeMerchant
	ComponentTypeShop
	ComponentTypePrice
	ComponentTypeAI
	ComponentTypePlayer
	ComponentTypeItem
)

// TransformComponent represents position and rotation
type TransformComponent struct {
	X, Y     float64
	Rotation float64
	Scale    float64
}

func (t *TransformComponent) Type() ComponentType {
	return ComponentTypeTransform
}

// RenderComponent represents visual information
type RenderComponent struct {
	Visible    bool
	Layer      int
	SpritePath string
	Color      uint32
	Alpha      float32
}

func (r *RenderComponent) Type() ComponentType {
	return ComponentTypeRender
}

// PhysicsComponent represents physics properties
type PhysicsComponent struct {
	VelocityX, VelocityY         float64
	AccelerationX, AccelerationY float64
	Mass                         float64
	Friction                     float64
}

func (p *PhysicsComponent) Type() ComponentType {
	return ComponentTypePhysics
}

// InventoryComponent represents an inventory container
type InventoryComponent struct {
	Capacity int
	Items    []EntityID // References to item entities
}

func (i *InventoryComponent) Type() ComponentType {
	return ComponentTypeInventory
}

// MerchantComponent represents a merchant NPC
type MerchantComponent struct {
	Name        string
	Gold        int
	Personality string
	Strategy    string
}

func (m *MerchantComponent) Type() ComponentType {
	return ComponentTypeMerchant
}

// ShopComponent represents a shop
type ShopComponent struct {
	Name       string
	Owner      EntityID
	Reputation float64
	Level      int
}

func (s *ShopComponent) Type() ComponentType {
	return ComponentTypeShop
}

// PriceComponent represents pricing information
type PriceComponent struct {
	BasePrice    int
	CurrentPrice int
	Discount     float64
}

func (p *PriceComponent) Type() ComponentType {
	return ComponentTypePrice
}

// AIComponent represents AI behavior
type AIComponent struct {
	BehaviorType string
	State        string
	Target       EntityID
	UpdateRate   float64 // Updates per second
}

func (a *AIComponent) Type() ComponentType {
	return ComponentTypeAI
}

// PlayerComponent represents the player
type PlayerComponent struct {
	Name string
	Gold int
	Rank string
}

func (p *PlayerComponent) Type() ComponentType {
	return ComponentTypePlayer
}

// ItemComponent represents an item
type ItemComponent struct {
	ItemID       string
	Name         string
	Category     string
	Quantity     int
	Durability   float64
	IsPerishable bool
}

func (i *ItemComponent) Type() ComponentType {
	return ComponentTypeItem
}
