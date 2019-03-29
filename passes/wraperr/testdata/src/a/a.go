package a

import (
	"a/lib"
	"context"

	"cloud.google.com/go/spanner"
	"google.golang.org/grpc/status"
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

type grpcStatusErr struct {
	error
}

// see: https://github.com/googleapis/google-cloud-go/issues/1223
func (*grpcStatusErr) GRPCStatus() *status.Status {
	return nil
}

func f5(ctx context.Context, client *spanner.Client) {
	client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		stmt := spanner.Statement{SQL: `SELECT 1`}
		_, err := client.Single().Query(ctx, stmt).Next()
		if err != nil {
			return &grpcStatusErr{err} // OK
		}
		return nil
	})
}

func f6(ctx context.Context, client *spanner.Client) {
	client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		return func() error {
			stmt := spanner.Statement{SQL: `SELECT 1`}
			_, err := client.Single().Query(ctx, stmt).Next()
			if err != nil {
				return err
			}
			return nil
		}() // OK
	})
}

func f7(ctx context.Context, client *spanner.Client) {
	client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		return lib.F(ctx, client) // OK
	})
}

func f8(ctx context.Context, client *spanner.Client) {
	client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		return lib.Err() // OK - other pacakge
	})
}

func f9(ctx context.Context, client *spanner.Client) {
	client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		return lib.SpannerErr() // OK
	})
}
