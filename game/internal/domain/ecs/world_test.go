package ecs

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWorld(t *testing.T) {
	world := NewWorld()
	assert.NotNil(t, world)
	assert.NotNil(t, world.GetEntityManager())
	assert.NotNil(t, world.GetSystemManager())
}

func TestWorldResources(t *testing.T) {
	world := NewWorld()

	// Add resources
	config := map[string]int{"maxEntities": 1000}
	world.AddResource("config", config)

	playerData := struct {
		Name string
		Gold int
	}{"Player1", 1000}
	world.AddResource("player", playerData)

	// Get resources
	retrievedConfig := world.GetResource("config")
	assert.Equal(t, config, retrievedConfig)

	retrievedPlayer := world.GetResource("player")
	assert.Equal(t, playerData, retrievedPlayer)

	// Remove resource
	world.RemoveResource("config")
	assert.Nil(t, world.GetResource("config"))
	assert.NotNil(t, world.GetResource("player"))
}

func TestWorldEntityOperations(t *testing.T) {
	world := NewWorld()

	// Create entities through world
	entity1 := world.CreateEntity()
	entity2 := world.CreateEntity()

	assert.NotNil(t, entity1)
	assert.NotNil(t, entity2)
	assert.NotEqual(t, entity1.ID(), entity2.ID())

	// Add components
	entity1.AddComponent(&TransformComponent{X: 10, Y: 20})
	entity2.AddComponent(&TransformComponent{X: 30, Y: 40})
	entity2.AddComponent(&PhysicsComponent{VelocityX: 5})

	// Query through world
	withTransform := world.QueryEntities(ComponentTypeTransform)
	assert.Equal(t, 2, len(withTransform))

	withPhysics := world.QueryEntities(ComponentTypePhysics)
	assert.Equal(t, 1, len(withPhysics))

	// Destroy entity through world
	world.DestroyEntity(entity1.ID())
	assert.False(t, entity1.IsAlive())

	withTransform = world.QueryEntities(ComponentTypeTransform)
	assert.Equal(t, 1, len(withTransform))
}

func TestWorldSystemOperations(t *testing.T) {
	world := NewWorld()

	// Add systems through world
	movementSys := &MovementSystem{}
	physicsSys := &PhysicsSystem{}

	world.AddSystem("movement", movementSys, 2)
	world.AddSystem("physics", physicsSys, 1)

	// Get system through world
	retrieved := world.GetSystem("movement")
	assert.Equal(t, movementSys, retrieved)

	// Remove system through world
	world.RemoveSystem("physics")
	assert.Nil(t, world.GetSystem("physics"))
}

func TestWorldUpdate(t *testing.T) {
	world := NewWorld()

	// Create entity with components
	entity := world.CreateEntity()
	entity.AddComponent(&TransformComponent{X: 0, Y: 0})
	entity.AddComponent(&PhysicsComponent{VelocityX: 100, VelocityY: 50})

	// Add movement system
	world.AddSystem("movement", &MovementSystem{}, 1)

	// Update world
	ctx := context.Background()
	deltaTime := 100 * time.Millisecond // 0.1 seconds

	err := world.Update(ctx, deltaTime)
	assert.NoError(t, err)

	// Check position was updated
	transform := entity.GetComponent(ComponentTypeTransform).(*TransformComponent)
	assert.Equal(t, 10.0, transform.X) // 100 * 0.1
	assert.Equal(t, 5.0, transform.Y)  // 50 * 0.1
}

func TestWorldClear(t *testing.T) {
	world := NewWorld()

	// Add entities and resources
	world.CreateEntity()
	world.CreateEntity()
	world.AddResource("test", "value")

	// Verify they exist
	assert.Equal(t, 2, len(world.GetEntityManager().GetAllEntities()))
	assert.NotNil(t, world.GetResource("test"))

	// Clear world
	world.Clear()

	// Verify everything is cleared
	assert.Equal(t, 0, len(world.GetEntityManager().GetAllEntities()))
	assert.Nil(t, world.GetResource("test"))
}

func TestWorldBuilder(t *testing.T) {
	// Build world with fluent interface
	world := NewWorldBuilder().
		WithSystem("movement", &MovementSystem{}, 2).
		WithSystem("physics", &PhysicsSystem{}, 1).
		WithResource("config", map[string]int{"fps": 60}).
		WithResource("debug", true).
		Build()

	// Verify systems were added
	assert.NotNil(t, world.GetSystem("movement"))
	assert.NotNil(t, world.GetSystem("physics"))

	// Verify resources were added
	config := world.GetResource("config").(map[string]int)
	assert.Equal(t, 60, config["fps"])

	debug := world.GetResource("debug").(bool)
	assert.True(t, debug)
}

func TestCreateDefaultWorld(t *testing.T) {
	world := CreateDefaultWorld()

	// Verify default systems are present
	assert.NotNil(t, world.GetSystem("physics"))
	assert.NotNil(t, world.GetSystem("movement"))
	assert.NotNil(t, world.GetSystem("ai"))
	assert.NotNil(t, world.GetSystem("inventory"))
	assert.NotNil(t, world.GetSystem("render"))
}

func TestWorldIntegration(t *testing.T) {
	// Create world with systems
	world := NewWorldBuilder().
		WithSystem("physics", &PhysicsSystem{}, 1).
		WithSystem("movement", &MovementSystem{}, 2).
		Build()

	// Create game entities
	player := world.CreateEntity()
	player.AddComponent(&TransformComponent{X: 100, Y: 100})
	player.AddComponent(&PhysicsComponent{
		VelocityX:     50,
		VelocityY:     0,
		AccelerationX: 10,
		AccelerationY: -9.8,
		Friction:      0.1,
	})
	player.AddComponent(&PlayerComponent{
		Name: "Hero",
		Gold: 1000,
		Rank: "Apprentice",
	})

	enemy := world.CreateEntity()
	enemy.AddComponent(&TransformComponent{X: 200, Y: 100})
	enemy.AddComponent(&PhysicsComponent{
		VelocityX: -20,
		VelocityY: 0,
	})
	enemy.AddComponent(&AIComponent{
		BehaviorType: "aggressive",
		State:        "pursuing",
		Target:       player.ID(),
	})

	item := world.CreateEntity()
	item.AddComponent(&TransformComponent{X: 150, Y: 80})
	item.AddComponent(&ItemComponent{
		ItemID:   "sword_001",
		Name:     "Iron Sword",
		Category: "weapon",
		Quantity: 1,
	})

	// Run simulation
	ctx := context.Background()
	deltaTime := 16 * time.Millisecond // ~60 FPS

	for i := 0; i < 10; i++ {
		err := world.Update(ctx, deltaTime)
		require.NoError(t, err)
	}

	// Check final positions
	playerTransform := player.GetComponent(ComponentTypeTransform).(*TransformComponent)
	assert.Greater(t, playerTransform.X, 100.0) // Player moved right

	enemyTransform := enemy.GetComponent(ComponentTypeTransform).(*TransformComponent)
	assert.Less(t, enemyTransform.X, 200.0) // Enemy moved left

	// Check physics updates
	playerPhysics := player.GetComponent(ComponentTypePhysics).(*PhysicsComponent)
	assert.Greater(t, playerPhysics.VelocityX, 50.0) // Accelerated
	assert.Less(t, playerPhysics.VelocityY, 0.0)     // Gravity applied
}
