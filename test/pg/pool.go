package pg

import (
	"context"

	"github.com/caarlos0/env/v6"
	"github.com/jackc/pgx/v5/pgxpool"
	. "github.com/onsi/gomega"
)

type Config struct {
	DBConnectionURL string `env:"DB_CONNECTION_URL" envDefault:"postgres://user:pass@localhost:5432/processdb"`
}

func MustInitDBPool(ctx context.Context) *pgxpool.Pool {
	cfg := Config{}
	Expect(env.Parse(&cfg)).To(Succeed())

	poolCfg, err := pgxpool.ParseConfig(cfg.DBConnectionURL)
	Expect(err).NotTo(HaveOccurred())

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(pool.Ping(ctx)).To(Succeed())

	return pool
}
