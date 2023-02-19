package discord

type webhookMessage struct {
	Embeds []*webhookMessageEmbed `json:"embeds"`
}

const webhookEmbedType = "rich"

// webhookMessageEmbed is the embed message of discord webhook.
// https://discord.com/developers/docs/resources/channel#embed-object
type webhookMessageEmbed struct {
	Title       string                      `json:"title"`
	Type        string                      `json:"type"`
	Description string                      `json:"description"`
	Timestamp   string                      `json:"timestamp"`
	Color       uint32                      `json:"color"`
	Fields      []*webhookMessageEmbedField `json:"fields"`
}

// webhookMessageEmbedField is the embed field of discord webhook message.
// https://discord.com/developers/docs/resources/channel#embed-object-embed-field-structure
type webhookMessageEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}
