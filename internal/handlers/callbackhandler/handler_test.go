package callbackhandler

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/evt/callback/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/golang/mock/gomock"
)

func TestPostObjectOffline(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	objectService := NewMockObjectService(ctrl)
	testerService := NewMockTesterService(ctrl)

	callbackHandler := New(objectService, testerService)

	ts := httptest.NewServer(http.HandlerFunc(callbackHandler.Post))
	defer ts.Close()

	var testObjectID uint = 1

	testerService.EXPECT().GetObject(ctx, testObjectID).Return(model.TesterObject{
		ID:     testObjectID,
		Online: false,
	}, nil)

	payload := fmt.Sprintf(`{"object_ids": [%d]}`, testObjectID)

	res, err := http.Post(ts.URL, "application/json", strings.NewReader(payload))
	if err != nil {
		t.Error(err)
	}
	defer res.Body.Close()

	content, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, content, []byte("ok"))
}

func TestPostObjectOnline(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	objectService := NewMockObjectService(ctrl)
	testerService := NewMockTesterService(ctrl)

	callbackHandler := New(objectService, testerService)

	ts := httptest.NewServer(http.HandlerFunc(callbackHandler.Post))
	defer ts.Close()

	var testObjectID uint = 1

	testerService.EXPECT().GetObject(ctx, testObjectID).Return(model.TesterObject{
		ID:     testObjectID,
		Online: true,
	}, nil)

	objectService.EXPECT().UpdateObject(ctx, &model.DBObject{
		ID: testObjectID,
	}).Return(nil)

	payload := `{"object_ids": [1]}`

	res, err := http.Post(ts.URL, "application/json", strings.NewReader(payload))
	if err != nil {
		t.Error(err)
	}
	defer res.Body.Close()

	content, err := io.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, content, []byte("ok"))
}
