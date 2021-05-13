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

	testCases := []types.CurveType{
		types.Secp256r1,
		types.Edwards25519,
		types.Tweedle,
		types.Secp256k1,
	}

	for _, tc := range testCases {
		ctx := context.Background()
		request := &types.ConstructionDeriveRequest{
			PublicKey: &types.PublicKey{},
		}
		request.PublicKey.CurveType = tc
		response, err := servicer.ConstructionDerive(ctx, request)

		if tc == types.Secp256k1 {
			assert.Nil(t, response)
		} else {
			assert.Nil(t, response)
			assert.Equal(t, ErrUnsupportedCurveType, err)
		}
	}
}

func TestConstructionDerive_PublicKey(t *testing.T) {
	servicer := setupConstructionAPIServicer()

	request := &types.ConstructionDeriveRequest{
		PublicKey: &types.PublicKey{
			CurveType: types.Secp256k1,
			Bytes: nil,
		},
	}
	ctx := context.Background()
	response, err := servicer.ConstructionDerive(ctx, request)

	assert.Nil(t, response)
	assert.Equal(t, ErrPublicKeyNil, err)
}