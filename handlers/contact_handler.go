package handlers

import (
	"fmt"
	"net/http"
	"net/smtp"
	"stajyer-backend/config"
	"stajyer-backend/models"

	"github.com/gin-gonic/gin"
)

func SendContactMessage(c *gin.Context) {
	var msg models.ContactMessage

	// 1. Form Verilerini Al
	msg.Name = c.PostForm("name")
	msg.Email = c.PostForm("email")
	msg.Message = c.PostForm("message")

	// 2. Veritabanına (Dataset) Kaydet
	if err := config.DB.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Mesaj kaydedilemedi"})
		return
	}

	// 3. E-posta Gönderimi (canankkrc@gmail.com)
	go sendEmailToAdmin(msg) // Arka planda gönder (kullanıcıyı bekletme)

	c.JSON(http.StatusOK, gin.H{"message": "Mesaj iletildi ve e-posta gönderildi!"})
}

func sendEmailToAdmin(msg models.ContactMessage) {
	// 1. Senin Sistem Mail Bilgilerin
	from := "canankkrc@gmail.com"    // Gönderici (Uygulama Şifresi aldığın mail)
	password := ""                   // Google'dan aldığın 16 haneli şifre
	to := "canan123karaca@gmail.com" // Mesajın düşeceği asıl hesap

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// 2. Dinamik İçerik Hazırlama
	// Kullanıcının adı konu başlığında görünecek
	subject := fmt.Sprintf("Subject: Yeni Mesaj: %s (%s)\n", msg.Name, msg.Email)

	// Reply-To ekleyerek "Yanıtla" dediğinde kullanıcıya gitmesini sağlıyoruz
	replyTo := fmt.Sprintf("Reply-To: %s\n", msg.Email)

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; padding: 20px; border: 1px solid #eee;">
			<h2 style="color: #6b46c1;">Yeni İletişim Formu Mesajı</h2>
			<p><b>Gönderen Adı:</b> %s</p>
			<p><b>Gönderen E-posta:</b> <a href="mailto:%s">%s</a></p>
			<hr style="border: 0; border-top: 1px solid #eee;">
			<p><b>Mesaj İçeriği:</b></p>
			<p style="background: #f9f9f9; padding: 15px; border-radius: 8px;">%s</p>
		</div>
	`, msg.Name, msg.Email, msg.Email, msg.Message)

	// 3. Gönderim İşlemi
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Mesaj başlıklarını birleştiriyoruz
	message := []byte(subject + replyTo + mime + body)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message)

	if err != nil {
		fmt.Printf("E-posta gönderilirken hata oluştu: %v\n", err)
	} else {
		fmt.Println("E-posta başarıyla canankkrc@gmail.com adresine iletildi.")
	}
}
