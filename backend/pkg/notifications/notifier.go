package notifications

import (
	"github.com/jminat01/dashboard-servers-go/backend/internal/models"
	"github.com/jminat01/dashboard-servers-go/backend/pkg/logger"
)

// Notifier interfaz para diferentes proveedores de notificaciones
type Notifier interface {
	// SendAlert envía una notificación de alerta
	SendAlert(alert *models.Alert) error

	// SendResolvedAlert envía una notificación de alerta resuelta
	SendResolvedAlert(alert *models.Alert) error
}

// NotificationManager gestiona diferentes proveedores de notificaciones
type NotificationManager struct {
	discordClient *DiscordClient
	// Futuros proveedores: email, webhooks externos, etc
	logger logger.Logger
}

// NotificationConfig configuración para las notificaciones
type NotificationConfig struct {
	// Discord
	DiscordEnabled    bool
	DiscordWebhookURL string
	DiscordBotName    string
	DiscordAvatarURL  string

	// Email (futuro)
	EmailEnabled bool
	SMTPServer   string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	EmailFrom    string

	// Webhook genérico (futuro)
	WebhookEnabled bool
	WebhookURL     string
	WebhookSecret  string
}

// NewNotificationManager crea un nuevo gestor de notificaciones
func NewNotificationManager(config *NotificationConfig, log logger.Logger) *NotificationManager {
	manager := &NotificationManager{
		logger: log,
	}

	// Inicializar cliente de Discord si está habilitado
	if config.DiscordEnabled && config.DiscordWebhookURL != "" {
		manager.discordClient = NewDiscordClient(
			config.DiscordWebhookURL,
			config.DiscordBotName,
			config.DiscordAvatarURL,
			log,
		)
		log.Info("Cliente de notificaciones Discord inicializado")
	}

	return manager
}

// NotifyAlert envía una alerta a todos los canales configurados
func (nm *NotificationManager) NotifyAlert(alert *models.Alert, threshold *models.AlertThreshold) error {
	var notifyChannels []string

	// Registrar canales utilizados
	if threshold.EnableDiscord && nm.discordClient != nil {
		if err := nm.discordClient.SendAlert(alert); err != nil {
			nm.logger.Errorf("Error al enviar alerta a Discord: %v", err)
		} else {
			notifyChannels = append(notifyChannels, "discord")
		}
	}

	// TODO: Implementar otros canales (email, webhook, etc)

	// Actualizar canales en la alerta
	alert.NotifyChannels = notifyChannels

	return nil
}

// NotifyResolvedAlert envía una notificación de alerta resuelta
func (nm *NotificationManager) NotifyResolvedAlert(alert *models.Alert) error {
	// Enviar solo si la alerta fue notificada previamente
	for _, channel := range alert.NotifyChannels {
		switch channel {
		case "discord":
			if nm.discordClient != nil {
				if err := nm.discordClient.SendResolvedAlert(alert); err != nil {
					nm.logger.Errorf("Error al enviar resolución de alerta a Discord: %v", err)
				}
			}
			// TODO: Otros canales
		}
	}

	return nil
}
