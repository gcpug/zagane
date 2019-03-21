package a

import (
	"context"

	"cloud.google.com/go/spanner"
)

func f(ctx context.Context, client *spanner.Client) {
	stmt := spanner.Statement{SQL: `SELECT 1`}
	_, _ = client.Single().Query(ctx, stmt).Next() // want "iterator must be stop"
	client.Single().Query(ctx, stmt).Stop()        // OK
	defer client.Single().Query(ctx, stmt).Stop()  // OK
}

func g(ctx context.Context, client *spanner.Client) {
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

func h(ctx context.Context, client *spanner.Client) {
	stmt := spanner.Statement{SQL: `SELECT 1`}
	iter := client.Single().Query(ctx, stmt) // want "iterator must be stop"
	if iter == nil {
		defer iter.Stop()
	}
}
