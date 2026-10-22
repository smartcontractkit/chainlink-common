package di

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/samber/do/v2"
)

// Fixtures: a chain of four, so an override of the one at the bottom has three levels of
// dependents above it to reach through.
type Foo struct {
	I int
}

type Bar struct {
	S string
}

type Zaz struct {
	B *Bar
}

// Deep never mentions Bar, so it proves an override reaches arbitrarily far down
// the graph.
type Deep struct {
	Z *Zaz
}

// Late is registered on the root only after the children already exist.
type Late struct {
	D *Deep
}

func ProvideZaz(r Resolver) (*Zaz, error) {
	bar, err := r.Invoke[*Bar]()
	if err != nil {
		return nil, err
	}
	return &Zaz{B: bar}, nil
}

func ProvideDeep(r Resolver) (*Deep, error) {
	zaz, err := r.Invoke[*Zaz]()
	if err != nil {
		return nil, err
	}
	return &Deep{Z: zaz}, nil
}

func ProvideLate(r Resolver) (*Late, error) {
	deep, err := r.Invoke[*Deep]()
	if err != nil {
		return nil, err
	}
	return &Late{D: deep}, nil
}

func newRoot(t *testing.T) *Container {
	t.Helper()
	c := New()
	c.ProvideValue(&Foo{I: 10})
	c.ProvideValue(&Bar{S: "root"})
	c.Provide(ProvideZaz)
	c.Provide(ProvideDeep)
	c.Provide(ProvideLate)
	return c
}

func mustBar(t *testing.T, c *Container) string {
	t.Helper()
	bar, err := c.Invoke[*Bar]()
	if err != nil {
		t.Fatalf("invoke *Bar: %v", err)
	}
	return bar.S
}

// deepestBar resolves Bar through the whole four-level chain.
func deepestBar(t *testing.T, c *Container) string {
	t.Helper()
	late, err := c.Invoke[*Late]()
	if err != nil {
		t.Fatalf("invoke *Late: %v", err)
	}
	return late.D.Z.B.S
}

func TestRootResolvesItsOwnGraph(t *testing.T) {
	root := newRoot(t)
	if got := mustBar(t, root); got != "root" {
		t.Errorf("Bar = %q, want %q", got, "root")
	}
	if got := deepestBar(t, root); got != "root" {
		t.Errorf("Late.D.Z.B = %q, want %q", got, "root")
	}
}

func TestDeriveOverrideResolvesThroughDeepChain(t *testing.T) {
	root := newRoot(t)
	child := root.Derive()
	child.Override(func(Resolver) (*Bar, error) { return &Bar{S: "child"}, nil })

	if got := deepestBar(t, child); got != "child" {
		t.Errorf("child Late.D.Z.B = %q, want %q", got, "child")
	}
	if got := deepestBar(t, root); got != "root" {
		t.Errorf("root Late.D.Z.B = %q, want %q (parent must be untouched)", got, "root")
	}
}

// The reason Clone alone was not enough: registrations made after the child
// exists still have to reach it.
func TestLateParentRegistrationCascadesToExistingChild(t *testing.T) {
	root := New()
	root.ProvideValue(&Bar{S: "root"})
	root.Provide(ProvideZaz)

	child := root.Derive()
	child.OverrideValue(&Bar{S: "child"})

	// Registered only now, after the child already exists and has overridden Bar.
	root.Provide(ProvideDeep)
	root.Provide(ProvideLate)

	if got := deepestBar(t, child); got != "child" {
		t.Errorf("child Late.D.Z.B = %q, want %q", got, "child")
	}
	if got := deepestBar(t, root); got != "root" {
		t.Errorf("root Late.D.Z.B = %q, want %q", got, "root")
	}
}

func TestChildRegistrationNotClobberedByLaterParentProvide(t *testing.T) {
	root := New()
	root.ProvideValue(&Bar{S: "root"})
	root.Provide(ProvideZaz)

	child := root.Derive()
	// Child pins a *Deep the root does not have yet.
	child.ProvideValue(&Deep{Z: &Zaz{B: &Bar{S: "pinned"}}})

	root.Provide(ProvideDeep)
	root.Provide(ProvideLate)

	if got := deepestBar(t, child); got != "pinned" {
		t.Errorf("child Late.D.Z.B = %q, want %q", got, "pinned")
	}
	if got := deepestBar(t, root); got != "root" {
		t.Errorf("root Late.D.Z.B = %q, want %q", got, "root")
	}
}

