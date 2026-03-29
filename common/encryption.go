package common

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"golang.org/x/crypto/nacl/secretbox"
	"log/slog"
)

const (
	keyLength   = 32
	nonceLength = 24
)

type encryptorPayload struct {
	Key     []byte
	Nonce   *[nonceLength]byte
	Message []byte
}

type EnvelopEncrypt struct {
	kmsClient *kms.Client
}

var envelopEncrypter *EnvelopEncrypt = nil
var awsConfig aws.Config

func InitEnvelopEncryptorConfigure(config aws.Config) {
	awsConfig = config
}

func getEnvelopEncrypt() *EnvelopEncrypt {
	if envelopEncrypter == nil {

		if awsConfig.Region == "" {
			cfg, err := config.LoadDefaultConfig(context.Background())
			if err != nil {
				slog.Error("can't load AWS configuration/credentials. This service encrypts and decrypts, we assume that is better to panic than to keep working")
				panic(err)
			}
			awsConfig = cfg
		}
		envelopEncrypter = &EnvelopEncrypt{
			kmsClient: kms.NewFromConfig(awsConfig),
		}
	}

	return envelopEncrypter
}

// EnvelopeEncrypt encrypt plain []byte using the AWS KMS keyId arn
func EnvelopeEncrypt(ctx context.Context, keyId string, plain []byte) ([]byte, error) {
	dataKeyInput := kms.GenerateDataKeyInput{
		KeyId:   &keyId,
		KeySpec: types.DataKeySpecAes256,
	}
	encryptor := getEnvelopEncrypt()

	dataKeyOutput, err := encryptor.kmsClient.GenerateDataKey(ctx, &dataKeyInput)
	if err != nil {
		return nil, err
	}

	p := &encryptorPayload{
		Key:   dataKeyOutput.CiphertextBlob,
		Nonce: &[nonceLength]byte{},
	}

	if _, err = rand.Read(p.Nonce[:]); err != nil {
		return nil, err
	}

	key := &[keyLength]byte{}
	copy(key[:], dataKeyOutput.Plaintext)

	p.Message = secretbox.Seal(p.Message, plain, p.Nonce, key)

	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(p); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func EnvelopeDecrypt(ctx context.Context, encrypted []byte) ([]byte, error) {
	var p encryptorPayload
	err := gob.NewDecoder(bytes.NewReader(encrypted)).Decode(&p)
	if err != nil {
		return nil, err
	}

	encryptor := getEnvelopEncrypt()

	// Decrypt a ciphertext that was previously encrypted.
	// Note that we don't actually specify the key name,
	// because the data key (kms and version) is stored with the data
	dataKeyOutput, err := encryptor.kmsClient.Decrypt(ctx, &kms.DecryptInput{
		CiphertextBlob: p.Key,
	})
	if err != nil {
		return nil, err
	}

	key := &[keyLength]byte{}
	copy(key[:], dataKeyOutput.Plaintext)

	var plaintext []byte
	plaintext, ok := secretbox.Open(plaintext, p.Message, p.Nonce, key)
	if !ok {
		return nil, fmt.Errorf("failed to open secretbox")
	}
	return plaintext, nil
}
