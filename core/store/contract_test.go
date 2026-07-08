package store_test

import (
	"testing"

	"github.com/DarkInno/gotenancy/core/store"
	"github.com/DarkInno/gotenancy/internal/testcontract"
)

func TestMemoryStoreContract(t *testing.T) {
	testcontract.RunStoreContract(t, func() store.Store {
		return store.NewMemoryStore()
	})
}
