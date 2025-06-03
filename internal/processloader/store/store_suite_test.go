package store_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/jackc/pgx/v5/pgxpool"

	testdb "github.com/ggsomnoev/ntt-ds-sap-process-api/test/pg"
)

var (
	ctx  context.Context
	pool *pgxpool.Pool
)

func TestStore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Store Suite")
}

var _ = BeforeSuite(func() {
	ctx = context.Background()
	pool = testdb.MustInitDBPool(ctx)
})

var _ = AfterSuite(func() {
	pool.Close()
})
