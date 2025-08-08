package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"connectrpc.com/connect"
	savepb "github.com/yourusername/merchant-tails/game/internal/gen/save"
	"github.com/yourusername/merchant-tails/game/internal/gen/service"
	"github.com/yourusername/merchant-tails/game/internal/gen/service/serviceconnect"
	"github.com/yourusername/merchant-tails/game/internal/infrastructure/save"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GameServer implements the Connect-RPC GameService
type GameServer struct {
	serviceconnect.UnimplementedGameServiceHandler

	// Game state management
	currentState    *savepb.GameState
	saveManager     *save.SaveManager
	settingsManager interface{} // TODO: Add settings manager

	// Stream management
	stateStreams  map[string]chan *service.GameStateUpdate
	marketStreams map[string]chan *service.MarketDataUpdate
	eventStreams  map[string]chan *service.EventUpdate

	// Game control
	isPaused  bool
	gameSpeed float64

	// Thread safety
	mu       sync.RWMutex
	streamMu sync.RWMutex
}

// NewGameServer creates a new game server instance
func NewGameServer(saveDir string) *GameServer {
	return &GameServer{
		saveManager:   save.NewSaveManager(saveDir),
		stateStreams:  make(map[string]chan *service.GameStateUpdate),
		marketStreams: make(map[string]chan *service.MarketDataUpdate),
		eventStreams:  make(map[string]chan *service.EventUpdate),
		gameSpeed:     1.0,
	}
}

// SaveGame implements the SaveGame RPC
func (s *GameServer) SaveGame(
	ctx context.Context,
	req *connect.Request[service.SaveGameRequest],
) (*connect.Response[service.SaveGameResponse], error) {
	s.mu.RLock()
	state := s.currentState
	s.mu.RUnlock()

	if state == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("no game state to save"))
	}

	options := &save.SaveOptions{
		Compress:  req.Msg.Compress,
		Encrypt:   req.Msg.Encrypt,
		Overwrite: true,
		Metadata:  req.Msg.Metadata,
	}

	err := s.saveManager.Save(state, req.Msg.SaveName, options)
	if err != nil {
		return connect.NewResponse(&service.SaveGameResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}), nil
	}

	return connect.NewResponse(&service.SaveGameResponse{
		Success:  true,
		SavePath: req.Msg.SaveName,
		SaveTime: timestamppb.Now(),
	}), nil
}

// LoadGame implements the LoadGame RPC
func (s *GameServer) LoadGame(
	ctx context.Context,
	req *connect.Request[service.LoadGameRequest],
) (*connect.Response[service.LoadGameResponse], error) {
	state, err := s.saveManager.Load(req.Msg.SaveName)
	if err != nil {
		return connect.NewResponse(&service.LoadGameResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}), nil
	}

	s.mu.Lock()
	s.currentState = state
	s.mu.Unlock()

	// Notify all state streams
	s.broadcastStateUpdate(&service.GameStateUpdate{
		Type:      service.GameStateUpdate_UPDATE_TYPE_FULL,
		Timestamp: timestamppb.Now(),
		Update: &service.GameStateUpdate_FullState{
			FullState: state,
		},
	})

	return connect.NewResponse(&service.LoadGameResponse{
		Success:   true,
		GameState: state,
	}), nil
}

// QuickSave implements the QuickSave RPC
func (s *GameServer) QuickSave(
	ctx context.Context,
	req *connect.Request[service.QuickSaveRequest],
) (*connect.Response[service.SaveGameResponse], error) {
	s.mu.RLock()
	state := s.currentState
	s.mu.RUnlock()

	if state == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("no game state to save"))
	}

	err := s.saveManager.QuickSave(state)
	if err != nil {
		return connect.NewResponse(&service.SaveGameResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}), nil
	}

	return connect.NewResponse(&service.SaveGameResponse{
		Success:  true,
		SaveTime: timestamppb.Now(),
	}), nil
}

// AutoSave implements the AutoSave RPC
func (s *GameServer) AutoSave(
	ctx context.Context,
	req *connect.Request[service.AutoSaveRequest],
) (*connect.Response[service.SaveGameResponse], error) {
	s.mu.RLock()
	state := s.currentState
	s.mu.RUnlock()

	if state == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("no game state to save"))
	}

	err := s.saveManager.AutoSave(state)
	if err != nil {
		return connect.NewResponse(&service.SaveGameResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}), nil
	}

	return connect.NewResponse(&service.SaveGameResponse{
		Success:  true,
		SaveTime: timestamppb.Now(),
	}), nil
}

