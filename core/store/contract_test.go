package store_test

import (
	"testing"

	"gotenancy/core/store"
	"gotenancy/internal/testcontract"
)

func TestMemoryStoreContract(t *testing.T) {
	testcontract.RunStoreContract(t, func() store.Store {
		return store.NewMemoryStore()
	})
}
