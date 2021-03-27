package ftx

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/c9s/bbgo/pkg/types"
)

type Stream struct {
	*types.StandardStream

	wsService *WebsocketService

	// publicOnly must be accessed atomically
	publicOnly int32
}

func NewStream(key, secret string) *Stream {
	wss := NewWebsocketService(key, secret)
	s := &Stream{
		StandardStream: &types.StandardStream{},
		wsService:      wss,
	}

	wss.OnMessage((&messageHandler{StandardStream: s.StandardStream}).handleMessage)
	return s
}

func (s *Stream) Connect(ctx context.Context) error {
	// If it's not public only, let's do the authentication.
	if atomic.LoadInt32(&s.publicOnly) == 0 {
		logger.Infof("subscribe private events")
		s.wsService.Subscribe(
			newLoginRequest(s.wsService.key, s.wsService.secret, time.Now()),
		)
	}

	if err := s.wsService.Connect(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Stream) SetPublicOnly() {
	atomic.StoreInt32(&s.publicOnly, 1)
}

func (s *Stream) Subscribe(channel types.Channel, symbol string, _ types.SubscribeOptions) {
	if channel != types.BookChannel {
		// TODO: return err
	}
	s.wsService.Subscribe(websocketRequest{
		Operation: subscribe,
		Channel:   orderbook,
		Market:    TrimUpperString(symbol),
	})
}

func (s *Stream) Close() error {
	if s.wsService != nil {
		return s.wsService.Close()
	}
	return nil
}
