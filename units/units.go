package units

import (
	"context"
	"fmt"
)

type Server struct {
	UnimplementedUnitsServer
}

// TODO add args once protobuffers are compiled and available here
func (s *Server) GetUnit(ctx context.Context, unitInfo *UnitInfo) (*Unit, error) {
	fmt.Printf("got a request to GetUnit with supplied '%v'", unitInfo)
	if *unitInfo.Type == UnitType_ABOLITIONIST {
		return &Abolitionist, nil
	}

	return nil, fmt.Errorf("the unit type given (%v) is not supported by the server", *unitInfo.Type)
}
