package ecs

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemInterface(t *testing.T) {
	// Test that our test system implements the System interface
	var _ System = (*testSystem)(nil)
}

func TestSystemManager(t *testing.T) {
	manager := NewSystemManager()
	assert.NotNil(t, manager)

	// Add systems
	movementSys := &MovementSystem{}
	renderSys := &RenderSystem{}

	manager.AddSystem("movement", movementSys, 1)
	manager.AddSystem("render", renderSys, 2)

	// Get system
	retrieved := manager.GetSystem("movement")
	assert.Equal(t, movementSys, retrieved)

	// Remove system
	manager.RemoveSystem("render")
	assert.Nil(t, manager.GetSystem("render"))
}

func TestSystemManagerUpdate(t *testing.T) {
	world := NewWorld()
	manager := world.GetSystemManager()

	// Create a test system that counts updates
	testSys := &testSystem{
		updateCount: 0,
	}
	manager.AddSystem("test", testSys, 1)

	// Update systems
	ctx := context.Background()
	deltaTime := 16 * time.Millisecond // ~60 FPS

	err := manager.Update(ctx, world, deltaTime)
	assert.NoError(t, err)
	assert.Equal(t, 1, testSys.updateCount)

	// Update again
	err = manager.Update(ctx, world, deltaTime)
	assert.NoError(t, err)
	assert.Equal(t, 2, testSys.updateCount)
}

func TestSystemPriority(t *testing.T) {
	world := NewWorld()
	manager := world.GetSystemManager()

	var executionOrder []string

	// Add systems with different priorities
	sys1 := &recordingSystem{name: "sys1", order: &executionOrder}
	sys2 := &recordingSystem{name: "sys2", order: &executionOrder}
	sys3 := &recordingSystem{name: "sys3", order: &executionOrder}

	manager.AddSystem("sys1", sys1, 3) // Lowest priority
	manager.AddSystem("sys2", sys2, 1) // Highest priority
	manager.AddSystem("sys3", sys3, 2) // Middle priority

	// Update should execute in priority order
	ctx := context.Background()
	err := manager.Update(ctx, world, time.Millisecond)
	assert.NoError(t, err)

	// Check execution order (lower number = higher priority)
	require.Equal(t, 3, len(executionOrder))
	assert.Equal(t, "sys2", executionOrder[0]) // Priority 1
	assert.Equal(t, "sys3", executionOrder[1]) // Priority 2
	assert.Equal(t, "sys1", executionOrder[2]) // Priority 3
}

func TestMovementSystem(t *testing.T) {
	world := NewWorld()
	entityManager := world.GetEntityManager()

	// Create entity with position and physics
	entity := entityManager.CreateEntity()
	transform := &TransformComponent{X: 100, Y: 100}
	physics := &PhysicsComponent{VelocityX: 10, VelocityY: -5}
	entity.AddComponent(transform)
	entity.AddComponent(physics)

	// Create movement system
	movementSys := &MovementSystem{}

	// Update system
	ctx := context.Background()
	deltaTime := 100 * time.Millisecond // 0.1 seconds

	err := movementSys.Update(ctx, world, deltaTime)
	assert.NoError(t, err)

	// Check position was updated based on velocity
	updatedTransform := entity.GetComponent(ComponentTypeTransform).(*TransformComponent)
	assert.Equal(t, 101.0, updatedTransform.X) // 100 + (10 * 0.1)
	assert.Equal(t, 99.5, updatedTransform.Y)  // 100 + (-5 * 0.1)
}

func TestPhysicsSystem(t *testing.T) {
	world := NewWorld()
	entityManager := world.GetEntityManager()

	// Create entity with physics
	entity := entityManager.CreateEntity()
	physics := &PhysicsComponent{
		VelocityX:     10,
		VelocityY:     20,
		AccelerationX: 5,
		AccelerationY: -10,
		Friction:      0.1,
	}
	entity.AddComponent(physics)

	// Create physics system
	physicsSys := &PhysicsSystem{}

	// Update system
	ctx := context.Background()
	deltaTime := 100 * time.Millisecond // 0.1 seconds

	err := physicsSys.Update(ctx, world, deltaTime)
	assert.NoError(t, err)

	// Check velocity was updated based on acceleration and friction
	// The physics system applies: velocity += acceleration * dt, then velocity -= velocity * friction * dt
	// Friction is applied to the UPDATED velocity
	// VelocityX after accel = 10 + (5 * 0.1) = 10.5
	// VelocityX after friction = 10.5 - (10.5 * 0.1 * 0.1) = 10.5 - 0.105 = 10.395
	// VelocityY after accel = 20 + (-10 * 0.1) = 19
	// VelocityY after friction = 19 - (19 * 0.1 * 0.1) = 19 - 0.19 = 18.81
	assert.InDelta(t, 10.395, physics.VelocityX, 0.02)
	assert.InDelta(t, 18.81, physics.VelocityY, 0.02)
}

// Test implementations

type testSystem struct {
	updateCount int
	shouldError bool
}

func (ts *testSystem) Name() string {
	return "test"
}

func (ts *testSystem) Update(ctx context.Context, world *World, deltaTime time.Duration) error {
	ts.updateCount++
	if ts.shouldError {
		return assert.AnError
	}
	return nil
}

func (ts *testSystem) GetRequiredComponents() []ComponentType {
	return []ComponentType{}
}

type recordingSystem struct {
	name  string
	order *[]string
}

func (rs *recordingSystem) Name() string {
	return rs.name
}

func (rs *recordingSystem) Update(ctx context.Context, world *World, deltaTime time.Duration) error {
	*rs.order = append(*rs.order, rs.name)
	return nil
}

func (rs *recordingSystem) GetRequiredComponents() []ComponentType {
	return []ComponentType{}
}
