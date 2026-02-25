package main

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
)

func SendOTPEmail(targetEmail string, otp string) {
	from := os.Getenv("SMTP_EMAIL")
	pass := os.Getenv("SMTP_PASS")

	if from == "" || pass == "" {
		fmt.Println("❌ SMTP credentials not set in environment")
		return
	}

	// Fix: Port 587 ki jagah 465 (SSL) use kar rahe hain kyunki Render 587 block karta hai
	smtpHost := "smtp.gmail.com"
	smtpPort := "465"

	auth := smtp.PlainAuth("", from, pass, smtpHost)

	subject := "Subject: Unity Student - OTP Verification\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	body := fmt.Sprintf(`
        <html>
            <body style="font-family: Arial, sans-serif; color: #333;">
                <div style="max-width: 600px; margin: auto; border: 1px solid #ddd; padding: 20px; border-radius: 10px;">
                    <h2 style="color: #7B2FF7; text-align: center;">Unity Student</h2>
                    <p>Hello,</p>
                    <p>Welcome to the community! Your OTP for account verification is:</p>
                    <div style="text-align: center; font-size: 24px; font-weight: bold; color: #7B2FF7; padding: 10px; background: #f4f4f4; border-radius: 5px;">
                        %s
                    </div>
                    <p style="margin-top: 20px;">This OTP is valid for 10 minutes. Please do not share this with anyone.</p>
                    <hr>
                    <p style="font-size: 12px; color: #888; text-align: center;">Verified by Unity Team</p>
                </div>
            </body>
        </html>`, otp)

	msg := []byte(subject + mime + body)

	// SSL Configuration for Port 465
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpHost,
	}

	// Port 465 ke liye direct dial karna padta hai
	conn, err := tls.Dial("tcp", smtpHost+":"+smtpPort, tlsconfig)
	if err != nil {
		fmt.Printf("❌ Connection error: %s\n", err)
		return
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		fmt.Printf("❌ SMTP Client error: %s\n", err)
		return
	}
	defer c.Quit()

	if err = c.Auth(auth); err != nil {
		fmt.Printf("❌ Auth error: %s\n", err)
		return
	}

	if err = c.Mail(from); err != nil {
		fmt.Printf("❌ Mail error: %s\n", err)
		return
	}

	if err = c.Rcpt(targetEmail); err != nil {
		fmt.Printf("❌ Rcpt error: %s\n", err)
		return
	}

	w, err := c.Data()
	if err != nil {
		fmt.Printf("❌ Data error: %s\n", err)
		return
	}

	_, err = w.Write(msg)
	if err != nil {
		fmt.Printf("❌ Write error: %s\n", err)
		return
	}

	err = w.Close()
	if err != nil {
		fmt.Printf("❌ Close error: %s\n", err)
		return
	}

	fmt.Printf("✅ OTP sent successfully to %s\n", targetEmail)
}
