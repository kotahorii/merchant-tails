package ecs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEntity(t *testing.T) {
	entity := NewEntity()
	assert.NotNil(t, entity)
	assert.NotEqual(t, EntityID(0), entity.ID())
	assert.True(t, entity.IsAlive())
	assert.Equal(t, 0, len(entity.GetComponentTypes()))
}

func TestEntityIDGeneration(t *testing.T) {
	// Entity IDs should be unique
	entity1 := NewEntity()
	entity2 := NewEntity()
	entity3 := NewEntity()

	assert.NotEqual(t, entity1.ID(), entity2.ID())
	assert.NotEqual(t, entity2.ID(), entity3.ID())
	assert.NotEqual(t, entity1.ID(), entity3.ID())
}

func TestEntityAddComponent(t *testing.T) {
	entity := NewEntity()

	// Create test components
	position := &PositionComponent{X: 10, Y: 20}
	velocity := &VelocityComponent{DX: 5, DY: -3}

	// Add components
	entity.AddComponent(position)
	entity.AddComponent(velocity)

	// Check components exist
	assert.True(t, entity.HasComponent(ComponentTypePosition))
	assert.True(t, entity.HasComponent(ComponentTypeVelocity))
	assert.False(t, entity.HasComponent(ComponentTypeHealth))

	// Check component count
	types := entity.GetComponentTypes()
	assert.Equal(t, 2, len(types))
}

func TestEntityGetComponent(t *testing.T) {
	entity := NewEntity()

	// Add position component
	position := &PositionComponent{X: 100, Y: 200}
	entity.AddComponent(position)

	// Get component
	retrieved := entity.GetComponent(ComponentTypePosition)
	require.NotNil(t, retrieved)

	// Type assert and verify
	pos, ok := retrieved.(*PositionComponent)
	require.True(t, ok)
	assert.Equal(t, float64(100), pos.X)
	assert.Equal(t, float64(200), pos.Y)

	// Try to get non-existent component
	nonExistent := entity.GetComponent(ComponentTypeHealth)
	assert.Nil(t, nonExistent)
}

func TestEntityRemoveComponent(t *testing.T) {
	entity := NewEntity()

	// Add components
	position := &PositionComponent{X: 10, Y: 20}
	velocity := &VelocityComponent{DX: 5, DY: -3}
	entity.AddComponent(position)
	entity.AddComponent(velocity)

	// Remove position component
	entity.RemoveComponent(ComponentTypePosition)

	// Check component was removed
	assert.False(t, entity.HasComponent(ComponentTypePosition))
	assert.True(t, entity.HasComponent(ComponentTypeVelocity))
	assert.Equal(t, 1, len(entity.GetComponentTypes()))
}

func TestEntityDestroy(t *testing.T) {
	entity := NewEntity()
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})

	// Entity should be alive initially
	assert.True(t, entity.IsAlive())

	// Destroy entity
	entity.Destroy()

	// Entity should not be alive
	assert.False(t, entity.IsAlive())

	// Components should be cleared
	assert.Equal(t, 0, len(entity.GetComponentTypes()))
}

func TestEntityManager(t *testing.T) {
	manager := NewEntityManager()
	assert.NotNil(t, manager)

	// Create entities
	entity1 := manager.CreateEntity()
	entity2 := manager.CreateEntity()
	entity3 := manager.CreateEntity()

	assert.NotNil(t, entity1)
	assert.NotNil(t, entity2)
	assert.NotNil(t, entity3)

	// All entities should have unique IDs
	assert.NotEqual(t, entity1.ID(), entity2.ID())
	assert.NotEqual(t, entity2.ID(), entity3.ID())

	// Get entity by ID
	retrieved := manager.GetEntity(entity1.ID())
	assert.Equal(t, entity1, retrieved)

	// Get all entities
	all := manager.GetAllEntities()
	assert.Equal(t, 3, len(all))

	// Destroy entity
	manager.DestroyEntity(entity2.ID())
	all = manager.GetAllEntities()
	assert.Equal(t, 2, len(all))

	// Entity should be marked as not alive
	assert.False(t, entity2.IsAlive())
}

func TestEntityManagerQuery(t *testing.T) {
	manager := NewEntityManager()

	// Create entities with different component combinations
	entity1 := manager.CreateEntity()
	entity1.AddComponent(&PositionComponent{X: 10, Y: 20})
	entity1.AddComponent(&VelocityComponent{DX: 1, DY: 2})

	entity2 := manager.CreateEntity()
	entity2.AddComponent(&PositionComponent{X: 30, Y: 40})
	entity2.AddComponent(&HealthComponent{Current: 100, Max: 100})

	entity3 := manager.CreateEntity()
	entity3.AddComponent(&PositionComponent{X: 50, Y: 60})
	entity3.AddComponent(&VelocityComponent{DX: 3, DY: 4})
	entity3.AddComponent(&HealthComponent{Current: 50, Max: 100})

	// Query entities with Position component
	withPosition := manager.QueryEntities(ComponentTypePosition)
	assert.Equal(t, 3, len(withPosition))

	// Query entities with Velocity component
	withVelocity := manager.QueryEntities(ComponentTypeVelocity)
	assert.Equal(t, 2, len(withVelocity))

	// Query entities with Health component
	withHealth := manager.QueryEntities(ComponentTypeHealth)
	assert.Equal(t, 2, len(withHealth))

	// Query entities with Position AND Velocity
	withPosAndVel := manager.QueryEntitiesMultiple(
		ComponentTypePosition,
		ComponentTypeVelocity,
	)
	assert.Equal(t, 2, len(withPosAndVel))

	// Query entities with all three components
	withAll := manager.QueryEntitiesMultiple(
		ComponentTypePosition,
		ComponentTypeVelocity,
		ComponentTypeHealth,
	)
	assert.Equal(t, 1, len(withAll))
	assert.Equal(t, entity3.ID(), withAll[0].ID())
}

func TestEntityManagerClear(t *testing.T) {
	manager := NewEntityManager()

	// Create some entities
	manager.CreateEntity()
	manager.CreateEntity()
	manager.CreateEntity()

	assert.Equal(t, 3, len(manager.GetAllEntities()))

	// Clear all entities
	manager.Clear()

	assert.Equal(t, 0, len(manager.GetAllEntities()))
}

// Test components
type PositionComponent struct {
	X, Y float64
}

func (p *PositionComponent) Type() ComponentType {
	return ComponentTypePosition
}

type VelocityComponent struct {
	DX, DY float64
}

func (v *VelocityComponent) Type() ComponentType {
	return ComponentTypeVelocity
}

type HealthComponent struct {
	Current, Max int
}

func (h *HealthComponent) Type() ComponentType {
	return ComponentTypeHealth
}

// Component types for testing
const (
	ComponentTypePosition ComponentType = iota
	ComponentTypeVelocity
	ComponentTypeHealth
)
