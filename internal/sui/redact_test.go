package sui

import (
	"strings"
	"testing"
)

func TestRedactSecret(t *testing.T) {
	mnemonic := "abandon ability able about above absent absorb abstract absurd abuse access accident"
	tests := []struct {
		name, out, secret string
		leaked            []string
	}{
		{"private key", "Error: invalid key suiprivkey1qqxyz", "suiprivkey1qqxyz", []string{"suiprivkey1qqxyz"}},
		{"whole mnemonic", "Error: bad phrase: " + mnemonic, mnemonic, strings.Fields(mnemonic)},
		{"partial mnemonic", "word 'absurd' at 9 not in list near able", mnemonic, []string{"absurd", "able"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redactSecret(tt.out, tt.secret)
			for _, w := range tt.leaked {
				if strings.Contains(got, w) {
					t.Errorf("redactSecret leaked %q: %q", w, got)
				}
			}
		})
	}

	for _, secret := range []string{"", mnemonic} {
		if got := redactSecret("Error: invalid input", secret); got != "Error: invalid input" {
			t.Errorf("redactSecret(_, %q) mangled unrelated text: %q", secret, got)
		}
	}
}
