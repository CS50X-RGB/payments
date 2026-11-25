package wallet

import (
	"context"
	"fmt"

	"github.com/CS50X-RGB/payments/internal/wallet/walletdb"
	"github.com/google/uuid"
)

// Posting is one line of transaction: money in or out of one account
type Posting struct {
	AccountID uuid.UUID
	Direction Direction
	Amount    int64
	Cuurency  string
}

type NewTransaction struct {
	Description string
	Postings    []Posting
}

type Transaction struct {
	walletdb.Transaction
	Enties []Entry `json:"entries"`
}

// inTx runs fn inside a db transaction,handing it queries bound to that
// transaction.It commits when fn returns nil,and rolls back otherwise
func (s *Store) inTx(ctx context.Context, fn func(q *walletdb.Queries) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin : %w", err)
	}
	// Rollback is a no op once the transaction has commited,so deffering it
	// covers every early return,and even a panic,in one line.
	defer tx.Rollback(ctx)

	if err := fn(s.q.WithTx(tx)); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// PostTransaction() records the transaction and all of its tranaction and all of its entries atomically.
func (s *Store) PostTransaction(ctx context.Context, nt NewTransaction) (Transaction, error) {
	var posted Transaction
	err := s.inTx(ctx, func(q *walletdb.Queries) error {
		journal, err := q.CreateTransaction(ctx, nt.Description)
		if err != nil {
			return fmt.Errorf("insert transaction: %w", err)
		}
		posted.Transaction = journal

		for _, p := range nt.Postings {
			entry, err := q.CreateEntry(ctx, walletdb.CreateEntryParams{
				TransactionID: journal.ID,
				AccountID:     p.AccountID,
				Direction:     p.Direction,
				Amount:        p.Amount,
				Currency:      p.Cuurency,
			})
			if err != nil {
				return fmt.Errorf("insert entry : %w", err)
			}
			posted.Enties = append(posted.Enties, entry)
		}
		return nil
	})
	if err != nil {
		return Transaction{}, fmt.Errorf("post transaction : %w", err)
	}
	return posted, nil
}
