package config

import (
	"os"
	"strconv"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
)

type MidtransConfig struct {
	ServerKey       string
	ClientKey       string
	Environment     midtrans.EnvironmentType
	NotificationURL string
	FinishURL       string
	ExpiryMinutes   int64
}

func LoadMidtransConfig() MidtransConfig {
	environment := midtrans.Sandbox
	if os.Getenv("MIDTRANS_ENVIRONMENT") == "production" {
		environment = midtrans.Production
	}

	expiryMinutes := int64(60)
	if value, err := strconv.ParseInt(os.Getenv("MIDTRANS_EXPIRY_MINUTES"), 10, 64); err == nil && value > 0 {
		expiryMinutes = value
	}

	return MidtransConfig{
		ServerKey:       os.Getenv("MIDTRANS_SERVER_KEY"),
		ClientKey:       os.Getenv("MIDTRANS_CLIENT_KEY"),
		Environment:     environment,
		NotificationURL: os.Getenv("MIDTRANS_NOTIFICATION_URL"),
		FinishURL:       os.Getenv("MIDTRANS_FINISH_URL"),
		ExpiryMinutes:   expiryMinutes,
	}
}

func NewMidtransClients(cfg MidtransConfig) (snap.Client, coreapi.Client) {
	var snapClient snap.Client
	snapClient.New(cfg.ServerKey, cfg.Environment)
	if cfg.NotificationURL != "" {
		snapClient.Options.SetPaymentOverrideNotification(cfg.NotificationURL)
	}

	var coreClient coreapi.Client
	coreClient.New(cfg.ServerKey, cfg.Environment)

	return snapClient, coreClient
}
