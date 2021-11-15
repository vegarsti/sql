package bolt

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/vegarsti/sql/object"
)

type Backend struct {
	file string
	db   *bolt.DB
}

func NewBackend(filename string) *Backend {
	return &Backend{file: filename}
}

func (b *Backend) Open() error {
	db, err := bolt.Open(b.file, 0600, nil)
	if err != nil {
		return fmt.Errorf("bolt open: %w", err)
	}
	b.db = db
	return nil
}

func (b *Backend) Close() error {
	if err := b.db.Close(); err != nil {
		return fmt.Errorf("bolt close: %w", err)
	}
	return nil
}

// CreateTable creates a Bolt bucket with the table name in the b.file Bolt file.
func (b *Backend) CreateTable(tableName string, columns []object.Column) error {
	// Create a bucket for this table and insert columns as JSON
	if err := b.db.Update(func(tx *bolt.Tx) error {
		tableBucketName := []byte(tableName)
		bucket, err := tx.CreateBucket(tableBucketName)
		if err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}
		marshalledColumns, err := json.Marshal(columns)
		if err != nil {
			return fmt.Errorf("json marshal columns: %w", err)
		}
		if err := bucket.Put([]byte("columns"), marshalledColumns); err != nil {
			return fmt.Errorf("bucket put columns: %w", err)
		}
		return nil
	}); err != nil {
		// return a nice error message if the table exists
		if errors.Is(err, bolt.ErrBucketExists) {
			return fmt.Errorf("table %s already exists", tableName)
		}
		return fmt.Errorf("update: %w", err)
	}
	return nil
}

// Insert inserts a row in the bucket for this table.
// When a row has been inserted, we increment the bucket sequence number.
// This number thus shows how many rows there are in the table, and is used when iterating over the rows in Backend.Rows().
// The n'th row is stored with the byte representation of n as its key, and
// the bytes stored contain the marshalled JSON representation of the `object.Row`.
func (b *Backend) Insert(tableName string, row object.Row) error {
	// Create a bucket for this table and insert columns as JSON
	if err := b.db.Update(func(tx *bolt.Tx) error {
		tableBucketName := []byte(tableName)
		bucket := tx.Bucket(tableBucketName)
		if bucket == nil {
			return fmt.Errorf("table %s doesn't exist", tableName)
		}
		marshalledRow, err := json.Marshal(row)
		if err != nil {
			return fmt.Errorf("json marshal row: %w", err)
		}
		id := itob(bucket.Sequence())
		if err := bucket.Put(id, marshalledRow); err != nil {
			return fmt.Errorf("bucket put row: %w", err)
		}
		if _, err := bucket.NextSequence(); err != nil {
			return fmt.Errorf("bucket next sequence: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("update: %w", err)
	}
	return nil
}

// Rows returns a slice of all the rows in the table.
func (b *Backend) Rows(tableName string) ([]object.Row, error) {
	var rows []object.Row
	if err := b.db.View(func(tx *bolt.Tx) error {
		tableBucketName := []byte(tableName)
		bucket := tx.Bucket(tableBucketName)
		if bucket == nil {
			return fmt.Errorf("table %s doesn't exist", tableName)
		}
		columns, err := b.Columns(tableName)
		if err != nil {
			return fmt.Errorf("columns: %w", err)
		}
		for i := uint64(0); i < bucket.Sequence(); i++ {
			var row object.Row
			row.Values = make([]object.Object, len(columns))
			// Since row.Values is a slice of interface values, we must indicate which implementation of
			// the interface is used. Since tables cannot change, we know that the columns of the tables
			// are static. This means that if the 2nd value in a row must be a value of the 2nd column type.
			for i, v := range columns {
				switch v.Type {
				case object.STRING:
					row.Values[i] = &object.String{}
				case object.INTEGER:
					row.Values[i] = &object.Integer{}
				case object.FLOAT:
					row.Values[i] = &object.Float{}
				default:
					panic(fmt.Sprintf("unknown type %s", v.Type))
				}
			}
			marshalledRow := bucket.Get(itob(i))
			if marshalledRow == nil {
				return fmt.Errorf("row %d not found", i)
			}
			if err := json.Unmarshal(marshalledRow, &row); err != nil {
				return fmt.Errorf("json unmarshal row: %w", err)
			}
			rows = append(rows, row)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("view: %w", err)
	}
	return rows, nil
}

// Columns returns the column information for this table, if it exists
func (b *Backend) Columns(tableName string) ([]object.Column, error) {
	var columns []object.Column
	if err := b.db.View(func(tx *bolt.Tx) error {
		tableBucketName := []byte(tableName)
		bucket := tx.Bucket(tableBucketName)
		if bucket == nil {
			return fmt.Errorf("table %s doesn't exist", tableName)
		}
		marshalledColumns := bucket.Get([]byte("columns"))
		if marshalledColumns == nil {
			return fmt.Errorf("no columns found for table %s", tableName)
		}
		if err := json.Unmarshal(marshalledColumns, &columns); err != nil {
			return fmt.Errorf("json unmarshal columns: %w", err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("view: %w", err)
	}
	return columns, nil
}

// itob returns an 8-byte big endian representation of v.
// We use this to generate keys for each row.
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
