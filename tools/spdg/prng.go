package main

// mulberry32 is a deterministic PRNG that produces identical sequences
// to the JavaScript implementation for the same seed. This enables
// byte-identical dataset generation across JS and Go.
func mulberry32(seed int) func() float64 {
	s := uint32(int32(seed))
	return func() float64 {
		s += 0x6d2b79f5
		t := imul(s^(s>>15), 1|s)
		t = (t + imul(t^(t>>7), 61|t)) ^ t
		return float64(t^(t>>14)) / 4294967296.0
	}
}

// imul replicates JavaScript's Math.imul: 32-bit integer multiply with truncation.
func imul(a, b uint32) uint32 {
	return uint32(int32(a) * int32(b))
}
