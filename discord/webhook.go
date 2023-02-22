package discord

// webhookMessage is the message payload of discord webhook.
// https://discord.com/developers/docs/resources/webhook#execute-webhook
type webhookMessage struct {
	Embeds []*messageEmbed `json:"embeds"`
}

// webhookEmbedType is the fixed content type of embed object in webhook message.
// messageEmbed.Type should always be set to this value.
const webhookEmbedType = "rich"

// messageEmbed is the embed object of discord message.
// https://discord.com/developers/docs/resources/channel#embed-object
type messageEmbed struct {
	Title       string                 `json:"title"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Timestamp   string                 `json:"timestamp"`
	Color       uint32                 `json:"color"`
	Image       *messageEmbedImage     `json:"image"`
	Thumbnail   *messageEmbedThumbnail `json:"thumbnail"`
	Author      *messageEmbedAuthor    `json:"author"`
	Fields      []*messageEmbedField   `json:"fields"`
}

// messageEmbedImage is the image of discord embed object in message.
// https://discord.com/developers/docs/resources/channel#embed-object-embed-image-structure
type messageEmbedImage struct {
	URL string `json:"url"`
}

// messageEmbedThumbnail is the thumbnail image of discord embed object in message.
// https://discord.com/developers/docs/resources/channel#embed-object-embed-thumbnail-structure
type messageEmbedThumbnail struct {
	URL string `json:"url"`
}

// messageEmbedAuthor is the author of discord embed object in message.
// https://discord.com/developers/docs/resources/channel#embed-object-embed-author-structure
type messageEmbedAuthor struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	IconURL string `json:"icon_url"`
}

// messageEmbedField is the field of discord embed object in message.
// https://discord.com/developers/docs/resources/channel#embed-object-embed-field-structure
type messageEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type webhookResponse struct {
	ID        string `json:"id"`
	Type      uint8  `json:"type"`
	ChannelID string `json:"channel_id"`
}

type messageAuthor struct {
	IsBot           bool   `json:"bot"`
	ID              string `json:"id"`
	Username        string `json:"username"`
	Avatar          string `json:"avatar"`
	Discriminator   string `json:"discriminator"`
	Timestamp       string `json:"timestamp"`
	EditedTimestamp string `json:"edited_timestamp"`
	Flags           uint8  `json:"flags"`
	WebhookID       string `json:"webhook_id"`
}
