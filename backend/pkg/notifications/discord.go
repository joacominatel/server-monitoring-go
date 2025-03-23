package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jminat01/dashboard-servers-go/backend/internal/models"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/logger"
)

// DiscordWebhook estructura para enviar mensajes a Discord
type DiscordWebhook struct {
	Content   string         `json:"content,omitempty"`
	Username  string         `json:"username,omitempty"`
	AvatarURL string         `json:"avatar_url,omitempty"`
	Embeds    []DiscordEmbed `json:"embeds,omitempty"`
}

// DiscordEmbed estructura para enviar embeds a Discord
type DiscordEmbed struct {
	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	URL         string              `json:"url,omitempty"`
	Color       int                 `json:"color,omitempty"` // Color en formato decimal
	Timestamp   string              `json:"timestamp,omitempty"`
	Footer      *DiscordEmbedFooter `json:"footer,omitempty"`
	Thumbnail   *DiscordThumbnail   `json:"thumbnail,omitempty"`
	Fields      []DiscordField      `json:"fields,omitempty"`
}

// DiscordEmbedFooter estructura para el footer de un embed
type DiscordEmbedFooter struct {
	Text    string `json:"text,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

// DiscordThumbnail estructura para la miniatura de un embed
type DiscordThumbnail struct {
	URL string `json:"url,omitempty"`
}

// DiscordField estructura para un campo de un embed
type DiscordField struct {
	Name   string `json:"name,omitempty"`
	Value  string `json:"value,omitempty"`
	Inline bool   `json:"inline,omitempty"`
}

// DiscordClient cliente para enviar notificaciones a Discord
type DiscordClient struct {
	webhookURL  string
	httpClient  *http.Client
	botUsername string
	avatarURL   string
	logger      logger.Logger
}

// NewDiscordClient crea un nuevo cliente de Discord
func NewDiscordClient(webhookURL, botUsername, avatarURL string, log logger.Logger) *DiscordClient {
	return &DiscordClient{
		webhookURL:  webhookURL,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		botUsername: botUsername,
		avatarURL:   avatarURL,
		logger:      log,
	}
}

// GetColorForSeverity devuelve un color según la severidad
func GetColorForSeverity(severity models.AlertSeverity) int {
	switch severity {
	case models.AlertSeverityCritical:
		return 15158332 // Rojo
	case models.AlertSeverityWarning:
		return 16776960 // Amarillo
	case models.AlertSeverityInfo:
		return 3447003 // Azul
	default:
		return 10197915 // Gris
	}
}

// SendAlert envía una alerta a Discord
func (dc *DiscordClient) SendAlert(alert *models.Alert) error {
	// Crear el embed para Discord
	embed := DiscordEmbed{
		Title:       alert.Title,
		Description: alert.Message,
		Color:       GetColorForSeverity(alert.Severity),
		Timestamp:   alert.TriggeredAt.Format(time.RFC3339),
		Footer: &DiscordEmbedFooter{
			Text: "Sistema de Monitoreo de Servidores",
		},
		Fields: []DiscordField{
			{
				Name:   "Servidor",
				Value:  alert.Server.Hostname,
				Inline: true,
			},
			{
				Name:   "IP",
				Value:  alert.Server.IP,
				Inline: true,
			},
			{
				Name:   "Métrica",
				Value:  string(alert.MetricType),
				Inline: true,
			},
			{
				Name:   "Valor",
				Value:  fmt.Sprintf("%.2f", alert.MetricValue),
				Inline: true,
			},
			{
				Name:   "Umbral",
				Value:  fmt.Sprintf("%s %.2f", alert.Operator, alert.Threshold),
				Inline: true,
			},
			{
				Name:   "Severidad",
				Value:  string(alert.Severity),
				Inline: true,
			},
		},
	}

	// Crear el mensaje completo
	webhook := DiscordWebhook{
		Username:  dc.botUsername,
		AvatarURL: dc.avatarURL,
		Embeds:    []DiscordEmbed{embed},
	}

	// Convertir a JSON
	jsonData, err := json.Marshal(webhook)
	if err != nil {
		dc.logger.Errorf("Error al serializar webhook Discord: %v", err)
		return err
	}

	// Enviar solicitud HTTP
	resp, err := dc.httpClient.Post(dc.webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		dc.logger.Errorf("Error al enviar webhook a Discord: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Verificar respuesta
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		dc.logger.Errorf("Error en respuesta de Discord. Código: %d", resp.StatusCode)
		return fmt.Errorf("error en respuesta de Discord. Código: %d", resp.StatusCode)
	}

	dc.logger.Infof("Alerta #%d enviada exitosamente a Discord", alert.ID)
	return nil
}

// SendResolvedAlert envía una notificación de alerta resuelta
func (dc *DiscordClient) SendResolvedAlert(alert *models.Alert) error {
	// Crear el embed para Discord con color verde para alerta resuelta
	embed := DiscordEmbed{
		Title:       fmt.Sprintf("✅ RESUELTA: %s", alert.Title),
		Description: fmt.Sprintf("La alerta ha sido resuelta automáticamente:\n%s", alert.Message),
		Color:       3066993, // Verde
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &DiscordEmbedFooter{
			Text: "Sistema de Monitoreo de Servidores",
		},
		Fields: []DiscordField{
			{
				Name:   "Servidor",
				Value:  alert.Server.Hostname,
				Inline: true,
			},
			{
				Name:   "Duración",
				Value:  getDurationText(alert.TriggeredAt, *alert.ResolvedAt),
				Inline: true,
			},
		},
	}

	webhook := DiscordWebhook{
		Username:  dc.botUsername,
		AvatarURL: dc.avatarURL,
		Embeds:    []DiscordEmbed{embed},
	}

	jsonData, err := json.Marshal(webhook)
	if err != nil {
		return err
	}

	resp, err := dc.httpClient.Post(dc.webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("error en respuesta de Discord. Código: %d", resp.StatusCode)
	}

	return nil
}

// Función auxiliar para calcular duración en texto
func getDurationText(start, end time.Time) string {
	duration := end.Sub(start)

	if duration < time.Minute {
		return fmt.Sprintf("%d segundos", int(duration.Seconds()))
	} else if duration < time.Hour {
		return fmt.Sprintf("%d minutos", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else {
		days := int(duration.Hours()) / 24
		hours := int(duration.Hours()) % 24
		return fmt.Sprintf("%dd %dh", days, hours)
	}
}
