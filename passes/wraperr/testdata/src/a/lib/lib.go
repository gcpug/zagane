package lib

import (
	"context"
	"errors"

	"cloud.google.com/go/spanner"
)

func F(ctx context.Context, client *spanner.Client) error {
	stmt := spanner.Statement{SQL: `SELECT 1`}
	_, err := client.Single().Query(ctx, stmt).Next()
	if err != nil {
		return err
	}
	return nil
}

func Err() error {
	return errors.New("error")
}

func SpannerErr() error {
	return &spanner.Error{}
}
