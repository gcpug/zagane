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
