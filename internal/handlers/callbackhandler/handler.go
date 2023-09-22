package callbackhandler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/evt/callback/internal/model"
)

const (
	badRequest      = http.StatusBadRequest
	internalError   = http.StatusInternalServerError
	objectRequestOK = "ok"
)

// CallbackHandler is a callback handler.
type CallbackHandler struct {
	objectService ObjectService
	testerService TesterService
}

// New creates a new callback service.
func New(service ObjectService, testerService TesterService) *CallbackHandler {
	return &CallbackHandler{
		objectService: service,
		testerService: testerService,
	}
}

// Post handles POST /callback requests
func (h *CallbackHandler) Post(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request model.CallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %s", err), badRequest)
		return
	}

	if len(request.ObjectIDs) == 0 {
		http.Error(w, "no object IDs provided", badRequest)
		return
	}

	var wg sync.WaitGroup
	receivedObjects := make(chan model.TesterObject, len(request.ObjectIDs))

	for i := range request.ObjectIDs {
		wg.Add(1)

		objectID := request.ObjectIDs[i]

		go func() {
			defer wg.Done()

			object, err := h.testerService.GetObject(ctx, objectID)
			if err != nil {
				log.Printf("[id: %d, total: %d] testerService.GetObject failed: %s\n", object.ID, len(request.ObjectIDs), err)
				return
			}

			receivedObjects <- object
		}()
	}

	go func() {
		wg.Wait()
		close(receivedObjects)
	}()

	var totalUpdated, totalReceived int

	for object := range receivedObjects {
		totalReceived++

		if !object.Online {
			continue
		}

		if err := h.objectService.UpdateObject(ctx, &model.DBObject{ID: object.ID}); err != nil {
			log.Printf("[id: %d] objectService.UpdateObject failed: %s\n", object.ID, err)
		} else {
			totalUpdated++
		}
	}

	log.Printf("objects: received=%d, updated(=online)=%d\n", totalReceived, totalUpdated)

	w.Write([]byte(objectRequestOK))
}
