package smtp

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendPasswordResetEmail(recipient, code string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")

	from := username
	to := []string{recipient}

	subject := "Запрос на смену пароля --PortfolioTracker\r\n"
	mime := "MIME-version: 1.0;\r\nContent-Type: text/plain; charset=\"UTF-8\";\r\n"
	body := "Ваш код: 123456. Никому не сообщайте этот код."
	message := []byte(subject + mime + "\r\n" + body)

	auth := smtp.PlainAuth("", username, password, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		return fmt.Errorf("Ошибка отправки: %w", err)
	}

	return nil
}
