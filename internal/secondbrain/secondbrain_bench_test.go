package secondbrain

import "testing"

func BenchmarkLoad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := Load("../../knowledge"); err != nil {
			b.Fatal(err)
		}
	}
}
