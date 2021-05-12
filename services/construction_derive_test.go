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
	secp256r1 := types.Secp256r1
	request := &types.ConstructionDeriveRequest{
		PublicKey: &types.PublicKey{
			CurveType: secp256r1,
		},
	}
	ctx := context.Background()
	response, err := servicer.ConstructionDerive(ctx, request)

	assert.Nil(t, response)
	assert.Equal(t, ErrUnsupportedCurveType.Code, err.Code)
}

func TestConstructionDerive_CurveValidation1(t *testing.T) {
	servicer := setupConstructionAPIServicer()
	tweedle := types.Tweedle
	request := &types.ConstructionDeriveRequest{
		PublicKey: &types.PublicKey{
			CurveType: tweedle,
		},
	}
	ctx := context.Background()
	response, err := servicer.ConstructionDerive(ctx, request)

	assert.Nil(t, response)
	assert.Equal(t, ErrUnsupportedCurveType.Code, err.Code)
}

func TestConstructionDerive_CurveValidation2(t *testing.T) {
	servicer := setupConstructionAPIServicer()
	edwards25519 := types.Edwards25519
	request := &types.ConstructionDeriveRequest{
		PublicKey: &types.PublicKey{
			CurveType: edwards25519,
		},
	}
	ctx := context.Background()
	response, err := servicer.ConstructionDerive(ctx, request)

	assert.Nil(t, response)
	assert.Equal(t, ErrUnsupportedCurveType.Code, err.Code)
}

func TestConstructionDerive_CurveValidation3(t *testing.T) {
	servicer := setupConstructionAPIServicer()
	secp256k1 := types.Secp256k1
	request := &types.ConstructionDeriveRequest{
		PublicKey: &types.PublicKey{
			CurveType: secp256k1,
		},
	}
	ctx := context.Background()
	response, err := servicer.ConstructionDerive(ctx, request)
	assert.Nil(t, response)
	assert.Nil(t, err)
}