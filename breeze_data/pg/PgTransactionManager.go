package pg

import (
	"context"
	"github.com/breezeframework/breeze_data/breeze_data"
	"github.com/breezeframework/breeze_data/breeze_data/transaction"
	"github.com/jackc/pgx/v5"

	"github.com/pkg/errors"
)

type PgTransactionManager struct {
	db breeze_data.Transactor
}

func NewPgTransactionManager(db breeze_data.Transactor) *PgTransactionManager {
	return &PgTransactionManager{
		db: db,
	}
}

func (m *PgTransactionManager) Transaction(ctx context.Context, opts transaction.TxOptions, fn breeze_data.TransactionalFlow) (err error) {
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
			err = errors.Errorf("panic recovered: %v", r)
		}

		if err != nil {
			if errRollback := tx.Rollback(ctx); errRollback != nil {
				err = errors.Wrapf(err, "errRollback: %v", errRollback)
			}

			return
		}

		if nil == err {
			err = tx.Commit(ctx)
			if err != nil {
				err = errors.Wrap(err, "tx commit failed")
			}
		}
	}()

	if err = fn(ctx); err != nil {
		err = errors.Wrap(err, "failed executing code inside transaction")
	}

	return err
}

func toPgOptions(txOptions transaction.TxOptions) pgx.TxOptions {
	return pgx.TxOptions{
		IsoLevel:       pgx.TxIsoLevel(txOptions.IsoLevel),
		AccessMode:     pgx.TxAccessMode(txOptions.AccessMode),
		DeferrableMode: pgx.TxDeferrableMode(txOptions.DeferrableMode),
	}
}

/*func (m *PgTransactionManager) ReadCommitted(ctx context.Context, f breeze_data.TransactionalFlow) error {
	txOpts := pgx.TxOptions{IsoLevel: pgx.ReadCommitted}
	return m.transaction(ctx, txOpts, f)
}*/
