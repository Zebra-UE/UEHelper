package misc

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
	"strings"
)

// ...existing code...

// --- ECB 模式实现 (支持 cipher.BlockMode 接口) ---
type ecb struct {
	b cipher.Block
}

type ecbEncrypter ecb
type ecbDecrypter ecb

func newECB(b cipher.Block) *ecb { return &ecb{b: b} }

func NewECBEncrypter(b cipher.Block) cipher.BlockMode { return (*ecbEncrypter)(newECB(b)) }
func NewECBDecrypter(b cipher.Block) cipher.BlockMode { return (*ecbDecrypter)(newECB(b)) }

func (x *ecbEncrypter) BlockSize() int { return (*ecb)(x).b.BlockSize() }
func (x *ecbDecrypter) BlockSize() int { return (*ecb)(x).b.BlockSize() }

func (x *ecbEncrypter) CryptBlocks(dst, src []byte) {
	b := (*ecb)(x).b
	bs := b.BlockSize()
	if len(src)%bs != 0 {
		panic("ecb encrypter: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("ecb encrypter: dst too small")
	}
	for i := 0; i < len(src); i += bs {
		b.Encrypt(dst[i:i+bs], src[i:i+bs])
	}
}

func (x *ecbDecrypter) CryptBlocks(dst, src []byte) {
	b := (*ecb)(x).b
	bs := b.BlockSize()
	if len(src)%bs != 0 {
		panic("ecb decrypter: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("ecb decrypter: dst too small")
	}
	for i := 0; i < len(src); i += bs {
		b.Decrypt(dst[i:i+bs], src[i:i+bs])
	}
}

// --- PKCS7 填充/去填充 ---
type pkcs7 struct {
	blockSize int
}

func NewPkcs7Padding(blockSize int) *pkcs7 { return &pkcs7{blockSize: blockSize} }

func (p *pkcs7) Pad(data []byte) []byte {
	if p.blockSize <= 0 {
		return data
	}
	padLen := p.blockSize - (len(data) % p.blockSize)
	if padLen == 0 {
		padLen = p.blockSize
	}
	out := make([]byte, len(data)+padLen)
	copy(out, data)
	for i := len(data); i < len(out); i++ {
		out[i] = byte(padLen)
	}
	return out
}

func (p *pkcs7) Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("pkcs7: data is empty")
	}
	if len(data)%p.blockSize != 0 {
		return nil, fmt.Errorf("pkcs7: data is not a multiple of block size (%d)", p.blockSize)
	}
	last := int(data[len(data)-1])
	if last == 0 || last > p.blockSize {
		return nil, errors.New("pkcs7: invalid padding size")
	}
	for i := 0; i < last; i++ {
		if data[len(data)-1-i] != byte(last) {
			return nil, errors.New("pkcs7: invalid padding")
		}
	}
	return data[:len(data)-last], nil
}

func Decrypt(ct, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ct)%block.BlockSize() != 0 {
		return nil, fmt.Errorf("ciphertext length must be a multiple of block size (%d)", block.BlockSize())
	}

	mode := NewECBDecrypter(block)
	pt := make([]byte, len(ct))
	mode.CryptBlocks(pt, ct)

	padder := NewPkcs7Padding(block.BlockSize())
	unp, err := padder.Unpad(pt) // try unpad plaintext after decryption
	if err != nil {
		// Some PAK indexes are not PKCS7-padded — tolerate that and return raw plaintext.
		// Only swallow padding-related errors; propagate others.
		if strings.Contains(err.Error(), "pkcs7: invalid padding") ||
			strings.Contains(err.Error(), "pkcs7: invalid padding size") ||
			strings.Contains(err.Error(), "pkcs7: data is empty") ||
			strings.Contains(err.Error(), "pkcs7: data is not a multiple of block size") {
			return pt, nil
		}
		return nil, err
	}
	return unp, nil
}

// 添加的 AES-256 ECB 解密函数
func DecryptAES256ECB(ct, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("aes-256: key length must be 32 bytes, got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ct)%block.BlockSize() != 0 {
		return nil, fmt.Errorf("ciphertext length must be a multiple of block size (%d)", block.BlockSize())
	}

	mode := NewECBDecrypter(block)
	pt := make([]byte, len(ct))
	mode.CryptBlocks(pt, ct)

	padder := NewPkcs7Padding(block.BlockSize())
	unp, err := padder.Unpad(pt)
	if err != nil {
		if strings.Contains(err.Error(), "pkcs7: invalid padding") ||
			strings.Contains(err.Error(), "pkcs7: invalid padding size") ||
			strings.Contains(err.Error(), "pkcs7: data is empty") ||
			strings.Contains(err.Error(), "pkcs7: data is not a multiple of block size") {
			return pt, nil
		}
		return nil, err
	}
	return unp, nil
}
