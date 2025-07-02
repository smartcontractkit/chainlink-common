package requests

import (
	"context"
	"time"

	"github.com/jonboulle/clockwork"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

type responseCacheEntry[R ConsensusResponse] struct {
	response  R
	entryTime time.Time
}

type ConsensusRequest[T any, R ConsensusResponse] interface {
	ID() string
	Copy() T
	ExpiryTime() time.Time
	SendResponse(ctx context.Context, response R)
	SendTimeout(ctx context.Context)
}

type ConsensusResponse interface {
	RequestID() string
}

type Handler[T ConsensusRequest[T, R], R ConsensusResponse] struct {
	services.Service
	eng *services.Engine

	store *Store[T]

	pendingRequests map[string]T

	responseCache   map[string]*responseCacheEntry[R]
	cacheExpiryTime time.Duration

	responseCh chan R
	requestCh  chan T

	clock clockwork.Clock
}

func NewHandler[T ConsensusRequest[T, R], R ConsensusResponse](lggr logger.Logger, s *Store[T], clock clockwork.Clock, responseExpiryTime time.Duration) *Handler[T, R] {
	h := &Handler[T, R]{
		store:           s,
		pendingRequests: map[string]T{},
		responseCache:   map[string]*responseCacheEntry[R]{},
		responseCh:      make(chan R),
		requestCh:       make(chan T),
		clock:           clock,
		cacheExpiryTime: responseExpiryTime,
	}
	h.Service, h.eng = services.Config{
		Name:  "Handler",
		Start: h.start,
	}.NewServiceEngine(lggr)
	return h
}

func (h *Handler[T, R]) SendResponse(ctx context.Context, resp R) {
	select {
	case <-ctx.Done():
		return
	case h.responseCh <- resp:
	}
}

func (h *Handler[T, R]) SendRequest(ctx context.Context, r T) {
	select {
	case <-ctx.Done():
		return
	case h.requestCh <- r:
	}
}

func (h *Handler[T, R]) start(_ context.Context) error {
	h.eng.Go(h.worker)
	return nil
}

func (h *Handler[T, R]) worker(ctx context.Context) {
	responseCacheExpiryTicker := h.clock.NewTicker(h.cacheExpiryTime)
	defer responseCacheExpiryTicker.Stop()

	// Set to tick at 1 second as this is a sufficient resolution for expiring requests without causing too much overhead
	pendingRequestsExpiryTicker := h.clock.NewTicker(1 * time.Second)
	defer pendingRequestsExpiryTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pendingRequestsExpiryTicker.Chan():
			h.expirePendingRequests(ctx)
		case <-responseCacheExpiryTicker.Chan():
			h.expireCachedResponses()
		case req := <-h.requestCh:
			h.pendingRequests[req.ID()] = req

			existingResponse := h.responseCache[req.ID()]
			if existingResponse != nil {
				delete(h.responseCache, req.ID())
				h.eng.Debugw("Found cached response for request", "requestID", req.ID)
				h.sendResponse(ctx, req, existingResponse.response)
				continue
			}

			if err := h.store.Add(req); err != nil {
				h.eng.Errorw("failed to add request to store", "err", err)
			}

		case resp := <-h.responseCh:
			req, wasPresent := h.store.Evict(resp.RequestID())
			if !wasPresent {
				h.responseCache[resp.RequestID()] = &responseCacheEntry[R]{
					response:  resp,
					entryTime: h.clock.Now(),
				}
				h.eng.Debugw("Caching response without request", "requestID", resp.RequestID())
				continue
			}

			h.sendResponse(ctx, req, resp)
		}
	}
}

func (h *Handler[T, R]) sendResponse(ctx context.Context, req T, resp R) {
	req.SendResponse(ctx, resp)
	delete(h.pendingRequests, req.ID())
}

func (h *Handler[T, R]) sendTimeout(ctx context.Context, req T) {
	req.SendTimeout(ctx)
	delete(h.pendingRequests, req.ID())
}

func (h *Handler[T, R]) expirePendingRequests(ctx context.Context) {
	now := h.clock.Now()

	for _, req := range h.pendingRequests {
		if now.After(req.ExpiryTime()) {
			h.store.Evict(req.ID())
			h.sendTimeout(ctx, req)
		}
	}
}

func (h *Handler[T, R]) expireCachedResponses() {
	for k, v := range h.responseCache {
		if h.clock.Since(v.entryTime) > h.cacheExpiryTime {
			delete(h.responseCache, k)
			h.eng.Debugw("Expired response", "requestID", k)
		}
	}
}
