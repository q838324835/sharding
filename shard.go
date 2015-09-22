package sharding

import (
	"io"
	"strconv"
	"strings"
	"time"

	"gopkg.in/pg.v3"
)

// Shard represents logical shard in Cluster.
type Shard struct {
	id     int64
	DB     *pg.DB
	oldnew []string
	repl   *strings.Replacer
}

func newShard(id int64, db *pg.DB, oldnew ...string) *Shard {
	return &Shard{
		id:     id,
		DB:     db,
		oldnew: oldnew,
		repl:   strings.NewReplacer(oldnew...),
	}
}

func (shard *Shard) Id() int64 {
	return shard.id
}

func (shard *Shard) Name() string {
	return "shard" + strconv.FormatInt(shard.id, 10)
}

func (shard *Shard) String() string {
	return shard.Name()
}

func (shard *Shard) UseTimeout(d time.Duration) *Shard {
	newShard := *shard
	newShard.DB = shard.DB.UseTimeout(d)
	return &newShard
}

func (shard *Shard) replaceVars(q string, args []interface{}) (string, error) {
	fq, err := pg.FormatQ(q, args...)
	if err != nil {
		return "", err
	}
	q = shard.repl.Replace(string(fq))
	return q, nil
}

func (shard *Shard) Exec(q string, args ...interface{}) (*pg.Result, error) {
	q, err := shard.replaceVars(q, args)
	if err != nil {
		return nil, err
	}
	return shard.DB.Exec(q)
}

func (shard *Shard) ExecOne(q string, args ...interface{}) (*pg.Result, error) {
	q, err := shard.replaceVars(q, args)
	if err != nil {
		return nil, err
	}
	return shard.DB.ExecOne(q)
}

func (shard *Shard) Query(coll pg.Collection, q string, args ...interface{}) (*pg.Result, error) {
	q, err := shard.replaceVars(q, args)
	if err != nil {
		return nil, err
	}
	return shard.DB.Query(coll, q)
}

func (shard *Shard) QueryOne(record interface{}, q string, args ...interface{}) (*pg.Result, error) {
	q, err := shard.replaceVars(q, args)
	if err != nil {
		return nil, err
	}
	return shard.DB.QueryOne(record, q)
}

func (shard *Shard) CopyFrom(r io.Reader, q string, args ...interface{}) (*pg.Result, error) {
	q, err := shard.replaceVars(q, args)
	if err != nil {
		return nil, err
	}
	return shard.DB.CopyFrom(r, q)
}

func (shard *Shard) CopyTo(w io.WriteCloser, q string, args ...interface{}) (*pg.Result, error) {
	q, err := shard.replaceVars(q, args)
	if err != nil {
		return nil, err
	}
	return shard.DB.CopyTo(w, q)
}

type Tx struct {
	shard *Shard
	Tx    *pg.Tx
}

func (shard *Shard) Begin() (*Tx, error) {
	tx, err := shard.DB.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{
		shard: shard,
		Tx:    tx,
	}, nil
}

func (tx *Tx) Commit() error {
	return tx.Tx.Commit()
}

func (tx *Tx) Rollback() error {
	return tx.Tx.Rollback()
}

func (tx *Tx) Exec(q string, args ...interface{}) (*pg.Result, error) {
	q, err := tx.shard.replaceVars(q, args)
	if err != nil {
		return nil, err
	}
	return tx.Tx.Exec(q)
}

func (tx *Tx) ExecOne(q string, args ...interface{}) (*pg.Result, error) {
	q, err := tx.shard.replaceVars(q, args)
	if err != nil {
		return nil, err
	}
	return tx.Tx.ExecOne(q)
}

func (tx *Tx) Query(coll pg.Collection, q string, args ...interface{}) (*pg.Result, error) {
	q, err := tx.shard.replaceVars(q, args)
	if err != nil {
		return nil, err
	}
	return tx.Tx.Query(coll, q)
}

func (tx *Tx) QueryOne(record interface{}, q string, args ...interface{}) (*pg.Result, error) {
	q, err := tx.shard.replaceVars(q, args)
	if err != nil {
		return nil, err
	}
	return tx.Tx.QueryOne(record, q)
}