func TestGrandchildOverrideWinsOverParentAndRoot(t *testing.T) {
	root := New()
	root.ProvideValue(&Bar{S: "root"})
	root.Provide(ProvideZaz)

	child := root.Derive()
	child.OverrideValue(&Bar{S: "child"})

	grandchild := child.Derive()
	grandchild.OverrideValue(&Bar{S: "grandchild"})

	// Registered on the root after all three levels exist.
	root.Provide(ProvideDeep)
	root.Provide(ProvideLate)

	for _, tc := range []struct {
		name string
		c    *Container
		want string
	}{
		{"root", root, "root"},
		{"child", child, "child"},
		{"grandchild", grandchild, "grandchild"},
	} {
		if got := deepestBar(t, tc.c); got != tc.want {
			t.Errorf("%s Late.D.Z.B = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestSiblingsAreIsolated(t *testing.T) {
	root := newRoot(t)
	a := root.Derive()
	a.OverrideValue(&Bar{S: "a"})
	b := root.Derive()
	b.OverrideValue(&Bar{S: "b"})

	if got := deepestBar(t, a); got != "a" {
		t.Errorf("a = %q, want %q", got, "a")
	}
	if got := deepestBar(t, b); got != "b" {
		t.Errorf("b = %q, want %q", got, "b")
	}
	if got := deepestBar(t, root); got != "root" {
		t.Errorf("root = %q, want %q", got, "root")
	}
}

// A child may Provide a type it only inherited, without the caller having to
// know the parent registered it.
func TestProvideShadowsInheritedTypeWithoutPanicking(t *testing.T) {
	root := New()
	root.ProvideValue(&Bar{S: "root"})
	root.Provide(ProvideZaz)

	child := root.Derive()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Provide of an inherited type panicked: %v", r)
		}
	}()
	child.ProvideValue(&Bar{S: "child"})

	if got := mustBar(t, child); got != "child" {
		t.Errorf("child Bar = %q, want %q", got, "child")
	}
	if got := mustBar(t, root); got != "root" {
		t.Errorf("root Bar = %q, want %q", got, "root")
	}
}

func TestProvideTwiceOnSameContainerPanics(t *testing.T) {
	c := New()
	c.ProvideValue(&Bar{S: "first"})

	defer func() {
		if recover() == nil {
			t.Fatal("expected a panic on duplicate Provide")
		}
	}()
	c.ProvideValue(&Bar{S: "second"})
}

func TestOverrideIsIdempotentAndNeverPanics(t *testing.T) {
	c := New()
	// Overriding a type that was never registered must work.
	c.OverrideValue(&Bar{S: "first"})
	c.OverrideValue(&Bar{S: "second"})
	c.OverrideValue(&Bar{S: "third"})

	if got := mustBar(t, c); got != "third" {
		t.Errorf("Bar = %q, want %q", got, "third")
	}
}

func TestOverrideDoesNotLeakToParent(t *testing.T) {
	root := newRoot(t)
	child := root.Derive()
	child.OverrideValue(&Bar{S: "child"})

	if got := mustBar(t, root); got != "root" {
		t.Errorf("root Bar = %q, want %q", got, "root")
	}
}

// A type introduced by Override on a parent still fills holes in children that
// predate it.
func TestOverrideOfNewTypeCascadesToExistingChild(t *testing.T) {
	root := New()
	root.ProvideValue(&Bar{S: "root"})
	child := root.Derive()

	root.OverrideValue(&Foo{I: 42})

	foo, err := child.Invoke[*Foo]()
	if err != nil {
		t.Fatalf("child invoke *Foo: %v", err)
	}
	if foo.I != 42 {
		t.Errorf("child Foo.I = %d, want 42", foo.I)
	}
}

func TestProvideValueSharesInstanceAcrossContainers(t *testing.T) {
	root := New()
	root.ProvideValue(&Foo{I: 10})
	child := root.Derive()
	grandchild := child.Derive()

	rootFoo := root.MustInvoke[*Foo]()
	if rootFoo != child.MustInvoke[*Foo]() {
		t.Error("child Foo is not the same instance as root Foo")
	}
	if rootFoo != grandchild.MustInvoke[*Foo]() {
		t.Error("grandchild Foo is not the same instance as root Foo")
	}
}

// Lazily provided services are per container, which is exactly what makes a
// child's override reach the whole subtree.
func TestProvidedServicesAreDistinctPerContainer(t *testing.T) {
	root := New()
	root.ProvideValue(&Bar{S: "root"})
	root.Provide(ProvideZaz)

	child := root.Derive()
	child.OverrideValue(&Bar{S: "child"})

	if root.MustInvoke[*Zaz]() == child.MustInvoke[*Zaz]() {
		t.Error("root and child share a *Zaz instance; the override could not have applied")
	}
}

func TestInvokeReportsMissingService(t *testing.T) {
	c := New()
	if _, err := c.Invoke[*Bar](); err == nil {
		t.Fatal("expected an error invoking an unregistered type")
	}
}

func TestProviderErrorPropagates(t *testing.T) {
	sentinel := errors.New("boom")
	c := New()
	c.Provide(func(Resolver) (*Bar, error) { return nil, sentinel })

	_, err := c.Invoke[*Bar]()
	if !errors.Is(err, sentinel) {
		t.Fatalf("err = %v, want it to wrap %v", err, sentinel)
	}
}

// A provider whose own dependency is missing must surface an error rather than
// resolve to something from another container.
func TestProviderMissingDependencyErrors(t *testing.T) {
	c := New()
	c.Provide(ProvideZaz) // *Bar never registered

	if _, err := c.Invoke[*Zaz](); err == nil {
		t.Fatal("expected an error when a provider's dependency is missing")
	}
}

// Resolver wraps the scope samber/do is building in, so cycles are still
// reported rather than overflowing the stack.
func TestCircularDependencyIsDetected(t *testing.T) {
	c := New()
	c.Provide(func(r Resolver) (*Zaz, error) {
		deep, err := r.Invoke[*Deep]()
		if err != nil {
			return nil, err
		}
		return &Zaz{B: deep.Z.B}, nil
	})
	c.Provide(ProvideDeep) // Deep -> Zaz -> Deep

	_, err := c.Invoke[*Zaz]()
	if !errors.Is(err, do.ErrCircularDependency) {
		t.Fatalf("err = %v, want it to wrap ErrCircularDependency", err)
	}
}

// Documents the one ordering rule callers must respect.
func TestOverrideAfterResolutionLeavesDependentsStale(t *testing.T) {
	root := New()
	root.ProvideValue(&Bar{S: "root"})
	root.Provide(ProvideZaz)

	child := root.Derive()
	if got := child.MustInvoke[*Zaz]().B.S; got != "root" {
		t.Fatalf("precondition: child Zaz.B = %q, want %q", got, "root")
	}

	child.OverrideValue(&Bar{S: "child"})

	if got := mustBar(t, child); got != "child" {
		t.Errorf("child Bar = %q, want %q", got, "child")
	}
	if got := child.MustInvoke[*Zaz]().B.S; got != "root" {
		t.Errorf("child Zaz.B = %q, want %q: an already built dependent keeps the old value", got, "root")
	}
}

func TestDeriveSnapshotsCurrentStateForNewChildren(t *testing.T) {
	root := New()
	root.ProvideValue(&Bar{S: "root"})

	early := root.Derive()
	early.OverrideValue(&Bar{S: "early"})

	root.Provide(ProvideZaz)

	// Derived after Zaz existed, so it gets Zaz from the clone rather than a cascade.
	late := root.Derive()
	late.OverrideValue(&Bar{S: "late"})

	if got := early.MustInvoke[*Zaz]().B.S; got != "early" {
		t.Errorf("early Zaz.B = %q, want %q", got, "early")
	}
	if got := late.MustInvoke[*Zaz]().B.S; got != "late" {
		t.Errorf("late Zaz.B = %q, want %q", got, "late")
	}
}

func TestDeepNestingFiveLevels(t *testing.T) {
	root := New()
	root.ProvideValue(&Bar{S: "root"})
	root.Provide(ProvideZaz)
	root.Provide(ProvideDeep)
	root.Provide(ProvideLate)

	current := root
	for i := range 5 {
		current = current.Derive()
		want := fmt.Sprintf("level-%d", i)
		current.OverrideValue(&Bar{S: want})

		if got := deepestBar(t, current); got != want {
			t.Fatalf("level %d Late.D.Z.B = %q, want %q", i, got, want)
		}
	}
	if got := deepestBar(t, root); got != "root" {
		t.Errorf("root Late.D.Z.B = %q, want %q", got, "root")
	}
}

type closer struct {
	name   string
	closed *[]string
	mu     *sync.Mutex
}

func (c *closer) Shutdown() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	*c.closed = append(*c.closed, c.name)
	return nil
}

