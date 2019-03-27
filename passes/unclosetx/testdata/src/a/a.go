package a

import (
	"context"

	"cloud.google.com/go/spanner"
)

func f1(ctx context.Context, client *spanner.Client) {
	client.ReadOnlyTransaction()         // want "transaction must be closed"
	client.ReadOnlyTransaction().Close() // OK
	tx := client.ReadOnlyTransaction()   // OK
	tx.Close()
	client.Single() // OK
}

func f2(ctx context.Context, client *spanner.Client) {
	tx1 := client.ReadOnlyTransaction() // want "transaction must be closed"
	if tx1 != nil {
		tx1.Close()
	}

	tx2 := client.ReadOnlyTransaction() // want "transaction must be closed"
	if tx2 != nil {
		defer tx2.Close()
	}

	tx3 := client.ReadOnlyTransaction() // OK
	defer tx3.Close()
}

func f3(ctx context.Context, client *spanner.Client) interface{} {
	tx := client.ReadOnlyTransaction() // want "transaction must be closed"
	return struct {
		tx *spanner.ReadOnlyTransaction
	}{
		tx: tx,
	}
}

func f4(ctx context.Context, client *spanner.Client) interface{} {
	tx := client.ReadOnlyTransaction() // OK
	defer tx.Close()
	return struct {
		tx *spanner.ReadOnlyTransaction
	}{
		tx: tx,
	}
}
