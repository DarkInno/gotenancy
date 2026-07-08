package enttenant

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	entgo "entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"

	"github.com/DarkInno/gotenancy"
	tenantctx "github.com/DarkInno/gotenancy/core/context"
	"github.com/DarkInno/gotenancy/core/types"
	"github.com/DarkInno/gotenancy/data"
)

func TestPredicateAddsTenantFilter(t *testing.T) {
	ctx := tenantctx.WithTenant(context.Background(), types.Tenant{ID: "tenant-a"})
	selector := testSelector()

	predicate, err := Predicate(ctx, Config{})
	if err != nil {
		t.Fatalf("Predicate() error = %v", err)
	}
	predicate(selector)

	query, args := selector.Query()
	if !strings.Contains(query, "`orders`.`tenant_id` = ?") {
		t.Fatalf("query missing tenant predicate: %s", query)
	}
	if !reflect.DeepEqual(args, []any{"tenant-a"}) {
		t.Fatalf("args = %#v", args)
	}
}

func TestPredicateAddsSoftDeleteFilter(t *testing.T) {
	ctx := tenantctx.WithTenant(context.Background(), types.Tenant{ID: "tenant-a"})
	selector := testSelector()

	err := Apply(ctx, selector, Config{SoftDeleteField: "deleted_at"})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	query, args := selector.Query()
	if !strings.Contains(query, "`orders`.`tenant_id` = ?") {
		t.Fatalf("query missing tenant predicate: %s", query)
	}
	if !strings.Contains(query, "`orders`.`deleted_at` IS NULL") {
		t.Fatalf("query missing soft-delete predicate: %s", query)
	}
	if !reflect.DeepEqual(args, []any{"tenant-a"}) {
		t.Fatalf("args = %#v", args)
	}
}

func TestPredicateCanIncludeSoftDeleted(t *testing.T) {
	ctx := tenantctx.WithTenant(context.Background(), types.Tenant{ID: "tenant-a"})
	selector := testSelector()

	err := Apply(ctx, selector, Config{SoftDeleteField: "deleted_at", IncludeSoftDeleted: true})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	query, args := selector.Query()
	if strings.Contains(query, "deleted_at") {
		t.Fatalf("query should not include soft-delete predicate: %s", query)
	}
	if !reflect.DeepEqual(args, []any{"tenant-a"}) {
		t.Fatalf("args = %#v", args)
	}
}

func TestPredicateHostContextIsNoop(t *testing.T) {
	ctx := tenantctx.WithHost(context.Background())
	selector := testSelector()

	err := Apply(ctx, selector, Config{})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	query, args := selector.Query()
	if strings.Contains(query, "WHERE") {
		t.Fatalf("host query should not be filtered: %s", query)
	}
	if len(args) != 0 {
		t.Fatalf("args = %#v", args)
	}
}

func TestPredicateRequiresTenantContext(t *testing.T) {
	_, err := Predicate(context.Background(), Config{})
	if !errors.Is(err, data.ErrNoTenant) {
		t.Fatalf("Predicate() error = %v, want %v", err, data.ErrNoTenant)
	}
}

func TestPredicateValidatesFieldNames(t *testing.T) {
	ctx := tenantctx.WithTenant(context.Background(), types.Tenant{ID: "tenant-a"})

	_, err := Predicate(ctx, Config{TenantField: "tenant_id;drop"})
	if !errors.Is(err, data.ErrInvalidFieldName) {
		t.Fatalf("Predicate() error = %v, want %v", err, data.ErrInvalidFieldName)
	}
}

func TestFilterQueryAddsTenantPredicate(t *testing.T) {
	ctx := tenantctx.WithTenant(context.Background(), types.Tenant{ID: "tenant-a"})
	query := &fakeQuery{}

	err := FilterQuery(ctx, query, Config{})
	if err != nil {
		t.Fatalf("FilterQuery() error = %v", err)
	}

	sqlText, args := query.sql()
	if !strings.Contains(sqlText, "`orders`.`tenant_id` = ?") {
		t.Fatalf("query missing tenant predicate: %s", sqlText)
	}
	if !reflect.DeepEqual(args, []any{"tenant-a"}) {
		t.Fatalf("args = %#v", args)
	}
}

