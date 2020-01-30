package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
)

const AESGCM_NONCE_SIZE = 12
const AESGCM_KEY_SIZE = 16

type Crypter struct {
	ciph      cipher.AEAD
	nonceSize int
}

func NewCrypter(keystring string) (Crypter, error) {
	var c Crypter
	c.nonceSize = AESGCM_NONCE_SIZE

	key, err := hex.DecodeString(keystring)
	if err != nil {
		return c, fmt.Errorf("Unable to load key: %v", err)
	}

	if len(key) != AESGCM_KEY_SIZE {
		return c, fmt.Errorf("Invalid key length, must be %d bytes", AESGCM_KEY_SIZE)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		// Because we have already checked the validity of the key, this should never occur
		panic(err.Error())
	}

	c.ciph, err = cipher.NewGCM(block)
	if err != nil {
		// I don't know how this could fail
		panic(err.Error())
	}

	return c, nil
}

// Encrypt will encrypt a given plaintext using AES in GCM mode
// The 12 byte nonce is prepended to the ciphertext
func (c Crypter) Encrypt(plaintext []byte) []byte {
	nonce := make([]byte, c.nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error()) // Unable to read random bytes into nonce
	}

	return c.ciph.Seal(nonce, nonce, plaintext, nil)
}

func (c Crypter) Decrypt(toDecrypt []byte) ([]byte, error) {

	if len(toDecrypt) < c.nonceSize {
		return nil, fmt.Errorf("Decryption failed, input too short")
	}

	ciphertext := toDecrypt[c.nonceSize:]
	nonce := toDecrypt[:c.nonceSize]

	return c.ciph.Open(nil, nonce, ciphertext, nil)
}

/* GenerateTag creates a new candidateTag string which stores an encrypted
sessionId and candidateId. The info is packed into 6 bytes in the following manner:

  byte 1    byte 2    byte 3    byte 4    byte 5    byte 6
+---------+---------+---------+---------+---------+---------+
|              sessionId                |    candidateId    |
+---------+---------+---------+---------+---------+---------+

 GenerateTag(0xdeadbeef, 3) becomes Encrypt(0xdeadbeef0003) */
func (c Crypter) GenerateTag(sessionId uint32, candidateId int) string {
	b := make([]byte, 6)
	binary.BigEndian.PutUint32(b[0:], sessionId)
	binary.BigEndian.PutUint16(b[4:], uint16(candidateId))

	e := c.Encrypt(b)
	return fmt.Sprintf("%x", e)
}

func (c Crypter) DecodeTag(cipherstring string, sessionId uint32) (int, error) {
	cipherbytes, err := hex.DecodeString(cipherstring)
	if err != nil {
		return 0, fmt.Errorf("Unable to decode hex string: %v", err)
	}

	plainbytes, err := c.Decrypt(cipherbytes)
	if err != nil {
		return 0, fmt.Errorf("Could not decrypt, %v", err)
	}

	if len(plainbytes) != 6 {
		return 0, fmt.Errorf("Plaintext invalid length")
	}

	if sessionId != binary.BigEndian.Uint32(plainbytes[:4]) {
		return 0, fmt.Errorf("SessionId did not match")
	}

	return int(binary.BigEndian.Uint16(plainbytes[4:6])), nil
}

func main() {

	c, err := NewCrypter("0dcdf3dd39e1f81f6ed0adc1c31ab40a")
	if err != nil {
		panic(err.Error())
	}

	/*
	   secret, _ := hex.DecodeString("012345670003")
	   ciphertext := c.Encrypt(secret)
	   fmt.Printf("%x\n",ciphertext)

	   decoded, _ := c.Decrypt(ciphertext)
	   fmt.Printf("%x\n",decoded)

	*/
	var sessionId uint32 = 0xdeadbeef
	tag := c.GenerateTag(sessionId, 45)

	fmt.Println(tag)

	decode, err := c.DecodeTag(tag, sessionId)
	if err != nil {
		fmt.Println("ERROR:", err)
	} else {
		fmt.Println("DECODE:", decode)
	}
}
