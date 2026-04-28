package api

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/user/mailadmin/internal/auth"
)

type DiagnosticResult struct {
	Domain      string          `json:"domain"`
	MX          []string        `json:"mx"`
	SPF         SPFResult       `json:"spf"`
	DMARC       DMARCResult     `json:"dmarc"`
	DKIM        []DKIMResult    `json:"dkim"`
	SSL         []SSLResult     `json:"ssl"`
	RBL         []RBLResult     `json:"rbl"`
	OverallStatus string        `json:"overall_status"`
}

type SPFResult struct {
	Value  string `json:"value"`
	Valid  bool   `json:"valid"`
	Status string `json:"status"`
}

type DMARCResult struct {
	Value  string `json:"value"`
	Valid  bool   `json:"valid"`
	Status string `json:"status"`
}

type DKIMResult struct {
	Selector string `json:"selector"`
	Value    string `json:"value"`
	Valid    bool   `json:"valid"`
}

type SSLResult struct {
	Host       string    `json:"host"`
	Port       int       `json:"port"`
	Valid      bool      `json:"valid"`
	Expires    time.Time `json:"expires"`
	DaysLeft   int       `json:"days_left"`
	Issuer     string    `json:"issuer"`
	Error      string    `json:"error,omitempty"`
}

type RBLResult struct {
	Server  string `json:"server"`
	Listed  bool   `json:"listed"`
	Message string `json:"message,omitempty"`
}

func RegisterDiagnosticsHandlers(g *echo.Group, secret string) {
	g.Use(auth.JWTMiddleware(secret))

	g.GET("/:domain", func(c echo.Context) error {
		domain := c.Param("domain")

		// CRIT-3: проверка доступа — только владелец домена или суперадмин
		claims := c.Get("user").(*auth.Claims)
		if !claims.SuperAdmin && !hasDomainAccess(claims, domain) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
		}

		result := DiagnosticResult{
			Domain: domain,
		}

		// 1. MX Records
		mxs, _ := net.LookupMX(domain)
		for _, mx := range mxs {
			result.MX = append(result.MX, mx.Host)
		}

		// 2. SPF
		txts, _ := net.LookupTXT(domain)
		for _, txt := range txts {
			if strings.HasPrefix(txt, "v=spf1") {
				result.SPF.Value = txt
				result.SPF.Valid = true
				result.SPF.Status = "Found"
				break
			}
		}
		if !result.SPF.Valid {
			result.SPF.Status = "Not found"
		}

		// 3. DMARC
		dmarctxts, _ := net.LookupTXT("_dmarc." + domain)
		for _, txt := range dmarctxts {
			if strings.HasPrefix(txt, "v=DMARC1") {
				result.DMARC.Value = txt
				result.DMARC.Valid = true
				result.DMARC.Status = "Found"
				break
			}
		}
		if !result.DMARC.Valid {
			result.DMARC.Status = "Not found"
		}

		// 4. DKIM (проверяем популярные селекторы)
		selectors := []string{"default", "mail", "postfix", "google"}
		for _, s := range selectors {
			dkimtxts, _ := net.LookupTXT(s + "._domainkey." + domain)
			for _, txt := range dkimtxts {
				if strings.Contains(txt, "v=DKIM1") || strings.Contains(txt, "k=rsa") {
					result.DKIM = append(result.DKIM, DKIMResult{
						Selector: s,
						Value:    txt,
						Valid:    true,
					})
				}
			}
		}

		// 5. SSL Check
		hostsToCheck := []string{"mail." + domain, domain}
		ports := []int{443, 993, 465}
		for _, host := range hostsToCheck {
			for _, port := range ports {
				sslRes := checkSSL(host, port)
				if sslRes.Error == "" || sslRes.Valid {
					result.SSL = append(result.SSL, sslRes)
				}
			}
		}

		// 6. RBL Check (проверяем IP первого MX-сервера)
		if len(result.MX) > 0 {
			mxHost := strings.TrimSuffix(result.MX[0], ".")
			ips, _ := net.LookupIP(mxHost)
			if len(ips) > 0 {
				ip := ips[0].String()
				result.RBL = checkRBL(ip)
			}
		}

		return c.JSON(http.StatusOK, result)
	})
}

func checkSSL(host string, port int) SSLResult {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 3 * time.Second}, "tcp", address, &tls.Config{
		InsecureSkipVerify: true, // Мы хотим проверить сам сертификат, даже если он просрочен
	})
	if err != nil {
		return SSLResult{Host: host, Port: port, Valid: false, Error: err.Error()}
	}
	defer conn.Close()

	cert := conn.ConnectionState().PeerCertificates[0]
	daysLeft := int(time.Until(cert.NotAfter).Hours() / 24)
	
	return SSLResult{
		Host:     host,
		Port:     port,
		Valid:    time.Now().Before(cert.NotAfter) && time.Now().After(cert.NotBefore),
		Expires:  cert.NotAfter,
		DaysLeft: daysLeft,
		Issuer:   cert.Issuer.CommonName,
	}
}

func checkRBL(ip string) []RBLResult {
	rbls := []string{
		"zen.spamhaus.org",
		"bl.spamcop.net",
		"dnsbl.sorbs.net",
	}

	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return nil
	}
	reverseIP := fmt.Sprintf("%s.%s.%s.%s", parts[3], parts[2], parts[1], parts[0])

	var results []RBLResult
	for _, rbl := range rbls {
		_, err := net.LookupHost(reverseIP + "." + rbl)
		results = append(results, RBLResult{
			Server: rbl,
			Listed: err == nil,
		})
	}
	return results
}
