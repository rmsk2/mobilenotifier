package tools

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
)

const EnvAdditionalRootCerts = "MN_ADDITIONAL_ROOTS"

func MakeCustomHttpClient() (*http.Client, error) {
	rootFilename, ok := os.LookupEnv(EnvAdditionalRootCerts)
	if !ok {
		return http.DefaultClient, nil
	}

	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	certs, err := os.ReadFile(rootFilename)
	if err != nil {
		return nil, fmt.Errorf("Unable to load additional root certs from '%s'", rootFilename)
	}

	ok = rootCAs.AppendCertsFromPEM(certs)
	if !ok {
		return nil, fmt.Errorf("No certs appended, using system certs only")
	}

	tlsConfig := &tls.Config{
		RootCAs: rootCAs,
	}

	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client := &http.Client{
		Transport: tr,
	}

	return client, nil
}