func TestShutdownClosesChildrenAndSelf(t *testing.T) {
	var mu sync.Mutex
	var closed []string

	root := New()
	root.ProvideValue(&closer{name: "root", closed: &closed, mu: &mu})
	child := root.Derive()
	child.OverrideValue(&closer{name: "child", closed: &closed, mu: &mu})

	// Force both to be instantiated so samber has something to shut down.
	root.MustInvoke[*closer]()
	child.MustInvoke[*closer]()

	if err := root.Shutdown(); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(closed) != 2 {
		t.Fatalf("closed = %v, want both root and child", closed)
	}
	if closed[0] != "child" {
		t.Errorf("closed = %v, want the child shut down first", closed)
	}
}

func TestConcurrentDeriveProvideAndInvoke(t *testing.T) {
	root := New()
	root.ProvideValue(&Bar{S: "root"})
	root.Provide(ProvideZaz)

	children := make([]*Container, 16)
	for i := range children {
		children[i] = root.Derive()
		children[i].OverrideValue(&Bar{S: fmt.Sprintf("child-%d", i)})
	}

	var wg sync.WaitGroup

	// Registrations on the root cascading while children resolve concurrently.
	wg.Add(1)
	go func() {
		defer wg.Done()
		root.Provide(ProvideDeep)
	}()

	for i, child := range children {
		wg.Add(1)
		go func(i int, child *Container) {
			defer wg.Done()
			want := fmt.Sprintf("child-%d", i)
			if got := child.MustInvoke[*Zaz]().B.S; got != want {
				t.Errorf("child %d Zaz.B = %q, want %q", i, got, want)
			}
			_ = child.Derive()
		}(i, child)
	}
	wg.Wait()

	// Deep landed in every child regardless of the interleaving.
	for i, child := range children {
		want := fmt.Sprintf("child-%d", i)
		deep, err := child.Invoke[*Deep]()
		if err != nil {
			t.Fatalf("child %d invoke *Deep: %v", i, err)
		}
		if deep.Z.B.S != want {
			t.Errorf("child %d Deep.Z.B = %q, want %q", i, deep.Z.B.S, want)
		}
	}
}

