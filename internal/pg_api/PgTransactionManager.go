package pg_api

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/simpleGorm/pg/internal/transaction"
)

type PgTransactionManager struct {
	db Transactor
}

func NewPgTransactionManager(db Transactor) *PgTransactionManager {
	return &PgTransactionManager{
		db: db,
	}
}

func (m *PgTransactionManager) Transaction(ctx context.Context, opts transaction.TxOptions, fn TransactionalFlow) (err error) {
	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return fn(ctx)
	}
	pgOpts := toPgOptions(opts)

	tx, err = m.db.BeginTx(ctx, pgOpts)
	if err != nil {
		return errors.Wrap(err, "can't begin transaction")
	}

	ctx = MakeContextTx(ctx, tx)

	defer func() {
		if r := recover(); r != nil {
			err = errors.Errorf("[PgTransactionManager] panic recovered: %w", r)
		}

		if err != nil {
			if errRollback := tx.Rollback(ctx); errRollback != nil {
				//errRollback := errors.Wrapf(err, "errRollback: %v", errRollback)
				panic(errRollback)
			}
		}

		if nil == err {
			err = tx.Commit(ctx)
			if err != nil {
				err = errors.Wrap(err, "tx commit failed")
				panic(err)
			}
		}
	}()

	if err = fn(ctx); err != nil {
		err = errors.Wrap(err, "failed executing code inside transaction")
	}

	/*if err != nil {
		panic(err)
	}*/
	return err
}

func toPgOptions(txOptions transaction.TxOptions) pgx.TxOptions {
	return pgx.TxOptions{
		IsoLevel:       pgx.TxIsoLevel(txOptions.IsoLevel),
		AccessMode:     pgx.TxAccessMode(txOptions.AccessMode),
		DeferrableMode: pgx.TxDeferrableMode(txOptions.DeferrableMode),
	}
}

/*func (m *PgTransactionManager) ReadCommitted(ctx context.Context, f pkg.TransactionalFlow) error {
	txOpts := pgx.TxOptions{IsoLevel: pgx.ReadCommitted}
	return m.transaction(ctx, txOpts, f)
}*/
