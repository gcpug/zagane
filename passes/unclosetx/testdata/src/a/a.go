package a

import (
	"context"

	"cloud.google.com/go/spanner"
	"golang.org/x/sync/errgroup"
)

func f1(ctx context.Context, client *spanner.Client) {
	client.ReadOnlyTransaction()         // want "transaction must be closed"
	client.ReadOnlyTransaction().Close() // OK
	tx := client.ReadOnlyTransaction()   // OK
	tx.Close()
	client.Single()              // OK
	client.ReadOnlyTransaction() //lint:ignore zagane OK
	client.ReadOnlyTransaction() //lint:ignore unclosetx OK
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
	tx := client.ReadOnlyTransaction() // OK
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

func f5(ctx context.Context, client *spanner.Client) {
	tx, _ := client.BatchReadOnlyTransaction(ctx, spanner.StrongRead()) // want "transaction must be closed"
	_ = tx
}

// see https://github.com/gcpug/zagane/issues/32
func f6(ctx context.Context, client *spanner.Client) error {
	tx := client.ReadOnlyTransaction() // OK
	defer tx.Close()

	var eg errgroup.Group

	eg.Go(func() error {
		_ = tx // use tx
		return nil
	})

	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func f7(ctx context.Context, client *spanner.Client) error {
	ro, _ := client.BatchReadOnlyTransaction(ctx, spanner.StrongRead())
	defer ro.Close()

	stmt := spanner.Statement{SQL: `SELECT 1`}

	iter := ro.Query(ctx, stmt)
	defer iter.Stop()

	for {
		_, err := iter.Next()
		if err != nil {
			break
		}
	}

	return nil
}
