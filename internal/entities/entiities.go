package entities

type Message struct {
	Id        int64     `json:"id"`
	Subscribe Subscribe `json:"subscribe"`
} //@name Message

type Subscribe struct {
	Channel Channel `json:"channel"`
} //@name Subscribe
