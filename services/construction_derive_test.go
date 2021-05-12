package services

import (
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/kava-labs/rosetta-kava/configuration"
	mocks "github.com/kava-labs/rosetta-kava/mocks/services"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"testing"
)

func setupConstructionAPIServicer() *ConstructionAPIService {
	cfg := &configuration.Configuration{
		Mode: configuration.Offline,
	}
	mockClient := &mocks.Client{}
	return NewConstructionAPIService(cfg, mockClient)
}

func TestConstructionDerive_CurveValidation(t *testing.T) {
	servicer := setupConstructionAPIServicer()
	Secp256r1 := types.CurveType("secp256r1")
	request := &types.ConstructionDeriveRequest{
		PublicKey: &types.PublicKey{
			CurveType: Secp256r1,
		},
	}
	ctx := context.Background()
	response, err := servicer.ConstructionDerive(ctx, request)

	assert.NotNil(t, response)
	assert.Equal(t, ErrUnsupportedCurveType.Code, err.Code)
}



