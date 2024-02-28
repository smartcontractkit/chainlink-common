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
package x_test

var FooImpl = staticFoo{
    expectedStr = "a"
    expectedInt = 1
}

type FooEvaluator interface {
    foo // some interface in the used in the composition of a loop
    // run methods of other, returns first error or error if
    // result of method invocation does not match expected value of embedded foo
    Evaluate(ctx context.Context, other foo) error
}

type staticFoo struct {
    expectedStr string
    expectedInt int
    ...
}

var _ FooEvaluator = staticFoo

```

```
package y_test

var BarImpl = staticBar{
    expectedFoo = x_test.FooImpl
    expectedBytes = []bytes{1:1}
}

type BarEvaluator interface {
    bar // some interface in the used in the composition of a loop
    // run methods of other, returns first error or error if
    // result of method invocation does not match expected value of embedded foo
    Evaluate(ctx context.Context, other foo) error
}

type staticBar struct {
    expectedFoo x_test.FooEvaluator
    expectedInt int
    ...
}

var _ BarEvaluator = staticBar

```