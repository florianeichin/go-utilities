package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"

	"github.com/moby/buildkit/utilities"
)

// NewEncryptionWriter creates a new writer encapsulating a writer
func NewEncryptionWriter(w io.Writer, key string) (io.WriteCloser, error) {
	utilities.NewPrinter("DEBUG", "", "encrypting...")
	k, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}
	c, err := aes.NewCipher(k)
	if err != nil {
		return nil, err
	}
	iv := make([]byte, aes.BlockSize)
	return &cipher.StreamWriter{S: cipher.NewOFB(c, iv), W: w}, nil
}

// NewDecryptionReader creates a new reader encapsulating a reader
func NewDecryptionReader(r io.Reader, key string) (io.Reader, error) {
	utilities.NewPrinter("DEBUG", "", "decrypting...")
	k, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}
	c, err := aes.NewCipher(k)
	if err != nil {
		return nil, err
	}
	iv := make([]byte, aes.BlockSize)
	return &cipher.StreamReader{S: cipher.NewOFB(c, iv), R: r}, nil
}

// NewDecryptionReadCloser wraps an DecryptionReader to apply to ReadClosers
func NewDecryptionReadCloser(r io.Reader, passphrase string) (io.ReadCloser, error) {
	rdr, err := NewDecryptionReader(r, passphrase)
	if err != nil {
		return nil, err
	}
	return &ReadCloser{rdr}, nil
}

type ReadCloser struct {
	io.Reader
}

func (rc *ReadCloser) Close() error {
	return nil
}

type WriteCloser struct {
	io.Writer
}

func (rc *WriteCloser) Close() error {
	return nil
}

func Encrypt(data []byte, passphrase string) ([]byte, error) {
	key, err := hex.DecodeString(passphrase)
	if err != nil {
		return nil, err
	}
	c, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(c)
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

func Decrypt(data []byte, passphrase string) ([]byte, error) {
	key, err := hex.DecodeString(passphrase)
	if err != nil {
		return nil, err
	}
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	nonce, data := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, data, nil)
}

func LoadKey(filePath, passphrase string) (string, error) {
	fmt.Println("loading key...")
	fmt.Println("filepath:", filePath)
	key, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	passphrase, err = hashPassphrase(passphrase)
	if err != nil {
		return "", err
	}
	key, err = Decrypt(key, passphrase)
	if err != nil {
		return "", err
	}
	return string(key[:]), nil
}

func LoadKeyByHash(hash string, passphrase string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	folder := usr.HomeDir + "/.encryption"

	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return "", err
	}

	for _, f := range files {
		filePath := folder + "/" + f.Name()
		key, err := os.Open(filePath)
		if err != nil {
			return "", err
		}
		defer key.Close()
		h := sha256.New()
		if _, err := io.Copy(h, key); err != nil {
			return "", err
		}
		if hash == string(hex.EncodeToString(h.Sum(nil)[:])) {
			return LoadKey(filePath, passphrase)
		}

	}

	return "", errors.New("No Key found with given digest")
}

func GenerateKey(passphrase string) (string, error) {
	fmt.Println("generating key...")
	key := make([]byte, 32)
	_, err := rand.Read(key)
	key = []byte(hex.EncodeToString(key))
	if err != nil {
		return "", err
	}
	passphrase, err = hashPassphrase(passphrase)
	if err != nil {
		return "", err
	}
	key, err = Encrypt(key, passphrase)

	h := sha256.New()
	if _, err := h.Write(key); err != nil {
		return "", err
	}
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	filePath := usr.HomeDir + "/.encryption/" + string(hex.EncodeToString(h.Sum(nil)[:]))
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = file.Write(key)
	if err != nil {
		return "", err
	}

	return string(hex.EncodeToString(h.Sum(nil)[:])), nil
}

func hashPassphrase(passphrase string) (string, error) {
	hasher := sha256.New()
	_, err := hasher.Write([]byte(passphrase))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil

}
