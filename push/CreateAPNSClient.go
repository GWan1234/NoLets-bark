package push

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime"

	"github.com/sunvc/NoLets/common"
	"github.com/sunvc/apns2"
	"github.com/sunvc/apns2/token"
	"golang.org/x/net/context"
	"golang.org/x/net/http2"
)

var (
	CLIENTS = make(chan *apns2.Client, 1)
)

func CreateAPNSClient(maxClientCount int) {

	maxServerInstance := func() int {

		if maxClientCount <= 0 {
			return runtime.NumCPU()
		}

		return maxClientCount
	}()

	CLIENTS = make(chan *apns2.Client, maxServerInstance)

	authKey, err := token.AuthKeyFromBytes([]byte(common.LocalConfig.Apple.ApnsPrivateKey))
	if err != nil {
		log.Println(fmt.Sprintf("failed to create APNS auth key: %v", err))
	}

	var rootCAs *x509.CertPool

	system := func() string { return runtime.GOOS }()

	if system == "windows" {
		rootCAs = x509.NewCertPool()
	} else {
		rootCAs, err = x509.SystemCertPool()
		if err != nil {
			log.Println(fmt.Sprintf("failed to get rootCAs: %v", err))
		}
	}

	for _, ca := range common.ApnsCAs {
		rootCAs.AppendCertsFromPEM([]byte(ca))
	}

	for i := 0; i < maxServerInstance; i++ {
		CLIENTS <- &apns2.Client{
			Token: &token.Token{
				AuthKey: authKey,
				KeyID:   common.LocalConfig.Apple.KeyID,
				TeamID:  common.LocalConfig.Apple.TeamID,
			},
			HTTPClient: &http.Client{
				Transport: &http2.Transport{
					DialTLSContext:  DialTLSContext,
					TLSClientConfig: &tls.Config{RootCAs: rootCAs},
				},
				Timeout: apns2.HTTPClientTimeout,
			},
			Host: selectPushMode(),
		}
	}

	log.Println(fmt.Sprintf("init %d*%s apns client success...\n", maxServerInstance, selectPushMode()))

}

func selectPushMode() string {
	if common.LocalConfig.Apple.Develop {
		return apns2.HostDevelopment
	} else {
		return apns2.HostProduction
	}
}

func DialTLSContext(context context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {

	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{
			Timeout:   apns2.TLSDialTimeout,
			KeepAlive: apns2.TCPKeepAlive,
		},
		Config: cfg,
	}

	return dialer.DialContext(context, network, addr)

}
