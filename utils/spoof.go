package utils

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"sort"
	"strings"
)

func SendSpoofedEmail(email SpoofEmail) (bool, error) {
	mxRecords, err := net.LookupMX(email.TargetDomain)
	if err != nil || len(mxRecords) == 0 {
		log.Fatalf("MX lookup failed for domain %s: %v", email.TargetDomain, err)
	}

	sort.Slice(mxRecords, func(i, j int) bool {
		return mxRecords[i].Pref < mxRecords[j].Pref
	})
	mxHost := mxRecords[0].Host

	fmt.Printf("Connecting to MX %s:25 ...\n", mxHost[:len(mxHost)-1])

	conn, err := net.Dial("tcp", mxHost+":25")
	if err != nil {
		return false, fmt.Errorf("failed to connect to %s:25: %v", mxHost, err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, mxHost)
	if err != nil {
		return false, fmt.Errorf("SMTP handshake failed: %v", err)
	}
	defer client.Quit()

	// 4) Envelope from: **must** be a valid, resolvable domain
	if err := client.Mail(email.From); err != nil {
		return false, fmt.Errorf("MAIL FROM error: %v", err)
	}
	if err := client.Rcpt(email.To); err != nil {
		return false, fmt.Errorf("RCPT TO error: %v", err)
	}

	wc, err := client.Data()
	if err != nil {
		log.Fatalf("DATA command error: %v", err)
	}
	defer wc.Close()

	msg := []string{
		fmt.Sprintf("From: %s", email.From),
		fmt.Sprintf("To: %s", email.To),
		fmt.Sprintf("Subject: %s", email.Subject),
		"MIME-VERSION: 1.0",
		fmt.Sprintf("Content-Type: %s; charset=\"UTF-8\"", email.Mimetype),
		"",
		email.Body,
	}
	raw := strings.Join(msg, "\r\n")

	writer := bufio.NewWriter(wc)
	if _, err := writer.WriteString(raw + "\r\n"); err != nil {
		log.Fatalf("Writing message failed: %v", err)
	}
	writer.Flush()

	return true, nil
}
