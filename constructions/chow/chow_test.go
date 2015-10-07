package chow

import (
	"bytes"
	"testing"

	"github.com/OpenWhiteBox/AES/primitives/matrix"

	test_vectors "github.com/OpenWhiteBox/AES/constructions/test"
)

var key = []byte{72, 101, 108, 108, 111, 32, 87, 111, 114, 108, 100, 33, 33, 33, 33, 33}

func TestShiftRows(t *testing.T) {
	in := []byte{99, 202, 183, 4, 9, 83, 208, 81, 205, 96, 224, 231, 186, 112, 225, 140}
	out := []byte{99, 83, 224, 140, 9, 96, 225, 4, 205, 112, 183, 81, 186, 202, 208, 231}

	constr, _, _ := GenerateKeys(key, key)
	constr.ShiftRows(in)

	if !bytes.Equal(out, in) {
		t.Fatalf("Real disagrees with result! %v != %v", out, in)
	}
}

func TestTyiTable(t *testing.T) {
	in := [16]byte{99, 83, 224, 140, 9, 96, 225, 4, 205, 112, 183, 81, 186, 202, 208, 231}
	out := [16]byte{95, 114, 100, 21, 87, 245, 188, 146, 247, 190, 59, 41, 29, 185, 249, 26}

	a, b, c, d := TyiTable(0), TyiTable(1), TyiTable(2), TyiTable(3)
	cand := [16]byte{}

	for i := 0; i < 16; i += 4 {
		e, f, g, h := a.Get(in[i+0]), b.Get(in[i+1]), c.Get(in[i+2]), d.Get(in[i+3])

		cand[i+0] = e[0] ^ f[0] ^ g[0] ^ h[0]
		cand[i+1] = e[1] ^ f[1] ^ g[1] ^ h[1]
		cand[i+2] = e[2] ^ f[2] ^ g[2] ^ h[2]
		cand[i+3] = e[3] ^ f[3] ^ g[3] ^ h[3]
	}

	if out != cand {
		t.Fatalf("Real disagrees with result! %v != %v", out, cand)
	}
}

func TestEncrypt(t *testing.T) {
	for n, vec := range test_vectors.AESVectors[0:10] {
		constr, input, output := GenerateKeys(vec.Key, vec.Key)

		inputInv, _ := input.Invert()
		outputInv, _ := output.Invert()

		in, out := make([]byte, 16), make([]byte, 16)

		copy(in, inputInv.Mul(matrix.Row(vec.In))) // Apply input encoding.

		constr.Encrypt(out, in)

		copy(out, outputInv.Mul(matrix.Row(out))) // Remove output encoding.

		if !bytes.Equal(vec.Out, out) {
			t.Fatalf("Real disagrees with result in test vector %v! %x != %x", n, vec.Out, out)
		}
	}
}

func TestPersistence(t *testing.T) {
	key := test_vectors.AESVectors[50].Key
	seed := test_vectors.AESVectors[51].Key
	input := test_vectors.AESVectors[50].In

	constr1, _, _ := GenerateKeys(key, seed)

	serialized := constr1.Serialize()
	constr2 := Parse(serialized)

	cand1, cand2 := make([]byte, 16), make([]byte, 16)

	constr1.Encrypt(cand1, input)
	constr2.Encrypt(cand2, input)

	if !bytes.Equal(cand1, cand2) {
		t.Fatalf("Real disagrees with parsed! %v != %v", cand1, cand2)
	}
}

// A "Live" Encryption is one based on table abstractions, so many computations are performed on-demand.
func BenchmarkLiveEncrypt(b *testing.B) {
	key := test_vectors.AESVectors[50].Key
	seed := test_vectors.AESVectors[51].Key
	input := test_vectors.AESVectors[50].In

	constr, _, _ := GenerateKeys(key, seed)

	out := make([]byte, 16)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		constr.Encrypt(out, input)
	}
}

// A "Dead" Encryption is one based on serialized tables, like we'd have in a real use case.
func BenchmarkDeadEncrypt(b *testing.B) {
	key := test_vectors.AESVectors[50].Key
	seed := test_vectors.AESVectors[51].Key
	input := test_vectors.AESVectors[50].In

	constr1, _, _ := GenerateKeys(key, seed)

	serialized := constr1.Serialize()
	constr2 := Parse(serialized)

	out := make([]byte, 16)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		constr2.Encrypt(out, input)
	}
}

func BenchmarkShiftRows(b *testing.B) {
	key := test_vectors.AESVectors[50].Key
	seed := test_vectors.AESVectors[51].Key
	input := test_vectors.AESVectors[50].In

	constr, _, _ := GenerateKeys(key, seed)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		constr.ShiftRows(input)
	}
}

func BenchmarkExpandWord(b *testing.B) {
	key := test_vectors.AESVectors[50].Key
	seed := test_vectors.AESVectors[51].Key
	input := test_vectors.AESVectors[50].In

	constr1, _, _ := GenerateKeys(key, seed)

	serialized := constr1.Serialize()
	constr2 := Parse(serialized)

	dst := make([]byte, 16)
	copy(dst, input)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		constr2.ExpandWord(constr2.TBoxTyiTable[0][0:4], dst[0:4])
	}
}

func BenchmarkExpandBlock(b *testing.B) {
	key := test_vectors.AESVectors[50].Key
	seed := test_vectors.AESVectors[51].Key
	input := test_vectors.AESVectors[50].In

	constr1, _, _ := GenerateKeys(key, seed)

	serialized := constr1.Serialize()
	constr2 := Parse(serialized)

	dst := make([]byte, 16)
	copy(dst, input)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		constr2.ExpandBlock(constr2.InputMask, dst)
	}
}

func BenchmarkSquashWords(b *testing.B) {
	key := test_vectors.AESVectors[50].Key
	seed := test_vectors.AESVectors[51].Key
	input := test_vectors.AESVectors[50].In

	constr1, _, _ := GenerateKeys(key, seed)

	serialized := constr1.Serialize()
	constr2 := Parse(serialized)

	dst := make([]byte, 16)
	copy(dst, input)

	stretched := constr2.ExpandWord(constr2.TBoxTyiTable[0][0:4], dst[0:4])

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		constr2.SquashWords(constr2.HighXORTable[0][0:8], stretched, dst[0:4])
		copy(dst[0:4], input)
	}
}

func BenchmarkSquashBlocks(b *testing.B) {
	key := test_vectors.AESVectors[50].Key
	seed := test_vectors.AESVectors[51].Key
	input := test_vectors.AESVectors[50].In

	constr1, _, _ := GenerateKeys(key, seed)

	serialized := constr1.Serialize()
	constr2 := Parse(serialized)

	dst := make([]byte, 16)
	copy(dst, input)

	stretched := constr2.ExpandBlock(constr2.InputMask, dst)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		constr2.SquashBlocks(constr2.InputXORTable, stretched, dst)
		copy(dst, input)
	}
}
