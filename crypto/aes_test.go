package crypto

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestAESEncrypt(t *testing.T) {
	v := "123456"

	cipher, err := Encrypt(v)
	if err != nil {
		t.Fatal(err)
		return
	}

	t.Log("Cipher: ", cipher)

	decodedValue, err := Decrypt(cipher)
	if err != nil {
		t.Fatal(err)
		return
	}

	t.Log("Decoded Value: ", decodedValue)

	assert.Equal(t, cipher, decodedValue)
}
