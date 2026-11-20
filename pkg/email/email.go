package email

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

// SendOTP mengirimkan email berisikan kode OTP untuk pemulihan password
func SendOTP(toEmail, toName, otp string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := os.Getenv("SMTP_FROM")

	if smtpFrom == "" {
		smtpFrom = "SiLapor Kampus <no-reply@silapor.ac.id>"
	}

	// Jika kredensial SMTP belum diisi di .env, kita log kode OTP ke console untuk kemudahan pengujian
	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		log.Printf("[DEV EMAIL SIMULATOR] Pengiriman OTP ke %s (%s). KODE OTP: %s\n", toEmail, toName, otp)
		return nil
	}

	port := 587
	if p, err := strconv.Atoi(smtpPortStr); err == nil && p > 0 {
		port = p
	}

	m := gomail.NewMessage()
	m.SetHeader("From", smtpFrom)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Kode OTP Reset Password SiLapor")

	htmlBody := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="utf-8">
		<style>
			body { font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif; background-color: #f8fafc; padding: 20px; color: #1e293b; }
			.container { max-width: 520px; margin: 0 auto; background: #ffffff; border-radius: 16px; padding: 32px; border: 1px solid #e2e8f0; }
			.logo { font-size: 24px; font-weight: bold; margin-bottom: 20px; }
			.logo span.si { color: #f97316; }
			.logo span.lapor { color: #4f46e5; }
			.otp-box { background: #f1f5f9; border-radius: 12px; padding: 16px; text-align: center; margin: 24px 0; font-size: 32px; font-weight: 800; letter-spacing: 8px; color: #4f46e5; border: 1px dashed #cbd5e1; }
			.footer { margin-top: 32px; font-size: 12px; color: #64748b; line-height: 1.5; border-top: 1px solid #f1f5f9; padding-top: 16px; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="logo"><span class="si">Si</span><span class="lapor">Lapor</span></div>
			<h2>Reset Password Akun</h2>
			<p>Halo <strong>%s</strong>,</p>
			<p>Kami menerima permintaan untuk mereset kata sandi akun SiLapor Anda. Gunakan kode OTP berikut untuk melanjutkan:</p>
			<div class="otp-box">%s</div>
			<p>Kode ini berlaku selama <strong>15 menit</strong>. Jangan bagikan kode ini kepada siapapun termasuk petugas kampus.</p>
			<div class="footer">
				<p>Jika Anda tidak merasa melakukan permintaan ini, silakan abaikan email ini. Akun Anda tetap aman.</p>
			</div>
		</div>
	</body>
	</html>
	`, toName, otp)

	m.SetBody("text/html", htmlBody)

	d := gomail.NewDialer(smtpHost, port, smtpUser, smtpPass)
	if err := d.DialAndSend(m); err != nil {
		log.Printf("Peringatan: Gagal mengirim email via SMTP (%v). Mencatat OTP ke log: %s\n", err, otp)
		// Kita tidak memblokir flow jika SMTP sedang terkendala, OTP tetap tercatat di log server
		return nil
	}

	return nil
}
