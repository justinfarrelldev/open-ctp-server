package ai

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

type typedProjection struct {
	model visibility.Projection
	err   error
}

func (p typedProjection) Projection(context.Context, int64, string) ([]byte, error) {
	return nil, errors.New("typed source must not serialize/reparse viewer models")
}
func (p typedProjection) ProjectionModel(context.Context, int64, string) (visibility.Projection, error) {
	return p.model, p.err
}

func TestTypedAIProjectionPreservesExactLegacySubmissionsAndCloseout(t *testing.T) {
	for _, status := range []string{"", visibility.PlanningStatusOpen, visibility.PlanningStatusQueued, visibility.PlanningStatusClosed} {
		var model visibility.Projection
		if err := json.Unmarshal(projectionJSON(t), &model); err != nil {
			t.Fatal(err)
		}
		model.PlanningStatus = status
		raw, err := json.Marshal(model)
		if err != nil {
			t.Fatal(err)
		}
		legacyGateway := &recordingGateway{revision: model.Revision}
		fastGateway := &recordingGateway{revision: model.Revision}
		config := Config{ActionBudget: 1, CloseoutAfterAction: true, History: staticHistory{}}
		legacy := NewDriver(staticProjection{bytes: raw}, legacyGateway, config)
		fast := NewDriver(typedProjection{model: model}, fastGateway, config)
		want, err := legacy.Drive(context.Background(), 44, "7", model.Revision)
		if err != nil {
			t.Fatal(err)
		}
		got, err := fast.Drive(context.Background(), 44, "7", model.Revision)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) || !reflect.DeepEqual(fastGateway.requests, legacyGateway.requests) {
			t.Fatalf("%s changed AI result/submissions: %+v %+v", status, fastGateway.requests, legacyGateway.requests)
		}
	}
}
func TestTypedAIProjectionErrorDoesNotFallBackOrSubmit(t *testing.T) {
	gateway := &recordingGateway{}
	sentinel := errors.New("projection unavailable")
	driver := NewDriver(typedProjection{err: sentinel}, gateway, Config{})
	if _, err := driver.Drive(context.Background(), 44, "7", 1); !errors.Is(err, sentinel) {
		t.Fatalf("projection error lost: %v", err)
	}
	if len(gateway.requests) != 0 {
		t.Fatal("failed projection produced a command")
	}
}
