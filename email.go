package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func SendOTPEmail(targetEmail string, otp string) {
	// Brevo API configuration
	apiKey := "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"
	url := "https://api.brevo.com/v3/smtp/email"

	// Email payload structure
	payload := map[string]interface{}{
		"sender": map[string]string{
			"name":  "Unity Student",
			"email": "unitystudent42@gmail.com",
		},
		"to": []map[string]string{
			{
				"email": targetEmail,
			},
		},
		"subject": "Unity Student - OTP Verification",
		"htmlContent": fmt.Sprintf(`
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
        </html>`, otp),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("❌ JSON error: %s\n", err)
		return
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ Request error: %s\n", err)
		return
	}

	// Set required headers for Brevo
	req.Header.Set("api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ API Connection error: %s\n", err)
		return
	}
	defer resp.Body.Close()

	// Check if email was sent successfully
	if resp.StatusCode == 201 || resp.StatusCode == 200 {
		fmt.Printf("✅ OTP sent successfully via Brevo API to %s\n", targetEmail)
	} else {
		fmt.Printf("❌ Brevo API failed. Status: %d\n", resp.StatusCode)
	}
}
