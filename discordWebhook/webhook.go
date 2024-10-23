package discordWebhook

import (
	"fmt"
	discordWebhook "github.com/dozerokz/discord-webhook-go"
)

const webhookUrl = "YOUR_DISCORD_WEBHOOK_URL"

func createWebhook(swapMessage, txHash, tokenLink string) discordWebhook.Webhook {
	// todo handle error
	webhook, _ := discordWebhook.CreateWebhook("", "", "")
	// todo handle error
	embed, _ := discordWebhook.CreateEmbed("Swap detected", "", "", discordWebhook.RGB{
		R: 250,
		G: 93,
		B: 99,
	})

	txHashLink := fmt.Sprintf("[%s](https://etherscan.io/tx/%s)", txHash, txHash)
	swap := discordWebhook.CreateField("Swap details", swapMessage, true)
	hash := discordWebhook.CreateField("Transaction hash", txHashLink, false)
	token := discordWebhook.CreateField("Token address", tokenLink, false)

	embed.SetTimestamp()

	embed.AddField(swap)
	embed.AddField(hash)
	embed.AddField(token)

	webhook.AddEmbed(embed)
	return webhook
}

func SendWebhook(swapMessage, txHash, tokenLink string) error {
	webhook := createWebhook(swapMessage, txHash, tokenLink)
	err := discordWebhook.SendWebhook(webhookUrl, webhook)
	return err
}