// ListSaves implements the ListSaves RPC
func (s *GameServer) ListSaves(
	ctx context.Context,
	req *connect.Request[service.ListSavesRequest],
) (*connect.Response[service.ListSavesResponse], error) {
	saves, err := s.saveManager.ListSaves()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert to protobuf format
	pbSaves := make([]*service.SaveInfo, 0, len(saves))
	for _, save := range saves {
		// Filter by type if requested
		if req.Msg.FilterType != "" && req.Msg.FilterType != "all" {
			switch req.Msg.FilterType {
			case "auto":
				if !save.IsAutoSave {
					continue
				}
			case "quick":
				if !save.IsQuickSave {
					continue
				}
			case "manual":
				if save.IsAutoSave || save.IsQuickSave {
					continue
				}
			}
		}

		pbSave := &service.SaveInfo{
			SaveName:    save.FileName,
			SavePath:    save.FilePath,
			SaveTime:    timestamppb.New(save.SaveTime),
			FileSize:    save.FileSize,
			Version:     save.Version,
			PlayerName:  save.PlayerName,
			DayNumber:   save.DayNumber,
			Gold:        save.Gold,
			IsAutoSave:  save.IsAutoSave,
			IsQuickSave: save.IsQuickSave,
		}
		pbSaves = append(pbSaves, pbSave)
	}

	// Apply pagination
	start := 0
	end := len(pbSaves)

	if req.Msg.Offset > 0 && req.Msg.Offset < int32(len(pbSaves)) {
		start = int(req.Msg.Offset)
	}

	if req.Msg.Limit > 0 && int(req.Msg.Limit) < end-start {
		end = start + int(req.Msg.Limit)
	}

	return connect.NewResponse(&service.ListSavesResponse{
		Saves:      pbSaves[start:end],
		TotalCount: int32(len(pbSaves)),
	}), nil
}

// DeleteSave implements the DeleteSave RPC
func (s *GameServer) DeleteSave(
	ctx context.Context,
	req *connect.Request[service.DeleteSaveRequest],
) (*connect.Response[emptypb.Empty], error) {
	err := s.saveManager.DeleteSave(req.Msg.SaveName)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

// GetGameState implements the GetGameState RPC
func (s *GameServer) GetGameState(
	ctx context.Context,
	req *connect.Request[emptypb.Empty],
) (*connect.Response[savepb.GameState], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.currentState == nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("no game state available"))
	}

	return connect.NewResponse(s.currentState), nil
}

// UpdateGameState implements the UpdateGameState RPC
func (s *GameServer) UpdateGameState(
	ctx context.Context,
	req *connect.Request[service.UpdateGameStateRequest],
) (*connect.Response[emptypb.Empty], error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// TODO: Implement partial updates based on update_fields
	s.currentState = req.Msg.GameState

	// Broadcast update to streams
	s.broadcastStateUpdate(&service.GameStateUpdate{
		Type:      service.GameStateUpdate_UPDATE_TYPE_FULL,
		Timestamp: timestamppb.Now(),
		Update: &service.GameStateUpdate_FullState{
			FullState: s.currentState,
		},
	})

	return connect.NewResponse(&emptypb.Empty{}), nil
}

// PauseGame implements the PauseGame RPC
func (s *GameServer) PauseGame(
	ctx context.Context,
	req *connect.Request[emptypb.Empty],
) (*connect.Response[emptypb.Empty], error) {
	s.mu.Lock()
	s.isPaused = true
	s.mu.Unlock()

	return connect.NewResponse(&emptypb.Empty{}), nil
}

// ResumeGame implements the ResumeGame RPC
func (s *GameServer) ResumeGame(
	ctx context.Context,
	req *connect.Request[emptypb.Empty],
) (*connect.Response[emptypb.Empty], error) {
	s.mu.Lock()
	s.isPaused = false
	s.mu.Unlock()

	return connect.NewResponse(&emptypb.Empty{}), nil
}

// SetGameSpeed implements the SetGameSpeed RPC
func (s *GameServer) SetGameSpeed(
	ctx context.Context,
	req *connect.Request[service.SetGameSpeedRequest],
) (*connect.Response[emptypb.Empty], error) {
	if req.Msg.Speed < 0.1 || req.Msg.Speed > 10.0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("game speed must be between 0.1 and 10.0"))
	}

	s.mu.Lock()
	s.gameSpeed = req.Msg.Speed
	s.mu.Unlock()

	return connect.NewResponse(&emptypb.Empty{}), nil
}

// StreamGameState implements the StreamGameState RPC
func (s *GameServer) StreamGameState(
	ctx context.Context,
	req *connect.Request[emptypb.Empty],
	stream *connect.ServerStream[service.GameStateUpdate],
) error {
	// Create a unique stream ID
	streamID := fmt.Sprintf("state_%d", time.Now().UnixNano())
	updateChan := make(chan *service.GameStateUpdate, 100)

	// Register the stream
	s.streamMu.Lock()
	s.stateStreams[streamID] = updateChan
	s.streamMu.Unlock()

	// Clean up on exit
	defer func() {
		s.streamMu.Lock()
		delete(s.stateStreams, streamID)
		s.streamMu.Unlock()
		close(updateChan)
	}()

	// Send initial state
	s.mu.RLock()
	if s.currentState != nil {
		initialUpdate := &service.GameStateUpdate{
			Type:      service.GameStateUpdate_UPDATE_TYPE_FULL,
			Timestamp: timestamppb.Now(),
			Update: &service.GameStateUpdate_FullState{
				FullState: s.currentState,
			},
		}
		s.mu.RUnlock()

		if err := stream.Send(initialUpdate); err != nil {
			return err
		}
	} else {
		s.mu.RUnlock()
	}

	// Stream updates
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update := <-updateChan:
			if err := stream.Send(update); err != nil {
				return err
			}
		}
	}
}

