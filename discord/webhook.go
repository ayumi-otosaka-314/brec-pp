package discord

// WebhookMessage is the message payload of discord webhook.
// https://discord.com/developers/docs/resources/webhook#execute-webhook
type WebhookMessage struct {
	Embeds []*MessageEmbed `json:"embeds"`
}

// WebhookEmbedType is the fixed content type of embed object in webhook message.
// MessageEmbed.Type should always be set to this value.
const WebhookEmbedType = "rich"

// MessageEmbed is the embed object of discord message.
// https://discord.com/developers/docs/resources/channel#embed-object
type MessageEmbed struct {
	Title       string                 `json:"title"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Timestamp   string                 `json:"timestamp"`
	Color       uint32                 `json:"color"`
	Image       *MessageEmbedImage     `json:"image"`
	Thumbnail   *MessageEmbedThumbnail `json:"thumbnail"`
	Author      *MessageEmbedAuthor    `json:"author"`
	Fields      []*MessageEmbedField   `json:"fields"`
}

// MessageEmbedImage is the image of discord embed object in message.
// https://discord.com/developers/docs/resources/channel#embed-object-embed-image-structure
type MessageEmbedImage struct {
	URL string `json:"url"`
}

// MessageEmbedThumbnail is the thumbnail image of discord embed object in message.
// https://discord.com/developers/docs/resources/channel#embed-object-embed-thumbnail-structure
type MessageEmbedThumbnail struct {
	URL string `json:"url"`
}

// MessageEmbedAuthor is the author of discord embed object in message.
// https://discord.com/developers/docs/resources/channel#embed-object-embed-author-structure
type MessageEmbedAuthor struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	IconURL string `json:"icon_url"`
}

// MessageEmbedField is the field of discord embed object in message.
// https://discord.com/developers/docs/resources/channel#embed-object-embed-field-structure
type MessageEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type WebhookResponse struct {
	ID              string         `json:"id"`
	Type            uint8          `json:"type"`
	ChannelID       string         `json:"channel_id"`
	Author          *MessageAuthor `json:"author"`
	Timestamp       string         `json:"timestamp"`
	EditedTimestamp string         `json:"edited_timestamp"`
	Flags           uint8          `json:"flags"`
	WebhookID       string         `json:"webhook_id"`
}

type MessageAuthor struct {
	IsBot         bool   `json:"bot"`
	ID            string `json:"id"`
	Username      string `json:"username"`
	Avatar        string `json:"avatar"`
	Discriminator string `json:"discriminator"`
}
