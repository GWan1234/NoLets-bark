package controller

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/sunvc/NoLets/common"
)

func CreateSSL() {
	keyPath := common.BaseDir("key.pem")
	certPath := common.BaseDir("cert.pem")

	// 如果证书已存在则跳过
	if _, err := os.Stat(certPath); err == nil {
		log.Printf("CreateSSL: cert already exists at %s, skip generation", certPath)
		common.LocalConfig.System.Key = keyPath
		common.LocalConfig.System.Cert = certPath
		return
	}

	// 生成ECDSA私钥
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	// 设置证书模板
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		panic(err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Country:      []string{"CN"},
			Province:     []string{"Beijing"},
			Locality:     []string{"Beijing"},
			Organization: []string{"NoLetter Inc."},
			CommonName:   "wzs.app",
			SerialNumber: "001",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(5 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"*"},
		EmailAddresses:        []string{"to@wzs.app"},
	}

	// 自签名证书
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		panic(err)
	}

	// 将证书和私钥写入文件
	certOut, err := os.Create(certPath)
	if err != nil {
		panic(err)
	}
	defer certOut.Close()

	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	keyOut, err := os.Create(keyPath)
	if err != nil {
		panic(err)
	}
	defer keyOut.Close()

	marshaledKey, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		panic(err)
	}

	pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: marshaledKey})

	println("生成自签名TLS证书成功")

	common.LocalConfig.System.Key = keyPath
	common.LocalConfig.System.Cert = certPath
}