func TestFilterQueryHostContextIsNoop(t *testing.T) {
	query := &fakeQuery{}

	err := FilterQuery(tenantctx.WithHost(context.Background()), query, Config{})
	if err != nil {
		t.Fatalf("FilterQuery(host) error = %v", err)
	}

	sqlText, args := query.sql()
	if strings.Contains(sqlText, "WHERE") {
		t.Fatalf("host query should not be filtered: %s", sqlText)
	}
	if len(args) != 0 {
		t.Fatalf("args = %#v", args)
	}
}

func TestFilterMutationCreateSetsTenantField(t *testing.T) {
	ctx := tenantctx.WithTenant(context.Background(), types.Tenant{ID: "tenant-a"})
	mutation := newFakeMutation(entgo.OpCreate)

	err := FilterMutation(ctx, mutation, Config{})
	if err != nil {
		t.Fatalf("FilterMutation(create) error = %v", err)
	}

	value, ok := mutation.Field("tenant_id")
	if !ok || value != "tenant-a" {
		t.Fatalf("tenant_id = %#v, %v", value, ok)
	}
	if len(mutation.predicates) != 0 {
		t.Fatalf("create mutation should not add predicates")
	}
}

func TestFilterMutationCreateRejectsTenantMismatch(t *testing.T) {
	ctx := tenantctx.WithTenant(context.Background(), types.Tenant{ID: "tenant-a"})
	mutation := newFakeMutation(entgo.OpCreate)
	mutation.fields["tenant_id"] = "tenant-b"

	err := FilterMutation(ctx, mutation, Config{})
	if !errors.Is(err, gotenancy.ErrTenantMismatch) {
		t.Fatalf("FilterMutation(create mismatch) error = %v, want %v", err, gotenancy.ErrTenantMismatch)
	}
}

func TestFilterMutationCreateReportsMissingTenantField(t *testing.T) {
	ctx := tenantctx.WithTenant(context.Background(), types.Tenant{ID: "tenant-a"})
	mutation := newFakeMutation(entgo.OpCreate)
	mutation.setErr = errors.New("unknown field")

	err := FilterMutation(ctx, mutation, Config{})
	if !errors.Is(err, ErrTenantFieldNotFound) {
		t.Fatalf("FilterMutation(missing field) error = %v, want %v", err, ErrTenantFieldNotFound)
	}
}

func TestFilterMutationUpdateAddsTenantPredicate(t *testing.T) {
	ctx := tenantctx.WithTenant(context.Background(), types.Tenant{ID: "tenant-a"})
	mutation := newFakeMutation(entgo.OpUpdate)

	err := FilterMutation(ctx, mutation, Config{})
	if err != nil {
		t.Fatalf("FilterMutation(update) error = %v", err)
	}

	sqlText, args := mutation.sql()
	if !strings.Contains(sqlText, "`orders`.`tenant_id` = ?") {
		t.Fatalf("mutation missing tenant predicate: %s", sqlText)
	}
	if !reflect.DeepEqual(args, []any{"tenant-a"}) {
		t.Fatalf("args = %#v", args)
	}
}

func TestFilterMutationDeleteAddsSoftDeletePredicate(t *testing.T) {
	ctx := tenantctx.WithTenant(context.Background(), types.Tenant{ID: "tenant-a"})
	mutation := newFakeMutation(entgo.OpDelete)

	err := FilterMutation(ctx, mutation, Config{SoftDeleteField: "deleted_at"})
	if err != nil {
		t.Fatalf("FilterMutation(delete) error = %v", err)
	}

	sqlText, args := mutation.sql()
	if !strings.Contains(sqlText, "`orders`.`tenant_id` = ?") {
		t.Fatalf("mutation missing tenant predicate: %s", sqlText)
	}
	if !strings.Contains(sqlText, "`orders`.`deleted_at` IS NULL") {
		t.Fatalf("mutation missing soft-delete predicate: %s", sqlText)
	}
	if !reflect.DeepEqual(args, []any{"tenant-a"}) {
		t.Fatalf("args = %#v", args)
	}
}

func TestFilterMutationRequiresTenantContext(t *testing.T) {
	err := FilterMutation(context.Background(), newFakeMutation(entgo.OpUpdate), Config{})
	if !errors.Is(err, data.ErrNoTenant) {
		t.Fatalf("FilterMutation(no tenant) error = %v, want %v", err, data.ErrNoTenant)
	}
}