// A value shared by ProvideValue is held by every container it cascaded into,
// and each of those shuts down the services it holds, so its Shutdown runs once
// per container rather than once overall.
func TestSharedValueShutsDownOncePerContainer(t *testing.T) {
	var mu sync.Mutex
	var closed []string

	root := New()
	root.ProvideValue(&closer{name: "shared", closed: &closed, mu: &mu})
	child := root.Derive()
	grandchild := child.Derive()

	root.MustInvoke[*closer]()
	child.MustInvoke[*closer]()
	grandchild.MustInvoke[*closer]()

	if err := root.Shutdown(); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(closed) != 3 {
		t.Errorf("shutdown calls = %d (%v), want 3, one per container holding the shared instance", len(closed), closed)
	}
}

// A parent Override reaches children that have not claimed the type themselves.
func TestParentOverrideCascadesToUnclaimedChild(t *testing.T) {
	root := New()
	root.ProvideValue(&Bar{S: "root"})
	root.Provide(ProvideZaz)
	root.Provide(ProvideDeep)
	root.Provide(ProvideLate)

	child := root.Derive()
	grandchild := child.Derive()

	root.OverrideValue(&Bar{S: "root-v2"})

	for _, tc := range []struct {
		name string
		c    *Container
	}{{"root", root}, {"child", child}, {"grandchild", grandchild}} {
		if got := deepestBar(t, tc.c); got != "root-v2" {
			t.Errorf("%s Late.D.Z.B = %q, want %q", tc.name, got, "root-v2")
		}
	}
}

