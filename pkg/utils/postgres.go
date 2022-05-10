package utils

import (
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
)

func GetPostgresStats(pgPoolStat *pgxpool.Stat) string {
	return fmt.Sprintf("AcquireCount: {%v}, MaxConns: {%v}, AcquiredConns: {%v}, TotalConns: {%v}, CanceledAcquireCount: {%v}, ConstructingConns: {%v}, EmptyAcquireCount: {%v}, IdleConns: {%v},",
		pgPoolStat.AcquireCount(),
		pgPoolStat.MaxConns(),
		pgPoolStat.AcquiredConns(),
		pgPoolStat.TotalConns(),
		pgPoolStat.CanceledAcquireCount(),
		pgPoolStat.ConstructingConns(),
		pgPoolStat.EmptyAcquireCount(),
		pgPoolStat.IdleConns(),
	)
}
