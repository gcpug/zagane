package a

import (
	"context"

	"cloud.google.com/go/spanner"
)

func f1(ctx context.Context, client *spanner.Client) {
	stmt := spanner.Statement{SQL: `SELECT 1`}
	_, _ = client.Single().Query(ctx, stmt).Next() // want "iterator must be stop"
	client.Single().Query(ctx, stmt).Stop()        // OK
	defer client.Single().Query(ctx, stmt).Stop()  // OK
}

func f2(ctx context.Context, client *spanner.Client) {
	stmt := spanner.Statement{SQL: `SELECT 1`}
	iter1 := client.Single().Query(ctx, stmt) // want "iterator must be stop"
	if iter1 == nil {
		iter1.Stop()
	}

	iter2 := client.Single().Query(ctx, stmt) // OK
	if iter2 == nil {
		iter2.Stop()
	}
	iter2.Stop()
}

func f3(ctx context.Context, client *spanner.Client) {
	stmt := spanner.Statement{SQL: `SELECT 1`}
	iter := client.Single().Query(ctx, stmt) // want "iterator must be stop"
	if iter == nil {
		defer iter.Stop()
	}
}

func f4(ctx context.Context, client *spanner.Client) *spanner.RowIterator {
	stmt := spanner.Statement{SQL: `SELECT 1`}
	return client.Single().Query(ctx, stmt) // OK
}

func f5(ctx context.Context, client *spanner.Client) *spanner.RowIterator {
	stmt := spanner.Statement{SQL: `SELECT 1`}
	iter := client.Single().Query(ctx, stmt) // want "iterator must be stop"
	if iter == nil {
		iter.Stop()
	}
	return client.Single().Query(ctx, stmt) // OK
}

func f6(ctx context.Context, client *spanner.Client) {
	iter := func() *spanner.RowIterator {
		stmt := spanner.Statement{SQL: `SELECT 1`}
		return client.Single().Query(ctx, stmt) // OK
	}() // want "iterator must be stop"
	if iter == nil {
		iter.Stop()
	}
}

func f7(ctx context.Context, client *spanner.Client) {
	stmt := spanner.Statement{SQL: `SELECT 1`}
	iter := client.Single().Query(ctx, stmt) // OK
	iter.Do(nil)
}
