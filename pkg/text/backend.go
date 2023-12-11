package text

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"
	"github.com/segmentio/textio"
)

// PrintBackend pretty prints a fastly.Backend structure in verbose format
// to a given io.Writer. Consumers can provide a prefix string which will
// be used as a prefix to each line, useful for indentation.
func PrintBackend(out io.Writer, prefix string, b *fastly.Backend) {
	out = textio.NewPrefixWriter(out, prefix)

	fmt.Fprintf(out, "Name: %s\n", fastly.ToValue(b.Name))
	fmt.Fprintf(out, "Comment: %v\n", fastly.ToValue(b.Comment))
	fmt.Fprintf(out, "Address: %v\n", fastly.ToValue(b.Address))
	fmt.Fprintf(out, "Port: %v\n", fastly.ToValue(b.Port))
	fmt.Fprintf(out, "Override host: %v\n", fastly.ToValue(b.OverrideHost))
	fmt.Fprintf(out, "Connect timeout: %v\n", fastly.ToValue(b.ConnectTimeout))
	fmt.Fprintf(out, "Max connections: %v\n", fastly.ToValue(b.MaxConn))
	fmt.Fprintf(out, "First byte timeout: %v\n", fastly.ToValue(b.FirstByteTimeout))
	fmt.Fprintf(out, "Between bytes timeout: %v\n", fastly.ToValue(b.BetweenBytesTimeout))
	fmt.Fprintf(out, "Auto loadbalance: %v\n", fastly.ToValue(b.AutoLoadbalance))
	fmt.Fprintf(out, "Weight: %v\n", fastly.ToValue(b.Weight))
	fmt.Fprintf(out, "Healthcheck: %v\n", fastly.ToValue(b.HealthCheck))
	fmt.Fprintf(out, "Shield: %v\n", fastly.ToValue(b.Shield))
	fmt.Fprintf(out, "Use SSL: %v\n", fastly.ToValue(b.UseSSL))
	fmt.Fprintf(out, "SSL check cert: %v\n", fastly.ToValue(b.SSLCheckCert))
	fmt.Fprintf(out, "SSL CA cert: %v\n", fastly.ToValue(b.SSLCACert))
	fmt.Fprintf(out, "SSL client cert: %v\n", fastly.ToValue(b.SSLClientCert))
	fmt.Fprintf(out, "SSL client key: %v\n", fastly.ToValue(b.SSLClientKey))
	fmt.Fprintf(out, "SSL cert hostname: %v\n", fastly.ToValue(b.SSLCertHostname))
	fmt.Fprintf(out, "SSL SNI hostname: %v\n", fastly.ToValue(b.SSLSNIHostname))
	fmt.Fprintf(out, "Min TLS version: %v\n", fastly.ToValue(b.MinTLSVersion))
	fmt.Fprintf(out, "Max TLS version: %v\n", fastly.ToValue(b.MaxTLSVersion))
	fmt.Fprintf(out, "SSL ciphers: %v\n", fastly.ToValue(b.SSLCiphers))
}
