package a

import (
	"context"

	"cloud.google.com/go/spanner"
)

type wrapErr struct {
	error
}

func wrap(err error) error {
	return &wrapErr{err}
}

func f1(ctx context.Context, client *spanner.Client) {
	client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		stmt := spanner.Statement{SQL: `SELECT 1`}
		_, err := client.Single().Query(ctx, stmt).Next()
		if err != nil {
			return err // OK
		}
		return nil
	})
}

func f2(ctx context.Context, client *spanner.Client) {
	client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		stmt := spanner.Statement{SQL: `SELECT 1`}
		_, err := client.Single().Query(ctx, stmt).Next()
		if err != nil {
			return wrap(err) // want "must not be wrapped"
		}
		return nil
	})
}

func f3(ctx context.Context, client *spanner.Client) {
	client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		stmt := spanner.Statement{SQL: `SELECT 1`}
		_, err := client.Single().Query(ctx, stmt).Next()
		if err != nil {
			return &wrapErr{err} // want "must not be wrapped"
		}
		return nil
	})
}

func f4(ctx context.Context, client *spanner.Client) {
	client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		stmt := spanner.Statement{SQL: `SELECT 1`}
		_, err := client.Single().Query(ctx, stmt).Next()
		if err != nil {
			return func() error {
				return &wrapErr{err}
			}() // want "must not be wrapped"
		}
		return nil
	})
}