// StreamMarketData implements the StreamMarketData RPC
func (s *GameServer) StreamMarketData(
	ctx context.Context,
	req *connect.Request[emptypb.Empty],
	stream *connect.ServerStream[service.MarketDataUpdate],
) error {
	streamID := fmt.Sprintf("market_%d", time.Now().UnixNano())
	updateChan := make(chan *service.MarketDataUpdate, 100)

	s.streamMu.Lock()
	s.marketStreams[streamID] = updateChan
	s.streamMu.Unlock()

	defer func() {
		s.streamMu.Lock()
		delete(s.marketStreams, streamID)
		s.streamMu.Unlock()
		close(updateChan)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update := <-updateChan:
			if err := stream.Send(update); err != nil {
				return err
			}
		}
	}
}

// StreamEvents implements the StreamEvents RPC
func (s *GameServer) StreamEvents(
	ctx context.Context,
	req *connect.Request[emptypb.Empty],
	stream *connect.ServerStream[service.EventUpdate],
) error {
	streamID := fmt.Sprintf("event_%d", time.Now().UnixNano())
	updateChan := make(chan *service.EventUpdate, 100)

	s.streamMu.Lock()
	s.eventStreams[streamID] = updateChan
	s.streamMu.Unlock()

	defer func() {
		s.streamMu.Lock()
		delete(s.eventStreams, streamID)
		s.streamMu.Unlock()
		close(updateChan)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update := <-updateChan:
			if err := stream.Send(update); err != nil {
				return err
			}
		}
	}
}

// Helper methods

// broadcastStateUpdate sends an update to all state streams
func (s *GameServer) broadcastStateUpdate(update *service.GameStateUpdate) {
	s.streamMu.RLock()
	defer s.streamMu.RUnlock()

	for _, ch := range s.stateStreams {
		select {
		case ch <- update:
		default:
			// Skip if channel is full
		}
	}
}

// broadcastMarketUpdate sends an update to all market streams
func (s *GameServer) broadcastMarketUpdate(update *service.MarketDataUpdate) {
	s.streamMu.RLock()
	defer s.streamMu.RUnlock()

	for _, ch := range s.marketStreams {
		select {
		case ch <- update:
		default:
			// Skip if channel is full
		}
	}
}

// broadcastEventUpdate sends an update to all event streams
func (s *GameServer) broadcastEventUpdate(update *service.EventUpdate) {
	s.streamMu.RLock()
	defer s.streamMu.RUnlock()

	for _, ch := range s.eventStreams {
		select {
		case ch <- update:
		default:
			// Skip if channel is full
		}
	}
}

// UpdateGold updates the player's gold and broadcasts the change
func (s *GameServer) UpdateGold(gold float64) {
	s.mu.Lock()
	if s.currentState != nil && s.currentState.Player != nil {
		s.currentState.Player.Gold = gold
	}
	s.mu.Unlock()

	s.broadcastStateUpdate(&service.GameStateUpdate{
		Type:      service.GameStateUpdate_UPDATE_TYPE_GOLD,
		Timestamp: timestamppb.Now(),
		Update: &service.GameStateUpdate_Gold{
			Gold: gold,
		},
	})
}

// UpdateInventory updates the inventory and broadcasts the change
func (s *GameServer) UpdateInventory(inventory *savepb.InventoryData) {
	s.mu.Lock()
	if s.currentState != nil {
		s.currentState.Inventory = inventory
	}
	s.mu.Unlock()

	s.broadcastStateUpdate(&service.GameStateUpdate{
		Type:      service.GameStateUpdate_UPDATE_TYPE_INVENTORY,
		Timestamp: timestamppb.Now(),
		Update: &service.GameStateUpdate_Inventory{
			Inventory: inventory,
		},
	})
}

// Start starts the Connect-RPC server
func (s *GameServer) Start(addr string) error {
	mux := http.NewServeMux()

	// Register the game service
	path, handler := serviceconnect.NewGameServiceHandler(s)
	mux.Handle(path, handler)

	// Start auto-save if configured
	s.saveManager.StartAutoSave(func() *savepb.GameState {
		s.mu.RLock()
		defer s.mu.RUnlock()
		return s.currentState
	})

	log.Printf("Starting Connect-RPC server on %s", addr)
	return http.ListenAndServe(addr, mux)
}
