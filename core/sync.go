package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/streadway/amqp"
)

func (g *Garden) syncAmqp(msg amqp.Delivery) {
	body := msg.Body
	var data amqpMsg
	if err := json.Unmarshal(body, &data); err != nil {
		g.Log(ErrorLevel, "syncRoutesYml", err)
		return
	}
	switch data.Cmd {
	case "routes":
		if err := g.receiveRoutes(data.Data); err != nil {
			g.Log(ErrorLevel, "receiveRoutes", err)
		}
	}

	if err := msg.Ack(true); err != nil {
		g.Log(ErrorLevel, "amqpAck", err)
	}
}

func (g *Garden) receiveRoutes(data MapData) error {
	if yml, ok := data["yml"]; ok {
		if err := writeFile("configs/routes.yml", []byte(yml.(string))); err != nil {
			return err
		}
		g.Log(InfoLevel, "syncRoutes", "Success")
		return nil
	}
	return errors.New("not found map index: yml")
}

func (g *Garden) sendRoutes() {
	fileData, err := readFile("configs/routes.yml")
	if err != nil {
		g.Log(ErrorLevel, "SyncRoutes", err)
		return
	}

	if len(fileData) == 0 {
		return
	}

	if bytes.Compare(g.syncCache, fileData) == 0 {
		return
	}

	g.syncCache = fileData

	msg := amqpMsg{
		Cmd: "routes",
		Data: MapData{
			"yml": string(fileData),
		},
	}
	s, _ := json.Marshal(msg)
	if err := g.amqpPublish("fanout", "sync", "", "sync", string(s)); err != nil {
		g.Log(ErrorLevel, "amqpPublish", err)
	}
}
