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

func (b *Backend) Columns(tableName string) ([]object.Column, error) {
	// Get the columns from the bucket for this table
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
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
