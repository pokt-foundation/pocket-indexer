package postgresdriver

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/pokt-foundation/pocket-go/provider"
	indexer "github.com/pokt-foundation/pocket-indexer-lib"
	"github.com/stretchr/testify/require"
)

func TestPostgresDriver_WriteTransactions(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	testProvStdTx := &provider.StdTx{
		Entropy: 3223323,
		Fee: []struct {
			Amount string "json:\"amount\""
			Denom  string "json:\"denom\""
		}{
			{
				Amount: "10000",
				Denom:  "upokt",
			},
		},
		Msg: struct {
			Type  string         "json:\"type\""
			Value map[string]any "json:\"value\""
		}{
			Type: "pos/Send",
			Value: map[string]any{
				"from_address": "addssd",
				"to_address":   "adasd",
				"chains": []any{
					"0021",
				},
			},
		},
		Signature: struct {
			PubKey    string "json:\"pub_key\""
			Signature string "json:\"signature\""
		}{
			PubKey: "adasdsfd",
		},
	}

	testStdTx := &stdTx{
		StdTx: testProvStdTx,
	}

	encodedTestStdTx, err := testStdTx.Value()
	c.NoError(err)

	mock.ExpectExec("INSERT into transactions").WithArgs("AF5BB3EAFF431E2E5E784D639825979FF20A779725BFE61D4521340F70C3996D0",
		"addssd", "adasd", "adasdsfd", pq.StringArray([]string{"0021"}), "pos/Send", int64(0), int64(0), []uint8{123, 125}, encodedTestStdTx,
		[]uint8{123, 125}, "", int64(3223323), int64(10000), "upokt").
		WillReturnResult(sqlmock.NewResult(1, 1))

	driver := NewPostgresDriverFromSQLDBInstance(db)

	transactionToSend := []*indexer.Transaction{
		{
			Hash:            "AF5BB3EAFF431E2E5E784D639825979FF20A779725BFE61D4521340F70C3996D0",
			FromAddress:     "addssd",
			ToAddress:       "adasd",
			AppPubKey:       "adasdsfd",
			Blockchains:     []string{"0021"},
			MessageType:     "pos/Send",
			Entropy:         3223323,
			Fee:             10000,
			FeeDenomination: "upokt",
			StdTx:           testProvStdTx,
		},
	}

	err = driver.WriteTransactions(transactionToSend)
	c.NoError(err)

	mock.ExpectExec("INSERT into transactions").WithArgs("AF5BB3EAFF431E2E5E784D639825979FF20A779725BFE61D4521340F70C3996D0",
		"addssd", "adasd", "adasdsfd", pq.StringArray([]string{"0021"}), "pos/Send", int64(0), int64(0), []uint8{123, 125}, encodedTestStdTx,
		[]uint8{123, 125}, "", int64(3223323), int64(10000), "upokt").
		WillReturnError(errors.New("dummy error"))

	err = driver.WriteTransactions(transactionToSend)
	c.EqualError(err, "dummy error")
}

func TestPostgresDriver_ReadTransactions(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	testStdTx := &stdTx{
		StdTx: &provider.StdTx{},
	}

	encodedTestStdTx, err := testStdTx.Value()
	c.NoError(err)

	testTxResult := &txResult{
		TxResult: &provider.TxResult{},
	}

	encodedTxResult, err := testTxResult.Value()
	c.NoError(err)

	testProof := &proof{
		TransactionProof: &provider.TransactionProof{},
	}

	encodedTestProof, err := testProof.Value()
	c.NoError(err)

	rows := sqlmock.NewRows([]string{"id", "hash", "from_address", "to_address", "stdtx", "tx_result", "proof"}).
		AddRow(1, "ABCD", "abcd", "dbcv", encodedTestStdTx, encodedTxResult, encodedTestProof).
		AddRow(2, "ABFD", "abfd", "fbcv", encodedTestStdTx, encodedTxResult, encodedTestProof)

	mock.ExpectQuery("^SELECT (.+) FROM transactions$").WillReturnRows(rows)

	driver := NewPostgresDriverFromSQLDBInstance(db)

	transactions, err := driver.ReadTransactions()
	c.NoError(err)
	c.Len(transactions, 2)

	mock.ExpectQuery("^SELECT (.+) FROM transactions$").WillReturnError(errors.New("dummy error"))

	transactions, err = driver.ReadTransactions()
	c.EqualError(err, "dummy error")
	c.Empty(transactions)
}

