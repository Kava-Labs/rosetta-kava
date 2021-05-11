package services

import (
	"context"
	"testing"

	"github.com/kava-labs/rosetta-kava/configuration"
	"github.com/kava-labs/rosetta-kava/kava"
	mocks "github.com/kava-labs/rosetta-kava/mocks/services"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupConstructionAPIServicer() *ConstructionAPIService {
	cfg := &configuration.Configuration{
		Mode: configuration.Offline,
	}
	mockClient := &mocks.Client{}
	return NewConstructionAPIService(cfg, mockClient)
}

func validConstructionPreprocessRequest() *types.ConstructionPreprocessRequest {
	defaultOps := []*types.Operation{
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 0},
			Type:                kava.TransferOpType,
			Account:             &types.AccountIdentifier{Address: "kava1esagqd83rhqdtpy5sxhklaxgn58k2m3s3mnpea"},
			Amount:              &types.Amount{Value: "-5000000", Currency: &types.Currency{Symbol: "KAVA", Decimals: 6}},
		},
		{
			OperationIdentifier: &types.OperationIdentifier{Index: 1},
			RelatedOperations:   []*types.OperationIdentifier{{Index: 0}},
			Type:                kava.TransferOpType,
			Account:             &types.AccountIdentifier{Address: "kava1mq9qxlhze029lm0frzw2xr6hem8c3k9ts54w0w"},
			Amount:              &types.Amount{Value: "5000000", Currency: &types.Currency{Symbol: "KAVA", Decimals: 6}},
		},
	}

	return &types.ConstructionPreprocessRequest{
		Operations: defaultOps,
		Metadata:   make(map[string]interface{}),
	}
}

func TestConstructionPreprocess_CurveValidation(t *testing.T) {
	servicer := setupConstructionAPIServicer()

	testCases := []struct {
		curveType types.CurveType
		isValid bool
	}{
		{
			curveType: "edwards25519",
			isValid: false,
		},
		{
			curveType: "secp256r1",
			isValid: false,
		},
		{
			curveType: "tweedle",
			isValid: false,
		},
		{
			curveType: "secp256k1",
			isValid: true,
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		request := validConstructionPreprocessRequest()

		request.CurveType = tc.curveType

		response, err := servicer.ConstructionPreprocess(ctx, request)
		require.Nil(t, err)

		actualCurveType, ok := response.isValid[true]
		require.True(t, ok)

		assert.Equal(t, tc.curveType, actualCurveType)
	}
}




