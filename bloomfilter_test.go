// Package bloomfilter is face-meltingly fast, thread-safe,
// marshalable, unionable, probability- and
// optimal-size-calculating Bloom filter in go
//
// https://github.com/steakknife/bloomfilter
//
// Copyright © 2014, 2015, 2018 Barry Allard
//
// MIT license
//
package bloomfilter

import (
	"math/rand"
	"testing"
)

// a read-only type that conforms to hash.Hash64, but only Sum64() works.
// It is set by writing the underlying value.
type hashableUint64 uint64

func (h hashableUint64) Write([]byte) (int, error) {
	panic("Unimplemented")
}

func (h hashableUint64) Sum([]byte) []byte {
	panic("Unimplemented")
}

func (h hashableUint64) Reset() {
	panic("Unimplemented")
}

func (h hashableUint64) BlockSize() int {
	panic("Unimplemented")
}

func (h hashableUint64) Size() int {
	panic("Unimplemented")
}

func (h hashableUint64) Sum64() uint64 {
	return uint64(h)
}

func hashableUint64Values() []hashableUint64 {
	return []hashableUint64{
		0,
		7,
		0x0c0ffee0,
		0xdeadbeef,
		0xffffffff,
	}
}

func hashableUint64NotValues() []hashableUint64 {
	return []hashableUint64{
		1,
		5,
		42,
		0xa5a5a5a5,
		0xfffffffe,
	}
}

func Test0(t *testing.T) {
	bf, _ := New(10000, 5)

	t.Log("Filled ratio before adds :", bf.PreciseFilledRatio())
	for _, x := range hashableUint64Values() {
		bf.Add(x)
	}
	t.Log("Filled ratio after adds :", bf.PreciseFilledRatio())

	// these may or may not be true
	for _, y := range hashableUint64Values() {
		if bf.Contains(y) {
			t.Log("value in set querties: may contain ", y)
		} else {
			t.Fatal("value in set queries: definitely does not contain ", y,
				", but it should")
		}
	}

	// these must all be false
	for _, z := range hashableUint64NotValues() {
		if bf.Contains(z) {
			t.Log("value not in set queries: may or may not contain ", z)
		} else {
			t.Log("value not in set queries: definitely does not contain ", z,
				" which is correct")
		}
	}
}

func BenchmarkAddX10kX5(b *testing.B) {
	bf, _ := New(10000, 5)
	b.Run("add-10kx5", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			bf.Add(hashableUint64(rand.Uint32()))
		}
	})
	b.Run("add-10kx5-hash", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			bf.AddHash(uint64(rand.Uint32()))
		}
	})
}

func TestAddX10kX5(t *testing.T) {
	b1, _ := New(10000, 5)
	b2, _ := b1.NewCompatible()

	verify := func() {
		for i := 0; i < len(b1.bits); i++ {
			if b1.bits[i] != b2.bits[i] {
				t.Fatalf("error at bit %d!", i)
			}
		}
	}
	for i := 0; i < 1000000; i++ {
		v := hashableUint64(rand.Uint32())
		b1.Add(v)
		b2.AddHash(v.Sum64())
		verify()
		if !b2.Contains(v) {
			t.Fatal("contain error")
		}
	}
}
func BenchmarkContains1kX10kX5(b *testing.B) {
	bf, _ := New(10000, 5)
	for i := 0; i < 1000; i++ {
		bf.Add(hashableUint64(rand.Uint32()))
	}
	b.Run("contains", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bf.Contains(hashableUint64(rand.Uint32()))
		}
	})
	b.Run("containsHash", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bf.ContainsHash(uint64(rand.Uint32()))
		}
	})
}

func BenchmarkContains100kX10BX20(b *testing.B) {
	rand.Seed(1337)
	b.StopTimer()
	bf, _ := New(10*1000*1000*1000, 20)
	for i := 0; i < 100*1000; i++ {
		bf.Add(hashableUint64(rand.Uint32()))
	}
	b.Run("contains", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bf.Contains(hashableUint64(rand.Uint32()))
		}
	})
	b.Run("containshash", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bf.ContainsHash(uint64(rand.Uint32()))
		}
	})
}

func TestContains(t *testing.T) {
	rand.Seed(1337)
	bf, _ := New(10*1000*1000, 20)
	for i := 0; i < 100*10000; i++ {
		x := hashableUint64(rand.Uint32())
		bf.Add(x)
		if !bf.Contains(x) {
			t.Fatalf("Did not contain newly added elem: %d", x.Sum64())
		}
	}
}

//BenchmarkUnionInPlace/union-8-6         	   15270	     77848 ns/op
func BenchmarkUnionInPlace(b *testing.B) {
	var filters []*Filter
	b1, _ := New(813129, 6)
	for i := 0; i < 2000; i++ {
		b1.Add(hashableUint64(rand.Uint32()))
	}
	filters = append(filters, b1)
	for i := 0; i < 7; i++ {
		b, _ := b1.NewCompatible()
		filters = append(filters, b)
	}
	b.ResetTimer()
	b.Run("union-8", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, bx := range filters {
				b1.UnionInPlace(bx)
			}
		}
	})
}

func BenchmarkContains94percentMisses(b *testing.B) {
	// This test should produce about
	// 5.4K hits and 94k misses
	rand.Seed(1337)
	b.StopTimer()
	bf, _ := New(10*1000*1000, 20)
	for i := 0; i < 100*1000; i++ {
		bf.Add(hashableUint64(rand.Uint32()))
	}
	b.Run("contains", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bf.Contains(hashableUint64(rand.Uint32()))
		}
	})
	b.Run("containsHash", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bf.ContainsHash(uint64(rand.Uint32()))
		}
	})
}
