package services

import (
	"encoding/base64"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/rosetta-kava/configuration"
	mocks "github.com/kava-labs/rosetta-kava/mocks/services"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"testing"
)

func setupConstructionAPIServicer() *ConstructionAPIService {
	cfg := &configuration.Configuration{
		Mode: configuration.Offline,
	}
	mockClient := &mocks.Client{}
	cdc := app.MakeCodec()
	return NewConstructionAPIService(cfg, mockClient, cdc)
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

func TestConstructionDerive_PublicKeyEmptyNil(t *testing.T) {
	servicer := setupConstructionAPIServicer()

	testCases := []struct {
		name  string
		bytes []byte
	}{
		{
			name:  "Nil",
			bytes: nil,
		},
		{
			name:  "Zero length",
			bytes: []byte{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			request := &types.ConstructionDeriveRequest{
				PublicKey: &types.PublicKey{
					CurveType: types.Secp256k1,
					Bytes:     tc.bytes,
				},
			}

			response, err := servicer.ConstructionDerive(ctx, request)
			originalError := errors.New("nil public key")
			wrappedPublicKeyErr := wrapErr(ErrPublicKeyNil, originalError)

			assert.Nil(t, response)
			assert.Equal(t, wrappedPublicKeyErr, err)
		})
	}
}

func TestConstructionDerive_PublicKeyValid(t *testing.T) {
	servicer := setupConstructionAPIServicer()

	testCases := []struct {
		name    string
		key     string
		address string
	}{
		{
			name:    "Compressed Key 1",
			key:     "AsAbWjsqD1ntOiVZCNRdAm1nrSP8rwZoNNin85jPaeaY",
			address: "kava1vlpsrmdyuywvaqrv7rx6xga224sqfwz3fyfhwq",
		},
		{
			name:    "Uncompressed Key 1",
			key:     "BMAbWjsqD1ntOiVZCNRdAm1nrSP8rwZoNNin85jPaeaYvrG35oB42m6Hc60r5UqINTyW/8Z1kyZ5Ju9w4af71RI=",
			address: "kava1vlpsrmdyuywvaqrv7rx6xga224sqfwz3fyfhwq",
		},
		{
			name:    "Compressed Key 2",
			key:     "AwoUgfwik9NNmPhuFqVjRXG1GVEG7QjGAim/ADlZc7aS",
			address: "kava1xg0ktvzyqq7z6nx57e4yhfzsxxwh9nft5xyh8j",
		},
		{
			name:    "Uncompressed Key 2",
			key:     "BAoUgfwik9NNmPhuFqVjRXG1GVEG7QjGAim/ADlZc7aS3PKYgitX26InU5cIkzEFftPUeY1eMZ1CQKjgUUtEC+U=",
			address: "kava1xg0ktvzyqq7z6nx57e4yhfzsxxwh9nft5xyh8j",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			PubKeyBytes, error := base64.StdEncoding.DecodeString(tc.key)
			require.NoError(t, error)

			request := &types.ConstructionDeriveRequest{
				PublicKey: &types.PublicKey{
					CurveType: types.Secp256k1,
					Bytes:     PubKeyBytes,
				},
			}

			response, err := servicer.ConstructionDerive(ctx, request)
			assert.Equal(t, tc.address, response.AccountIdentifier.Address)
			assert.Nil(t, err)
		})
	}
}
