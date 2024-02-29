This package contains test implementations for all the components that compose OCR2 and OCR3 LOOPPs.

The directory layout is intended to mirror the natural interfaces composition of libocr:
- ocr2: contains static test implementations for reuseable reporting plugin, factory, and plugin provider
- ocr3: implements the ocr3 generic-based reporting plugin and reused common components from ocr2
- resources: represents core-node provided resources used to in calls to core node api(s) eg pipeline_runner, keystore, etc
- median: product specific test implementations required to fulfill the MedianProvider abstraction and median factory generator construction
- mercury: product specific test implementation required to fulfill MercuryProvider and constructors

Every test implementation follows the pattern wrapping an interface and provider one or more funcs to compare to another
instance of the interface. Every package attempts to share exposed the bare minimum surface area to avoid entanglement with logically seperated tests and domains.

In practice this is accomplished by exporting a static implementation and an interface that the implementation satisfies. The interface is used by other packages to declaration dependencies and the implementation is used in the test implementation of those other packaged

Example

```
package test_types

type Evaluator[T any] interface {
     // run methods of other, returns first error or error if
    // result of method invocation does not match expected value
    Evaluate(ctx context.Context, other T) error
}

```
package x_test

var FooEvaluator = staticFoo{
    expectedStr = "a"
    expectedInt = 1
}


type staticFoo struct {
    types.Foo // the interface to be tested, used to compose a loop
    expectedStr string
    expectedInt int
    ...
}

var _ test_types.Evaulator[Foo] = staticFoo
var _ types.Foo = staticFoo

func (f Foo) Evaluate(ctx context.Context, other Foo) {
    // test implementation of types.Foo interface
    s, err := other.GetStr()
    if err ! = nil {
        return fmt.Errorf("other failed to get str: %w", err)
    } 
    if s != f.expectedStr {
        return fmt.Errorf(" expected str %s got %s", s.expectedStr, s)
    }
    ...
}

// implements types.Foo
func (f Foo) GetStr() (string, error) {
    return f.expectedStr, nil
}
...

```

```
package y_test

var BarEvaluator = staticBar{
    expectedFoo = x_test.FooImpl
    expectedBytes = []bytes{1:1}
}


type staticBar struct {
    types.Bar // test implementation of Bar interface
    expectedFoo types_test[types.Foo]
    expectedInt int
    ...
}

// BarEvaluation implement [types.Bar] and [types_test.Evaluator[types.Bar]] to be used in tests
var _ BarEvaluator = staticBar {
    expectedFoo x_test.FooEvaluator
    expectedInt = 7
    ...
}

var _ types_test[types.Bar] = staticBar

// implement types.Bar
...
// implement types_test.Evaluator[types.Bar]
...

```