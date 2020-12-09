package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"log"
	"math/big"
	"os"
	"reflect"
	"time"
)

// EncodeBase64 converts bytes to base 64
func EncodeBase64(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

// DecodeBase64 converts base 64 to bytes
func DecodeBase64(input string) []byte {
	output, _ := base64.StdEncoding.DecodeString(input)
	return output
}

// RemoveDevAddr removes the address at the given index
func RemoveDevAddr(slice []uint16, i int) []uint16 {
	return append(slice[:i], slice[i+1:]...)
}

// Delete removes the element at the given index
func Delete(arr interface{}, index int) {
	vField := reflect.ValueOf(arr)
	value := vField.Elem()
	if value.Kind() == reflect.Slice || value.Kind() == reflect.Array {
		result := reflect.AppendSlice(value.Slice(0, index), value.Slice(index+1, value.Len()))
		value.Set(result)
	}
}

// WriteCert writes a new cert to the application dir
func WriteCert() {
	// Make private key
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	// Set key usage
	keyUsage := x509.KeyUsageDigitalSignature
	// Set time limits
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	// Make serial number
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, serialNumberLimit)
	// Create cert template
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Home"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	// Set ca to true
	template.IsCA = true
	template.KeyUsage |= x509.KeyUsageCertSign
	// Create the cert
	derBytes, _ := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	// Write cert to file
	certOut, _ := os.Create("/app/cert.pem")
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()
	log.Print("wrote cert.pem\n")
	// Write key to file
	keyOut, _ := os.OpenFile("/app/key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	privBytes, _ := x509.MarshalPKCS8PrivateKey(priv)
	pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	keyOut.Close()
	log.Print("wrote key.pem\n")
}

// CheckIfConfigured returns a bool indacating if the hub is configured
func CheckIfConfigured() bool {
	if _, err := os.Stat("home.data"); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
