package src

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"hash"
	"math/rand"
	"time"
)

//Taa carrier auth algorithm
type Taa struct {
	block cipher.Block
	mac   hash.Hash
	token TAuthToken
}

//GenToken svr generate token
func (aa *Taa) GenToken() {
	aa.token.challange = uint64(rand.Int63())
	aa.token.timestamp = uint64(time.Now().UnixNano())
}

//GenCipherBlock generate cipher block
func (aa *Taa) GenCipherBlock(token *TAuthToken) []byte {
	if token == nil {
		token = &aa.token
	}

	dst := make([]byte, TaaBlockSize)
	aa.block.Encrypt(dst, token.toBytes())
	aa.mac.Write(dst[:TaaTokenSize])
	sign := aa.mac.Sum(nil)
	aa.mac.Reset()

	copy(dst[TaaTokenSize:], sign)
	return dst
}

//CheckSignature check sig
func (aa *Taa) CheckSignature(src []byte) bool {
	aa.mac.Write(src[:TaaTokenSize])
	expectedMac := aa.mac.Sum(nil)
	aa.mac.Reset()
	return hmac.Equal(src[TaaTokenSize:], expectedMac)
}

//ExchangeCipherBlock exchange cipher block
func (aa *Taa) ExchangeCipherBlock(src []byte) ([]byte, bool) {
	if len(src) != TaaBlockSize {
		return nil, false
	}

	if !aa.CheckSignature(src) {
		return nil, false
	}

	dst := make([]byte, TaaTokenSize)
	aa.block.Decrypt(dst, src)
	(&aa.token).fromBytes(dst)

	// complement challenge
	token := aa.token.complement()
	return aa.GenCipherBlock(&token), true
}

//VerifyCipherBlock verify cipher block
func (aa *Taa) VerifyCipherBlock(src []byte) bool {
	if len(src) != TaaBlockSize {
		return false
	}

	if !aa.CheckSignature(src) {
		return false
	}

	var token TAuthToken
	dst := make([]byte, TaaTokenSize)
	aa.block.Decrypt(dst, src)
	(&token).fromBytes(dst)
	return aa.token.isComplemenary(token)
}

func (aa *Taa) getRc4key() []byte {
	return bytes.Repeat(aa.token.toBytes(), 8)
}

func newTaa(key string) *Taa {
	token := sha256.Sum256([]byte(key))
	block, _ := aes.NewCipher(token[:TaaTokenSize])
	mac := hmac.New(md5.New, token[TaaTokenSize:])
	aa := &Taa{
		block: block,
		mac:   mac,
	}
	return aa
}