func TestFilterMutationHostContextIsNoop(t *testing.T) {
	mutation := newFakeMutation(entgo.OpUpdate)

	err := FilterMutation(tenantctx.WithHost(context.Background()), mutation, Config{})
	if err != nil {
		t.Fatalf("FilterMutation(host) error = %v", err)
	}
	if len(mutation.predicates) != 0 {
		t.Fatalf("host mutation should not be filtered")
	}
}

func TestHookAppliesMutationFilter(t *testing.T) {
	ctx := tenantctx.WithTenant(context.Background(), types.Tenant{ID: "tenant-a"})
	mutation := &fakeEntMutation{fakeMutation: newFakeMutation(entgo.OpUpdate)}
	called := false

	mutator := Hook(Config{})(entgo.MutateFunc(func(context.Context, entgo.Mutation) (entgo.Value, error) {
		called = true
		return "ok", nil
	}))

	value, err := mutator.Mutate(ctx, mutation)
	if err != nil {
		t.Fatalf("Mutate() error = %v", err)
	}
	if value != "ok" {
		t.Fatalf("value = %#v", value)
	}
	if !called {
		t.Fatalf("next mutator was not called")
	}
	if len(mutation.predicates) != 1 {
		t.Fatalf("predicates = %d, want 1", len(mutation.predicates))
	}
}

func testSelector() *sql.Selector {
	return sql.Dialect(dialect.MySQL).
		Select("*").
		From(sql.Table("orders"))
}

type fakeQuery struct {
	predicates []SelectorPredicate
}

func (query *fakeQuery) WhereP(predicates ...SelectorPredicate) {
	query.predicates = append(query.predicates, predicates...)
}

func (query *fakeQuery) sql() (string, []any) {
	selector := testSelector()
	for _, predicate := range query.predicates {
		predicate(selector)
	}
	return selector.Query()
}

type fakeMutation struct {
	op         entgo.Op
	fields     map[string]entgo.Value
	setErr     error
	predicates []SelectorPredicate
}

func newFakeMutation(op entgo.Op) *fakeMutation {
	return &fakeMutation{
		op:     op,
		fields: make(map[string]entgo.Value),
	}
}

func (mutation *fakeMutation) Op() entgo.Op {
	return mutation.op
}

func (mutation *fakeMutation) WhereP(predicates ...SelectorPredicate) {
	mutation.predicates = append(mutation.predicates, predicates...)
}

func (mutation *fakeMutation) Field(name string) (entgo.Value, bool) {
	value, ok := mutation.fields[name]
	return value, ok
}

func (mutation *fakeMutation) SetField(name string, value entgo.Value) error {
	if mutation.setErr != nil {
		return mutation.setErr
	}
	mutation.fields[name] = value
	return nil
}

func (mutation *fakeMutation) sql() (string, []any) {
	selector := testSelector()
	for _, predicate := range mutation.predicates {
		predicate(selector)
	}
	return selector.Query()
}

type fakeEntMutation struct {
	*fakeMutation
}

func (mutation *fakeEntMutation) Type() string {
	return "Order"
}

func (mutation *fakeEntMutation) Fields() []string {
	return nil
}

func (mutation *fakeEntMutation) AddedFields() []string {
	return nil
}

func (mutation *fakeEntMutation) AddedField(string) (entgo.Value, bool) {
	return nil, false
}

func (mutation *fakeEntMutation) AddField(string, entgo.Value) error {
	return nil
}

func (mutation *fakeEntMutation) ClearedFields() []string {
	return nil
}

func (mutation *fakeEntMutation) FieldCleared(string) bool {
	return false
}

func (mutation *fakeEntMutation) ClearField(string) error {
	return nil
}

func (mutation *fakeEntMutation) ResetField(string) error {
	return nil
}

func (mutation *fakeEntMutation) AddedEdges() []string {
	return nil
}

func (mutation *fakeEntMutation) AddedIDs(string) []entgo.Value {
	return nil
}

func (mutation *fakeEntMutation) RemovedEdges() []string {
	return nil
}

func (mutation *fakeEntMutation) RemovedIDs(string) []entgo.Value {
	return nil
}

func (mutation *fakeEntMutation) ClearedEdges() []string {
	return nil
}

func (mutation *fakeEntMutation) EdgeCleared(string) bool {
	return false
}

func (mutation *fakeEntMutation) ClearEdge(string) error {
	return nil
}

func (mutation *fakeEntMutation) ResetEdge(string) error {
	return nil
}

func (mutation *fakeEntMutation) OldField(context.Context, string) (entgo.Value, error) {
	return nil, errors.New("old field is not implemented")
}