// ...but stops at any container that claimed the type itself.
func TestParentOverrideDoesNotCascadePastChildOverride(t *testing.T) {
	root := New()
	root.ProvideValue(&Bar{S: "root"})
	root.Provide(ProvideZaz)
	root.Provide(ProvideDeep)
	root.Provide(ProvideLate)

	claimed := root.Derive()
	claimed.OverrideValue(&Bar{S: "claimed"})
	belowClaimed := claimed.Derive()

	unclaimed := root.Derive()

	root.OverrideValue(&Bar{S: "root-v2"})

	for _, tc := range []struct {
		name string
		c    *Container
		want string
	}{
		{"root", root, "root-v2"},
		{"claimed", claimed, "claimed"},
		{"belowClaimed", belowClaimed, "claimed"},
		{"unclaimed", unclaimed, "root-v2"},
	} {
		if got := deepestBar(t, tc.c); got != tc.want {
			t.Errorf("%s Late.D.Z.B = %q, want %q", tc.name, got, tc.want)
		}
	}
}

// The nearest ancestor that set a type outranks every further one.
func TestNearerAncestorOutranksFurtherAncestor(t *testing.T) {
	root := New()
	root.ProvideValue(&Bar{S: "root"})
	root.Provide(ProvideZaz)

	middle := root.Derive()
	leaf := middle.Derive()

	// middle claims Bar after leaf already exists, so it must cascade down.
	middle.OverrideValue(&Bar{S: "middle"})
	if got := leaf.MustInvoke[*Zaz]().B.S; got != "middle" {
		t.Errorf("leaf Zaz.B = %q, want %q", got, "middle")
	}

	// The root is further away, so it must not displace middle's claim.
	root.OverrideValue(&Bar{S: "root-v2"})
	if got := leaf.MustInvoke[*Bar]().S; got != "middle" {
		t.Errorf("leaf Bar = %q, want %q", got, "middle")
	}
	if got := middle.MustInvoke[*Bar]().S; got != "middle" {
		t.Errorf("middle Bar = %q, want %q", got, "middle")
	}
	if got := root.MustInvoke[*Bar]().S; got != "root-v2" {
		t.Errorf("root Bar = %q, want %q", got, "root-v2")
	}
}

// A container's own claim outranks an ancestor's, regardless of ordering.
func TestDeepestClaimWins(t *testing.T) {
	root := New()
	root.ProvideValue(&Bar{S: "root"})
	root.Provide(ProvideZaz)

	middle := root.Derive()
	leaf := middle.Derive()
	leaf.OverrideValue(&Bar{S: "leaf"})

	middle.OverrideValue(&Bar{S: "middle"})
	root.OverrideValue(&Bar{S: "root-v2"})

	if got := leaf.MustInvoke[*Bar]().S; got != "leaf" {
		t.Errorf("leaf Bar = %q, want %q", got, "leaf")
	}
	if got := middle.MustInvoke[*Bar]().S; got != "middle" {
		t.Errorf("middle Bar = %q, want %q", got, "middle")
	}
}

// Re-registering at the same depth is an update, not a demotion: the ancestor
// that already owned the type may keep revising it.
func TestAncestorMayReviseItsOwnClaim(t *testing.T) {
	root := New()
	root.ProvideValue(&Bar{S: "v1"})
	child := root.Derive()

	root.OverrideValue(&Bar{S: "v2"})
	root.OverrideValue(&Bar{S: "v3"})

	if got := child.MustInvoke[*Bar]().S; got != "v3" {
		t.Errorf("child Bar = %q, want %q", got, "v3")
	}
}

// Provide on a child shadows an inherited type and then holds it against the
// ancestor, the same as Override does.
func TestChildProvideClaimsAgainstLaterParentOverride(t *testing.T) {
	root := New()
	root.ProvideValue(&Bar{S: "root"})
	child := root.Derive()
	child.ProvideValue(&Bar{S: "child"})

	root.OverrideValue(&Bar{S: "root-v2"})

	if got := child.MustInvoke[*Bar]().S; got != "child" {
		t.Errorf("child Bar = %q, want %q", got, "child")
	}
}
