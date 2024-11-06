package intset

import (
    "bytes"
    "fmt"
)

type IntSet struct {
    words []uint64
}

// Add adds a non-negative value x to the set.
func (s *IntSet) Add(x int) {
    word, bit := x/64, uint(x%64)
    for word >= len(s.words) {
        s.words = append(s.words, 0)
    }
    s.words[word] |= 1 << bit
}

// Has reports whether the set contains the non-negative value x.
func (s *IntSet) Has(x int) bool {
    word, bit := x/64, uint(x%64)
    return word < len(s.words) && s.words[word]&(1<<bit) != 0
}

// String returns the set as a string of the form "{1 2 3}".
func (s *IntSet) String() string {
    var buf bytes.Buffer
    buf.WriteByte('{')
    for i, word := range s.words {
        if word == 0 {
            continue
        }
        for j := 0; j < 64; j++ {
            if word&(1<<uint(j)) != 0 {
                if buf.Len() > len("{") {
                    buf.WriteByte(' ')
                }
                fmt.Fprintf(&buf, "%d", 64*i+j)
            }
        }
    }
    buf.WriteByte('}')
    return buf.String()
}

// Len returns the number of elements in the set.
func (s *IntSet) Len() int {
    count := 0
    for _, word := range s.words {
        count += popCount(word)
    }
    return count
}

// Remove removes x from the set.
func (s *IntSet) Remove(x int) {
    word, bit := x/64, uint(x%64)
    if word < len(s.words) {
        s.words[word] &^= 1 << bit
    }
}

// Clear removes all elements from the set.
func (s *IntSet) Clear() {
    s.words = nil
}

// Copy returns a copy of the set.
func (s *IntSet) Copy() *IntSet {
    copy := IntSet{}
    copy.words = make([]uint64, len(s.words))
    for i, word := range s.words {
        copy.words[i] = word
    }
    return &copy
}


// AddAll adds a list of values to the set.
func (s *IntSet) AddAll(values ...int) {
    for _, x := range values {
        s.Add(x)
    }
}

// UnionWith creates a set union with another set.
func (s *IntSet) UnionWith(t *IntSet) {
    for i, tword := range t.words {
        if i < len(s.words) {
            s.words[i] |= tword
        } else {
            s.words = append(s.words, tword)
        }
    }
}

// IntersectWith creates a set intersection with another set.
func (s *IntSet) IntersectWith(t *IntSet) {
    for i, word := range t.words {
        if i < len(s.words) {
            s.words[i] &= word
        }
    }
}

// DifferenceWith creates a set difference with another set.
func (s *IntSet) DifferenceWith(t *IntSet) {
    for i, word := range t.words {
        if i < len(s.words) {
            s.words[i] &^= word
        }
    }
}

// SymmetricDifference creates a set symmetric difference with another set.
func (s *IntSet) SymmetricDifference(t *IntSet) {
    for i, word := range t.words {
        if i < len(s.words) {
            s.words[i] ^= word
        } else {
            s.words = append(s.words, word)
        }
    }
}

// popCount returns the population count (number of set bits) of x.
func popCount(x uint64) int {
    count := 0
    for x != 0 {
        x &= x - 1
        count++
    }
    return count
}
