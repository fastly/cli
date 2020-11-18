package text

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v2/fastly"
	"github.com/segmentio/textio"
)

// PrintBackend pretty prints a fastly.Backend structure in verbose format
// to a given io.Writer. Consumers can provider an prefix string which will
// be used as a prefix to each line, useful for indentation.
func PrintBackend(out io.Writer, prefix string, b *fastly.Backend) {
	out = textio.NewPrefixWriter(out, prefix)

	fmt.Fprintf(out, "Name: %s\n", b.Name)
	fmt.Fprintf(out, "Comment: %v\n", b.Comment)
	fmt.Fprintf(out, "Address: %v\n", b.Address)
	fmt.Fprintf(out, "Port: %v\n", b.Port)
	fmt.Fprintf(out, "Override host: %v\n", b.OverrideHost)
	fmt.Fprintf(out, "Connect timeout: %v\n", b.ConnectTimeout)
	fmt.Fprintf(out, "Max connections: %v\n", b.MaxConn)
	fmt.Fprintf(out, "First byte timeout: %v\n", b.FirstByteTimeout)
	fmt.Fprintf(out, "Between bytes timeout: %v\n", b.BetweenBytesTimeout)
	fmt.Fprintf(out, "Auto loadbalance: %v\n", b.AutoLoadbalance)
	fmt.Fprintf(out, "Weight: %v\n", b.Weight)
	fmt.Fprintf(out, "Healthcheck: %v\n", b.HealthCheck)
	fmt.Fprintf(out, "Shield: %v\n", b.Shield)
	fmt.Fprintf(out, "Use SSL: %v\n", b.UseSSL)
	fmt.Fprintf(out, "SSL check cert: %v\n", b.SSLCheckCert)
	fmt.Fprintf(out, "SSL CA cert: %v\n", b.SSLCACert)
	fmt.Fprintf(out, "SSL client cert: %v\n", b.SSLClientCert)
	fmt.Fprintf(out, "SSL client key: %v\n", b.SSLClientKey)
	fmt.Fprintf(out, "SSL cert hostname: %v\n", b.SSLCertHostname)
	fmt.Fprintf(out, "SSL SNI hostname: %v\n", b.SSLSNIHostname)
	fmt.Fprintf(out, "Min TLS version: %v\n", b.MinTLSVersion)
	fmt.Fprintf(out, "Max TLS version: %v\n", b.MaxTLSVersion)
	fmt.Fprintf(out, "SSL ciphers: %v\n", b.SSLCiphers)
}
