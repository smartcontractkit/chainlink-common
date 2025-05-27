package requests

import (
	"context"
	"fmt"
	"time"

	"github.com/jonboulle/clockwork"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

type responseCacheEntry struct {
	response  Response
	entryTime time.Time
}

type Handler[T ConsensusRequest[T]] struct {
	services.Service
	eng *services.Engine

	store *Store[T]

	pendingRequests map[string]T

	responseCache   map[string]*responseCacheEntry
	cacheExpiryTime time.Duration

	responseCh chan Response
	requestCh  chan T

	clock clockwork.Clock
}

func NewHandler[T ConsensusRequest[T]](lggr logger.Logger, s *Store[T], clock clockwork.Clock, responseExpiryTime time.Duration) *Handler[T] {
	h := &Handler[T]{
		store:           s,
		pendingRequests: map[string]T{},
		responseCache:   map[string]*responseCacheEntry{},
		responseCh:      make(chan Response),
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

func (h *Handler[T]) SendResponse(ctx context.Context, resp Response) {
	select {
	case <-ctx.Done():
		return
	case h.responseCh <- resp:
	}
}

func (h *Handler[T]) SendRequest(ctx context.Context, r T) {
	select {
	case <-ctx.Done():
		return
	case h.requestCh <- r:
	}
}

func (h *Handler[T]) start(_ context.Context) error {
	h.eng.Go(h.worker)
	return nil
}

func (h *Handler[T]) worker(ctx context.Context) {
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
			req, wasPresent := h.store.evict(resp.WorkflowExecutionID)
			if !wasPresent {
				h.responseCache[resp.WorkflowExecutionID] = &responseCacheEntry{
					response:  resp,
					entryTime: h.clock.Now(),
				}
				h.eng.Debugw("Caching response without request", "workflowExecutionID", resp.WorkflowExecutionID)
				continue
			}

			h.sendResponse(ctx, req, resp)
		}
	}
}

func (h *Handler[T]) sendResponse(ctx context.Context, req T, resp Response) {
	req.SendResponse(ctx, resp)
	delete(h.pendingRequests, req.ID())
}

func (h *Handler[T]) expirePendingRequests(ctx context.Context) {
	now := h.clock.Now()

	for _, req := range h.pendingRequests {
		if now.After(req.ExpiryTime()) {
			resp := Response{
				WorkflowExecutionID: req.ID(),
				Err:                 fmt.Errorf("timeout exceeded: could not process request before expiry, request ID %s", req.ID()),
			}
			h.store.evict(req.ID())
			h.sendResponse(ctx, req, resp)
		}
	}
}

func (h *Handler[T]) expireCachedResponses() {
	for k, v := range h.responseCache {
		if h.clock.Since(v.entryTime) > h.cacheExpiryTime {
			delete(h.responseCache, k)
			h.eng.Debugw("Expired response", "requestID", k)
		}
	}
}
