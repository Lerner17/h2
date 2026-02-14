package passwordgen

import (
	"crypto/rand"
	"fmt"
)

const generatedPasswordLength = 32

var (
	lower = "abcdefghijklmnopqrstuvwxyz"
	upper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digit = "0123456789"
	all   = lower + upper + digit
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Generate() (string, error) {
	buf := make([]byte, generatedPasswordLength)
	for i := range buf {
		idx, err := secureInt(len(all))
		if err != nil {
			return "", err
		}
		buf[i] = all[idx]
	}

	if err := forceClass(buf, 0, lower); err != nil {
		return "", err
	}
	if err := forceClass(buf, 1, upper); err != nil {
		return "", err
	}
	if err := forceClass(buf, 2, digit); err != nil {
		return "", err
	}

	return string(buf), nil
}

func forceClass(buf []byte, pos int, charset string) error {
	idx, err := secureInt(len(charset))
	if err != nil {
		return err
	}
	buf[pos] = charset[idx]
	return nil
}

func secureInt(max int) (int, error) {
	if max <= 0 {
		return 0, fmt.Errorf("invalid max: %d", max)
	}
	var b [1]byte
	limit := byte(256 - (256 % max))
	for {
		if _, err := rand.Read(b[:]); err != nil {
			return 0, err
		}
		if b[0] < limit {
			return int(b[0]) % max, nil
		}
	}
}
