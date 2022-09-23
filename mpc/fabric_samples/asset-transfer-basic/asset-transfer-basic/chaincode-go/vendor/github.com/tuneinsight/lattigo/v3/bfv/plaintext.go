package bfv

import (
	"github.com/tuneinsight/lattigo/v3/ring"
	"github.com/tuneinsight/lattigo/v3/rlwe"
)

// Plaintext is a Element with only one Poly. It represents a Plaintext element in R_q that is the
// result of scaling the corresponding element of R_t up by Q/t. This is a generic all-purpose type
// of plaintext: it will work with for all operations. It is however less compact than PlaintextRingT
// and will result in less efficient Ciphert-Plaintext multiplication than PlaintextMul. See bfv/encoder.go
// for more information on plaintext types.
type Plaintext struct {
	*rlwe.Plaintext
}

// PlaintextRingT represents a plaintext element in R_t.
// This is the most compact representation of a plaintext, but performing operations have the extra-cost of performing
// the scaling up by Q/t. See bfv/encoder.go for more information on plaintext types.
type PlaintextRingT Plaintext

// PlaintextMul represents a plaintext element in R_q, in NTT and Montgomery form, but without scale up by Q/t.
// A PlaintextMul is a special-purpose plaintext for efficient Ciphertext-Plaintext multiplication. However,
// other operations on plaintexts are not supported. See bfv/encoder.go for more information on plaintext types.
type PlaintextMul Plaintext

// NewPlaintext creates and allocates a new plaintext in RingQ (multiple moduli of Q).
// The plaintext will be in RingQ and scaled by Q/t.
// Slower encoding and larger plaintext size
func NewPlaintext(params Parameters) *Plaintext {
	plaintext := &Plaintext{rlwe.NewPlaintext(params.Parameters, params.MaxLevel())}
	return plaintext
}

// NewPlaintextLvl creates and allocates a new plaintext in RingQ (multiple moduli of Q).
// The plaintext is allocated with level+1 moduli.
// The plaintext will be in RingQ and scaled by Q/t.
// Slower encoding and larger plaintext size
func NewPlaintextLvl(params Parameters, level int) *Plaintext {
	plaintext := &Plaintext{rlwe.NewPlaintext(params.Parameters, level)}
	return plaintext
}

// NewPlaintextAtLevelFromPoly construct a new Plaintext at a specific level
// where the message is set to the passed poly. No checks are performed on poly and
// the returned Plaintext will share its backing array of coefficient.
func NewPlaintextAtLevelFromPoly(level int, poly *ring.Poly) *Plaintext {
	if len(poly.Coeffs) < level+1 {
		panic("cannot NewPlaintextAtLevelFromPoly: provided ring.Poly level is too small")
	}
	v0 := new(ring.Poly)
	v0.Coeffs = poly.Coeffs[:level+1]
	return &Plaintext{Plaintext: &rlwe.Plaintext{Value: v0}}
}

// NewPlaintextRingT creates and allocates a new plaintext in RingT (single modulus T).
// The plaintext will be in RingT.
func NewPlaintextRingT(params Parameters) *PlaintextRingT {
	plaintext := &PlaintextRingT{rlwe.NewPlaintext(params.Parameters, 0)}
	return plaintext
}

// NewPlaintextMul creates and allocates a new plaintext optimized for ciphertext x plaintext multiplication.
// The plaintext will be in the NTT and Montgomery domain of RingQ and not scaled by Q/t.
func NewPlaintextMul(params Parameters) *PlaintextMul {
	plaintext := &PlaintextMul{rlwe.NewPlaintext(params.Parameters, params.MaxLevel())}
	return plaintext
}

// NewPlaintextMulLvl creates and allocates a new plaintext optimized for ciphertext x plaintext multiplication.
// The plaintext is allocated with level+1 moduli.
// The plaintext will be in the NTT and Montgomery domain of RingQ and not scaled by Q/t.
func NewPlaintextMulLvl(params Parameters, level int) *PlaintextMul {
	plaintext := &PlaintextMul{rlwe.NewPlaintext(params.Parameters, level)}
	return plaintext
}
