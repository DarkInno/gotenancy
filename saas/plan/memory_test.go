package plan

import (
	"context"
	"errors"
	"testing"
)

func TestMemoryServiceCRUD(t *testing.T) {
	ctx := context.Background()
	service := NewMemoryService()
	plan := testPlan("starter")

	if err := service.Create(ctx, plan); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := service.Create(ctx, plan); !errors.Is(err, ErrPlanAlreadyExists) {
		t.Fatalf("Create(duplicate) error = %v, want ErrPlanAlreadyExists", err)
	}

	got, err := service.Get(ctx, plan.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.ID != plan.ID || got.Name != plan.Name {
		t.Fatalf("Get() = %+v, want plan", got)
	}

	plan.Name = "Starter Updated"
	plan.Quotas[0].Limit = 200
	if err := service.Update(ctx, plan); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	got, err = service.Get(ctx, plan.ID)
	if err != nil {
		t.Fatalf("Get() after update error = %v", err)
	}
	if got.Name != "Starter Updated" || got.Quotas[0].Limit != 200 {
		t.Fatalf("Get() after update = %+v, want updated", got)
	}

	if err := service.Delete(ctx, plan.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := service.Get(ctx, plan.ID); !errors.Is(err, ErrPlanNotFound) {
		t.Fatalf("Get(deleted) error = %v, want ErrPlanNotFound", err)
	}
}

func TestMemoryServiceCopiesPlan(t *testing.T) {
	ctx := context.Background()
	service := NewMemoryService()
	plan := testPlan("starter")

	if err := service.Create(ctx, plan); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	plan.Features[0].Config["seats"] = "999"

	got, err := service.Get(ctx, "starter")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Features[0].Config["seats"] != "5" {
		t.Fatalf("stored feature config = %q, want 5", got.Features[0].Config["seats"])
	}

	got.Features[0].Config["seats"] = "1"
	again, err := service.Get(ctx, "starter")
	if err != nil {
		t.Fatalf("Get() again error = %v", err)
	}
	if again.Features[0].Config["seats"] != "5" {
		t.Fatalf("returned feature config mutated store, got %q", again.Features[0].Config["seats"])
	}
}

func TestMemoryServiceValidation(t *testing.T) {
	ctx := context.Background()
	service := NewMemoryService()

	tests := []Plan{
		{},
		{ID: "starter"},
		{ID: "starter", Name: "Starter", Features: []Feature{{}}},
		{ID: "starter", Name: "Starter", Quotas: []Quota{{Resource: "api", Limit: -1, Period: QuotaPeriodDay}}},
		{ID: "starter", Name: "Starter", Quotas: []Quota{{Resource: "api", Limit: 1}}},
	}

	for i, plan := range tests {
		if err := service.Create(ctx, plan); !errors.Is(err, ErrInvalidPlan) {
			t.Fatalf("Create(invalid %d) error = %v, want ErrInvalidPlan", i, err)
		}
	}
}

func testPlan(id string) Plan {
	return Plan{
		ID:   id,
		Name: "Plan " + id,
		Features: []Feature{{
			Key:     "members",
			Enabled: true,
			Config:  map[string]string{"seats": "5"},
		}},
		Quotas: []Quota{{
			Resource: "api_calls",
			Limit:    100,
			Period:   QuotaPeriodDay,
		}},
	}
}
