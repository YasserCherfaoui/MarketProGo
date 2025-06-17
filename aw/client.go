package aw

import (
	"github.com/YasserCherfaoui/MarketProGo/cfg"
	"github.com/appwrite/sdk-for-go/appwrite"
	"github.com/appwrite/sdk-for-go/client"
)

func NewAppwriteClient(cfg *cfg.AppConfig) client.Client {
	client := appwrite.NewClient(
		appwrite.WithEndpoint(cfg.AppwriteEndpoint),
		appwrite.WithProject(cfg.AppwriteProject),
		appwrite.WithKey(cfg.AppwriteKey),
	)

	return client
}
