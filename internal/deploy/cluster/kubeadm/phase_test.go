package kubeadm

import (
	"context"
	"strings"
	"testing"
)

/*
测试用例设计
需要覆盖 5 个场景：

测试函数	场景	验证点
TestRegistryRegisterAndGet	注册 + 查询	Register 后 Get 能拿到；Get 不存在的返回 nil
TestRegistrySorted	拓扑排序正确性	3 个阶段 A→B→C，排序后顺序为 A,B,C
TestRegistrySortedCircular	循环依赖检测	A→B→C→A 应返回 error
TestRegistrySortedUnknownDep	依赖不存在的阶段	A 依赖 X（不存在）应返回 error
TestRegistrySortedCacheInvalidation	缓存失效	Register 后 Sorted 应重新计算
TestRunAll	全量执行	所有阶段按序执行，记录执行顺序
TestRunOne	单阶段 + 依赖	RunOne("C") 应执行 A,B,C
TestRunAllDryRun	DryRun 模式	不应调用 ExecuteFunc

*/

//mackPhase 用于测试的Phase 实现
type mockPhase struct {
	name         string
	description  string
	dependencies []string
}

func (m *mockPhase) Name() string                     { return m.name }
func (m *mockPhase) Description() string              { return m.description }
func (m *mockPhase) Dependencies() []string           { return m.dependencies }
func (m *mockPhase) Command(cfg PhaseConfig) []string { return nil }

// newMockPhase 创建一个mockPhase 实例
func newMockPhase(name string, deps ...string) *mockPhase {
	return &mockPhase{
		name:         name,
		description:  name + "-desc",
		dependencies: deps,
	}
}

// Test 1: 注册和查询
func TestRegistryRegistryAndGet(t *testing.T) {
	r := NewPhaseRegistry()
	p := newMockPhase("preflight")
	r.Register(p)
	got := r.Get("preflight")
	if got == nil {
		t.Fatal("expected phase 'preflight', got nil to exist")
	}
	if got.Name() != "preflight" {
		t.Errorf("expected preflight, got %s", got.Name())
	}
	if r.Get("nonexistent") != nil {
		t.Errorf("expected nonexistent, got %s", r.Get("nonexistent"))
	}
}

// Test 2: 拓扑排序
/*

这里 A→B, B→C，正确的拓扑顺序是 C, B, A（无依赖的先执行）或 A, B, C。取决于依赖链的方向。

在 kubeadm 的设计里，Dependencies() 返回的是当前阶段依赖的前置阶段。所以：

A 无依赖 → A 入度 0 → 先排序
B 依赖 A → B 入度 1 → A 执行后 B 入度变 0
C 依赖 B → C 入度 1 → B 执行后 C 入度变 0
排序结果应为：A → B → C

*/
func TestRegistrySorted(t *testing.T) {
	r := NewPhaseRegistry()
	
	// A 无依赖,B 依赖 A,C 依赖B
	r.Register(newMockPhase("a"))
	r.Register(newMockPhase("b", "a"))
	r.Register(newMockPhase("c", "b"))
	
	//sorted, err := r.Sorted()
	sorted, err := r.Sorted()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sorted) != 3 {
		t.Fatalf("expected 3 phases, got: %d", len(sorted))
	}
	//验证顺序
	names := make([]string, len(sorted))
	for i, p := range sorted {
		names[i] = p.Name()
	}
	expected := []string{"a", "b", "c"}
	for i := range names {
		if names[i] != expected[i] {
			t.Errorf("index %d expected %q, got %q", i, expected[i], names[i])
			t.Logf("full order: %v", names)
		}
	}
}

// Test 3: 循环依赖检测
func TestRegistrySortedCircular(t *testing.T) {
	r := NewPhaseRegistry()
	// a -> b -> c -> a形成循环
	r.Register(newMockPhase("a", "c"))
	r.Register(newMockPhase("b", "a"))
	r.Register(newMockPhase("c", "b"))
	
	_, err := r.Sorted()
	if err == nil {
		t.Fatal("expected cicrcular dependency error")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Fatalf("expected 'circular dependency' in error message, got: %s", err.Error())
	}
}

// Test 4: 依赖不存在的阶段
func TestRegistrySortedUnknown(t *testing.T) {
	r := NewPhaseRegistry()
	r.Register(newMockPhase("a", "nonexistent"))
	_, err := r.Sorted()
	if err == nil {
		t.Fatal("expected error for unknown dependency")
	}
	if !strings.Contains(err.Error(), "unknown phase") {
		t.Fatalf("expected 'unknown phase' in error got %s", err.Error())
	}
}

// Test 5: 缓存失效
func TestRegistrySortedCacheInvalidation(t *testing.T) {
	r := NewPhaseRegistry()
	r.Register(newMockPhase("a"))
	// -- 建立缓存
	_, _ = r.Sorted()
	
	// 新增阶段后缓存应失效
	r.Register(newMockPhase("b", "a"))
	sorted, err := r.Sorted()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sorted) != 2 {
		t.Errorf("expected 2 phases after re-sort, got: %d", len(sorted))
	}
}

// Test 6: 全量执行
func TestRunAll(t *testing.T) {
	registry := NewPhaseRegistry()
	registry.Register(newMockPhase("a"))
	registry.Register(newMockPhase("b", "a"))
	registry.Register(newMockPhase("c", "a"))
	
	var executed []string
	exec := func(ctx context.Context, phase Phase, cfg PhaseConfig) error {
		executed = append(executed, phase.Name())
		return nil
	}
	
	ctx := context.Background()
	cfg := PhaseConfig{}
	if err := registry.RunAll(ctx, cfg, exec); err != nil {
		t.Fatalf("RunAll failed: %v", err)
	}
	if len(executed) != 3 {
		t.Errorf("expected 3 phases, got: %d: %v", len(executed), executed)
	}
	if executed[0] != "a" || executed[1] != "b" || executed[2] != "c" {
		t.Errorf("wrong order: got %v", executed)
	}
}

// Test 7: RunOne 模式 (单阶段+依赖)
func TestRun(t *testing.T) {
	registry := NewPhaseRegistry()
	registry.Register(newMockPhase("a"))
	registry.Register(newMockPhase("b", "a"))
	registry.Register(newMockPhase("c", "b"))
	var executed []string
	exec := func(ctx context.Context, phase Phase, cfg PhaseConfig) error {
		executed = append(executed, phase.Name())
		return nil
	}
	// 只要求执行c ,应自定执行 a -> b -> c
	if err := registry.RunOne(context.Background(), "c", PhaseConfig{}, exec); err != nil {
		t.Fatalf("RunOne failed: %v", err)
	}
	if len(executed) != 3 || executed[0] != "a" || executed[1] != "b" || executed[2] != "c" {
		t.Errorf("wrong order: got %v", executed)
	}
}

// Test 8: DryRun不执行任何操作
func TestRunAllDryRun(t *testing.T) {
	registry := NewPhaseRegistry()
	registry.Register(newMockPhase("a"))
	registry.Register(newMockPhase("b", "a"))
	
	called := false
	exec := func(ctx context.Context, phase Phase, cfg PhaseConfig) error {
		called = true
		return nil
	}
	cfg := PhaseConfig{DryRun: true}
	if err := registry.RunAll(context.Background(), cfg, exec); err != nil {
		t.Fatalf("RunAll dry-run failed: %v", err)
	}
	if called {
		t.Fatalf("dry-run was called unexpectedly")
	}
}
