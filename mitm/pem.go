package mitm

import (
	"encoding/base64"
	"io"
)

// EncodeCertificatePEM writes a single certificate DER block as PEM.
func EncodeCertificatePEM(w io.Writer, der []byte) error {
	if _, err := w.Write([]byte("-----BEGIN CERTIFICATE-----\n")); err != nil {
		return err
	}
	b := make([]byte, base64.StdEncoding.EncodedLen(len(der)))
	base64.StdEncoding.Encode(b, der)
	for i := 0; i < len(b); i += 64 {
		j := i + 64
		if j > len(b) {
			j = len(b)
		}
		if _, err := w.Write(b[i:j]); err != nil {
			return err
		}
		if _, err := w.Write([]byte("\n")); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("-----END CERTIFICATE-----\n"))
	return err
}
