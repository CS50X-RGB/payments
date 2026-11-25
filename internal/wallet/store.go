// Package wallet is the wallet service : accounts,balances,and the ledger.
package wallet

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"

	"github.com/CS50X-RGB/payments/internal/wallet/walletdb"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	Account        = walletdb.Account
	AccountBalance = walletdb.AccountBalance
	Entry          = walletdb.Entry
	NewAccount     = walletdb.CreateAccountParams
	Direction      = walletdb.EntryDirection
)

const (
	Credit = walletdb.EntryDirectionCredit
	Debit  = walletdb.EntryDirectionDebit
)

// ErrAccountNotFound is returned when an account does not exist.
var ErrAccountNotFound = errors.New("wallet : account not found")

// Store is the wallet's data access layer,often called a repository.It wraps
// the sqlc queries and owns db transactions.

type Store struct {
	pool *pgxpool.Pool
	q    *walletdb.Queries
}

// NewStore returns a Store that runs its queries on pool.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool, q: walletdb.New(pool)}
}

// CreateAccount() opens a new account.OwnerId is nil for system accounts.
func (s *Store) CreateAccount(ctx context.Context, a NewAccount) (Account, error) {
	acc, err := s.q.CreateAccount(ctx, a)
	if err != nil {
		return Account{}, fmt.Errorf("create Account: %w", err)
	}
	return acc, nil
}

// Balance returns an account's balance,derived from its entries bt the
// account_balances view.Accounts are minor units of the account's currency.
func (s *Store) Balance(ctx context.Context, accountid uuid.UUID) (AccountBalance, error) {
	b, err := s.q.GetAccountBalance(ctx, accountid)
	if errors.Is(err, pgx.ErrNoRows) {
		return AccountBalance{}, ErrAccountNotFound
	}
	if err != nil {
		return AccountBalance{}, fmt.Errorf("get balance %s:%x", accountid, err)
	}
	return b, nil
}

// Enties returns a pafe of an account's entries,newest first.Pass 0 as
// before for the first page,then id of the last entry you received.
func (s *Store) Entries(ctx context.Context, id uuid.UUID, before int64, limit int32) ([]Entry, error) {
	if before <= 0 {
		before = math.MaxInt64
	}

	entries, err := s.q.ListEntriesByAccount(ctx, walletdb.ListEntriesByAccountParams{
		AccountID: id,
		BeforeID:  before,
		PageSize:  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list entries %s: %w", id, err)
	}
	return entries, nil
}
