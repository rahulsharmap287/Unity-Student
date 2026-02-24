package main

import (
	"fmt"
	"net/smtp"
	"os"
)

// Is function ko purane wale se replace karein
func SendOTPEmail(targetEmail string, otp string) {
	// Railway Dashboard par set kiye gaye variables uthayega
	from := os.Getenv("SMTP_EMAIL")
	pass := os.Getenv("SMTP_PASS")

	if from == "" || pass == "" {
		fmt.Println("❌ SMTP credentials not set in environment (Check Railway Variables)")
		return
	}

	// SMTP Configuration
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Authentication Setup
	auth := smtp.PlainAuth("", from, pass, smtpHost)

	// Professional HTML Message Structure
	subject := "Subject: Unity Student - OTP Verification\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	// Aapka professional email template
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

	// ✨ Yahan sirf 'targetEmail' (User ka email) ko bhej rahe hain
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{targetEmail}, msg)

	if err != nil {
		fmt.Printf("❌ Email sending failed to %s: %s\n", targetEmail, err)
	} else {
		fmt.Printf("✅ OTP sent successfully to %s\n", targetEmail)
	}
}
