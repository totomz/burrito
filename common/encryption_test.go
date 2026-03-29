package common

import (
	"context"
	"encoding/base64"
	"testing"
)

func TestEnvelopeEncryption(t *testing.T) {
	t.Skip()
	// valueEncryptedOld is value but encrypted with a kms that has been rotated
	valueEncryptedOld := "PH8DAQEQZW5jcnlwdG9yUGF5bG9hZAH/gAABAwEDS2V5AQoAAQVOb25jZQH/ggABB01lc3NhZ2UBCgAAABn/gQEBAQlbMjRddWludDgB/4IAAQYBMAAA/gES/4AB/7gBAgMAePa3CyybcoqZBTAgbZDHPcr2HCoqFtVP60g8BS2Qx60IARtUQ09PTrmaLV98Ye08TXMAAAB+MHwGCSqGSIb3DQEHBqBvMG0CAQAwaAYJKoZIhvcNAQcBMB4GCWCGSAFlAwQBLjARBAzpvhZFFfmhCt26/OkCARCAO55tafZnL5OHKO31jCektNXo3FHSVtBfSvZAZZTV0NAU6ANyoy8oW0z0i2o8sA3Y0bGwsfohTt2yQDKVARj/xhL/3P+DX//oWP/A/8n/yS1J/8daQhl7CGlrFP/THjkBLx50XDBZFUCNU5uwzCZv8lIT3x7HUFuYJl2oL/RDzV86RGVNTueBuY4jGY/gDYLpAA=="
	value := "vola nel blu una rondinella blu"
	keyId := "arn:aws:kms:eu-west-1:767398121280:key/6b9171ab-b8f3-4bd6-9e94-d6609d020ffc" // hack
	ctx := context.Background()

	encrypted, err := EnvelopeEncrypt(ctx, keyId, []byte(value))
	if err != nil {
		t.Fatal(err)
	}

	decrypted, err := EnvelopeDecrypt(ctx, encrypted)
	if err != nil {
		t.Fatal(err)
	}

	if string(decrypted) != value {
		t.Errorf("got %q, want %q", decrypted, value)
	}

	encryptedOld, err := base64.StdEncoding.DecodeString(valueEncryptedOld)
	if err != nil {
		t.Fatal(err)
	}

	decryptedOld, err := EnvelopeDecrypt(ctx, encryptedOld)
	if err != nil {
		t.Fatal(err)
	}

	if string(decryptedOld) != value {
		t.Errorf("KMS roated key got %q, want %q", decrypted, value)
	}
}
