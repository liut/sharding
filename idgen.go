package sharding

import (
	"sync/atomic"
	"time"
)

const (
	shardBits = 11
	seqBits   = 12
)

const (
	epoch     = int64(1262304000000)    // 2010-01-01 00:00:00 +0000 UTC
	shardMask = int64(1)<<shardBits - 1 // 2047
	seqMask   = int64(1)<<seqBits - 1   // 4095
)

// IdGen generates sortable unique int64 numbers that consist of:
// - 41 bits for time in milliseconds.
// - 11 bits for shard id.
// - 12 bits for auto-incrementing sequence.
//
// As a result we can generate 4096 ids per millisecond for each of 2048 shards.
// Minimum supported time is 1976-01-01, maximum is 2044-12-31.
type IdGen struct {
	seq   int64
	shard int64
}

// NewIdGen returns id generator for the shard.
func NewIdGen(shard int64) *IdGen {
	return &IdGen{
		shard: shard % 2048,
	}
}

// NextTime returns incremental id for the time. Note that you can only
// generate 4096 unique numbers per millisecond.
func (g *IdGen) NextTime(tm time.Time) int64 {
	seq := atomic.AddInt64(&g.seq, 1) - 1
	id := tm.UnixNano()/int64(time.Millisecond) - epoch
	id <<= (shardBits + seqBits)
	id |= g.shard << seqBits
	id |= seq % (seqMask + 1)
	return id
}

// Next acts like NextTime, but returns id for the current time.
func (g *IdGen) Next() int64 {
	return g.NextTime(time.Now())
}

// MaxTime returns max id for the time.
func (g *IdGen) MaxTime(tm time.Time) int64 {
	id := tm.UnixNano()/int64(time.Millisecond) - epoch
	id <<= (shardBits + seqBits)
	id |= g.shard << seqBits
	id |= seqMask
	return id
}

// SplitId splits id into time, shard id, and sequence id.
func SplitId(id int64) (tm time.Time, shardId int64, seqId int64) {
	ms := id>>(shardBits+seqBits) + epoch
	sec := ms / 1000
	tm = time.Unix(sec, (ms-sec*1000)*int64(time.Millisecond))
	shardId = (id >> seqBits) & shardMask
	seqId = id & seqMask
	return
}

// MaxIdTime returns max id for the time.
func MaxIdTime(tm time.Time) int64 {
	return NewIdGen(shardMask).MaxTime(tm)
}
