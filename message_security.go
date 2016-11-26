//微信消息的加密解密

package wechat

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
)

//生成消息签名，通常会传入：token, timestamp, nonce, msg_encrypt 这几个参数
func MakeMsgSignature(sortParams ...string) string {
	sort.Strings(sortParams)
	s := sha1.New()
	io.WriteString(s, strings.Join(sortParams, ""))
	return fmt.Sprintf("%x", s.Sum(nil))
}

//解密微信消息的密文，得到明文数据
func DecryptMsg(body string, encodingKey string, appId string) (string, error) {
	var plainText, cipherData []byte
	var msgStr string
	//解密秘钥
	aesKey, _ := base64.StdEncoding.DecodeString(encodingKey + "=")
	//解密消息体
	cipherData, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		return msgStr, err
	}
	plainText, err = aesDecrypt(cipherData, aesKey)
	if err != nil {
		return msgStr, err
	}
	//解密出来的消息，做校验
	return parseEncryptTextRequestBody(plainText, appId)
}

//微信ase方式解密封装
func aesDecrypt(cipherData []byte, aesKey []byte) ([]byte, error) {
	k := len(aesKey) //PKCS#7
	if len(cipherData)%k != 0 {
		return nil, errors.New("crypto/cipher: ciphertext size is not multiple of aes key length")
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	blockMode := cipher.NewCBCDecrypter(block, iv)
	plainData := make([]byte, len(cipherData))
	blockMode.CryptBlocks(plainData, cipherData)
	return plainData, nil
}

//微信密文解明文
func parseEncryptTextRequestBody(plainText []byte, appId string) (string, error) {
	// Read length
	buf := bytes.NewBuffer(plainText[16:20])
	var length int32
	binary.Read(buf, binary.BigEndian, &length)
	// appID validation
	appIDstart := 20 + length
	id := plainText[appIDstart : int(appIDstart)+len(appId)]
	if string(id) != appId {
		return "", errors.New("Appid is invalid")
	}
	return string(plainText[20 : 20+length]), nil
}

//把明文的XML的byte数据，加密成密文版本的XML数据
func MakeEncryptXmlData(body []byte, appId, aesKey []byte) (string, error) {
	var err error
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, int32(len(body)))
	if err != nil {
		return "", err
	}
	bodyLength := buf.Bytes()

	// Encrypt part1: Random bytes
	randomBytes := []byte("abcdefghijklmnop")

	// Encrypt Part, with part4 - appID
	plainData := bytes.Join([][]byte{randomBytes, bodyLength, body, appId}, nil)
	cipherData, err := aesEncrypt(plainData, aesKey)
	if err != nil {
		return "", errors.New("aesEncrypt error")
	}

	return base64.StdEncoding.EncodeToString(cipherData), nil
}

//微信ase方式加密封装
func aesEncrypt(plainData []byte, aesKey []byte) ([]byte, error) {
	k := len(aesKey)
	if len(plainData)%k != 0 {
		plainData = PKCS7Pad(plainData, k)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	cipherData := make([]byte, len(plainData))
	blockMode := cipher.NewCBCEncrypter(block, iv)
	blockMode.CryptBlocks(cipherData, plainData)

	return cipherData, nil
}

func PKCS7Pad(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

//将微信的EncodingAESKey，解码。因为实际上微信的43位的EncodingAESKey，是编码过的。
func encodingAESKey2AESKey(encodingKey string) []byte {
	data, _ := base64.StdEncoding.DecodeString(encodingKey + "=")
	return data
}
