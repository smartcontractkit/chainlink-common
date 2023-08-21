package pg

import (
	"time"
)

// QOpt pattern for ORM methods aims to clarify usage and remove some common footguns, notably:
// 1. It should be easy and obvious how to pass a parent context or a transaction into an ORM method
// 2. Simple queries should not be cluttered
// 3. It should have compile-time safety and be explicit
// 4. It should enforce some sort of context deadline on all queries by default
// 5. It should optimise for clarity and readability
// 6. It should mandate using sqlx everywhere, gorm is forbidden in new code
// 7. It should make using sqlx a little more convenient by wrapping certain methods
// 8. It allows easier mocking of DB calls (Queryer is an interface)
//
// The two main concepts introduced are:
//
// A `Q` struct that wraps a `sqlx.DB` or `sqlx.Tx` and implements the `pg.Queryer` interface.
//
// This struct is initialised with `QOpts` which define how the queryer should behave. `QOpts` can define a parent context, an open transaction or other options to configure the Queryer.
//
// A sample ORM method looks like this:
//
//	func (o *orm) GetFoo(id int64, qopts ...pg.QOpt) (Foo, error) {
//		q := pg.NewQ(q, qopts...)
//		return q.Exec(...)
//	}
//
// Now you can call it like so:
//
//	orm.GetFoo(1) // will automatically have default query timeout context set
//	orm.GetFoo(1, pg.WithParentCtx(ctx)) // will wrap the supplied parent context with the default query context
//	orm.GetFoo(1, pg.WithQueryer(tx)) // allows to pass in a running transaction or anything else that implements Queryer
//	orm.GetFoo(q, pg.WithQueryer(tx), pg.WithParentCtx(ctx)) // options can be combined
type QOpt func(*Q)

type Q struct {
	Queryer
	QueryTimeout time.Duration
}