func TestPostgresDriver_ReadTransaction(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	testStdTx := &stdTx{
		StdTx: &provider.StdTx{},
	}

	encodedTestStdTx, err := testStdTx.Value()
	c.NoError(err)

	testTxResult := &txResult{
		TxResult: &provider.TxResult{},
	}

	encodedTxResult, err := testTxResult.Value()
	c.NoError(err)

	testProof := &proof{
		TransactionProof: &provider.TransactionProof{},
	}

	encodedTestProof, err := testProof.Value()
	c.NoError(err)

	rows := sqlmock.NewRows([]string{"id", "hash", "from_address", "to_address", "stdtx", "tx_result", "proof"}).
		AddRow(1, "ABCD", "abcd", "dbcv", encodedTestStdTx, encodedTxResult, encodedTestProof)

	mock.ExpectQuery("^SELECT (.+) FROM transactions (.+)").WillReturnRows(rows)

	driver := NewPostgresDriverFromSQLDBInstance(db)

	transaction, err := driver.ReadTransaction("ABCD")
	c.NoError(err)
	c.NotEmpty(transaction)

	mock.ExpectQuery("^SELECT (.+) FROM transactions (.+)").WillReturnError(errors.New("dummy error"))

	transaction, err = driver.ReadTransaction("ABCD")
	c.EqualError(err, "dummy error")
	c.Empty(transaction)
}

func TestPostgresDriver_WriteBlock(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	mock.ExpectExec("INSERT into blocks").WithArgs("AF5BB3EAFF431E2E5E784D639825979FF20A779725BFE61D4521340F70C3996D0",
		21, time.Date(1999, time.July, 21, 0, 0, 0, 0, time.Local), "A2143929B30CBC3E7A30C2DE06B385BCF874134B", 32, 21).
		WillReturnResult(sqlmock.NewResult(1, 1))

	driver := NewPostgresDriverFromSQLDBInstance(db)

	err = driver.WriteBlock(&indexer.Block{
		Hash:            "AF5BB3EAFF431E2E5E784D639825979FF20A779725BFE61D4521340F70C3996D0",
		Height:          21,
		Time:            time.Date(1999, time.July, 21, 0, 0, 0, 0, time.Local),
		ProposerAddress: "A2143929B30CBC3E7A30C2DE06B385BCF874134B",
		TXCount:         32,
		RelayCount:      21,
	})
	c.NoError(err)

	mock.ExpectExec("INSERT into blocks").WithArgs("AF5BB3EAFF431E2E5E784D639825979FF20A779725BFE61D4521340F70C3996D0",
		21, time.Date(1999, time.July, 21, 0, 0, 0, 0, time.Local), "A2143929B30CBC3E7A30C2DE06B385BCF874134B", 32, 21).
		WillReturnError(errors.New("dummy error"))

	err = driver.WriteBlock(&indexer.Block{
		Hash:            "AF5BB3EAFF431E2E5E784D639825979FF20A779725BFE61D4521340F70C3996D0",
		Height:          21,
		Time:            time.Date(1999, time.July, 21, 0, 0, 0, 0, time.Local),
		ProposerAddress: "A2143929B30CBC3E7A30C2DE06B385BCF874134B",
		TXCount:         32,
		RelayCount:      21,
	})
	c.EqualError(err, "dummy error")
}

func TestPostgresDriver_ReadBlocks(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "hash", "height", "time", "proposer_address", "tx_count", "relay_count"}).
		AddRow(1, "ABCD", 21, time.Date(1999, time.July, 21, 0, 0, 0, 0, time.Local), "ABCD", 21, 21).
		AddRow(2, "ABCD", 21, time.Date(1999, time.July, 21, 0, 0, 0, 0, time.Local), "ABCD", 21, 21)

	mock.ExpectQuery("^SELECT (.+) FROM blocks$").WillReturnRows(rows)

	driver := NewPostgresDriverFromSQLDBInstance(db)

	blocks, err := driver.ReadBlocks()
	c.NoError(err)
	c.Len(blocks, 2)

	mock.ExpectQuery("^SELECT (.+) FROM blocks$").WillReturnError(errors.New("dummy error"))

	blocks, err = driver.ReadBlocks()
	c.EqualError(err, "dummy error")
	c.Empty(blocks)
}

func TestPostgresDriver_ReadBlock(t *testing.T) {
	c := require.New(t)

	db, mock, err := sqlmock.New()
	c.NoError(err)

	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "hash", "height", "time", "proposer_address", "tx_count", "relay_count"}).
		AddRow(1, "ABCD", 21, time.Date(1999, time.July, 21, 0, 0, 0, 0, time.Local), "ABCD", 21, 21)

	mock.ExpectQuery("^SELECT (.+) FROM blocks (.+)").WillReturnRows(rows)

	driver := NewPostgresDriverFromSQLDBInstance(db)

	block, err := driver.ReadBlock("ABCD")
	c.NoError(err)
	c.NotEmpty(block)

	mock.ExpectQuery("^SELECT (.+) FROM blocks (.+)").WillReturnError(errors.New("dummy error"))

	block, err = driver.ReadBlock("ABCD")
	c.EqualError(err, "dummy error")
	c.Empty(block)
}
