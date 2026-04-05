package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	apiKey string
	from   string
}

func NewClient(apiKey, from string) *Client {
	return &Client{apiKey: apiKey, from: from}
}

func (c *Client) SendPasswordReset(toEmail, toName, resetURL string) error {
	body := map[string]any{
		"from":    c.from,
		"to":      []string{toEmail},
		"subject": "Reset your NotTennis password",
		"html": fmt.Sprintf(`
<div style="font-family:sans-serif;max-width:480px;margin:0 auto;padding:40px 24px">
  <h1 style="font-size:24px;font-weight:800;margin:0 0 8px">Reset your password</h1>
  <p style="color:#666;margin:0 0 32px">Hi %s, click the button below to set a new password. This link expires in 1 hour.</p>
  <a href="%s" style="display:inline-block;background:#1a7a4a;color:#fff;font-weight:700;text-decoration:none;padding:14px 28px;border-radius:12px">
    Reset password →
  </a>
  <p style="color:#999;font-size:12px;margin:32px 0 0">If you didn't request this, you can safely ignore this email.</p>
</div>`, toName, resetURL),
	}

	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("resend: unexpected status %d", resp.StatusCode)
	}
	return nil
}
