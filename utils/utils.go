package utils

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nlopes/slack"
)

// GlobalUsers - indexed user space
type GlobalUsers struct {
	Users     []slack.User
	nameIndex map[string]int
	idIndex   map[string]int
}

// GetUsers - get all users to local variable from Slack
func (g *GlobalUsers) GetUsers(api *slack.Client) {
	var err error

	g.nameIndex = make(map[string]int)
	g.idIndex = make(map[string]int)

	g.Users, err = api.GetUsers()
	if err != nil {
		log.Printf("Error getting users: %v", err)
	}

	for i := 0; i < len(g.Users); i++ {
		user := g.Users[i]
		g.nameIndex[user.Name] = i
		g.idIndex[user.ID] = i
	}
}

// SetPresenceByID - set user presence in local variable (not in Slack)
func (g *GlobalUsers) SetPresenceByID(id string, presence string) {
	idx, ok := g.idIndex[id]
	if !ok {
		log.Printf("Unknown input ID: %s", id)
		return
	}

	g.Users[idx].Presence = presence
}

// GetPresenceByName - get user presence from local variable (not from Slack)
func (g *GlobalUsers) GetPresenceByName(name string) string {
	idx, ok := g.nameIndex[name]
	if !ok {
		log.Printf("Unknown input name: %s", name)
		// return translation result as below
		return fmt.Sprintf("UNKNOWN_PRESENCE_FROM_NAME_%s", name)
	}
	return g.Users[idx].Presence
}

// GetPresenceByID - get user presence from local variable (not from Slack)
func (g *GlobalUsers) GetPresenceByID(id string) string {
	idx, ok := g.idIndex[id]
	if !ok {
		log.Printf("Unknown input ID: %s", id)
		// return translation result as below
		return fmt.Sprintf("UNKNOWN_PRESENCE_FROM_ID_%s", id)
	}
	return g.Users[idx].Presence
}

// NameToID - translate Slack name (without @) to Slack ID
func (g *GlobalUsers) NameToID(name string) string {
	idx, ok := g.nameIndex[name]
	if !ok {
		log.Printf("Unknown input name: %s", name)
		// return translation result as below
		return fmt.Sprintf("UNKNOWN_ID_FROM_NAME_%s", name)
	}
	return g.Users[idx].ID
}

// IDToName - translate Slack ID  to Slack name (without @)
func (g *GlobalUsers) IDToName(id string) string {
	idx, ok := g.idIndex[id]
	if !ok {
		log.Printf("Unknown input ID: %s", id)
		// return translation result as below
		return fmt.Sprintf("UNKNOWN_NAME_FROM_ID_%s", id)
	}
	return g.Users[idx].Name
}

type idxGrpChan struct {
	t string
	i int
}

// GlobalChannels - indexed channel space
type GlobalChannels struct {
	Channels  []slack.Channel
	Groups    []slack.Group
	IMs       []slack.IM
	nameIndex map[string]idxGrpChan
	idIndex   map[string]idxGrpChan
}

// GetChannels - get all channels to local variable from Slack
func (g *GlobalChannels) GetChannels(api *slack.Client) {
	var err error

	g.nameIndex = make(map[string]idxGrpChan)
	g.idIndex = make(map[string]idxGrpChan)

	g.Channels, err = api.GetChannels(true)
	if err != nil {
		log.Printf("Error getting channels: %v", err)
		return
	}

	for i := 0; i < len(g.Channels); i++ {
		channel := g.Channels[i]
		// log.Printf("Channel debug: %s", StructPrettyPrint(channel))
		log.Printf("Channel Name: %s, ID: %s, IsChannel: %t\n", channel.Name, channel.ID, channel.IsChannel)
		g.nameIndex[channel.Name] = idxGrpChan{t: "c", i: i}
		g.idIndex[channel.ID] = idxGrpChan{t: "c", i: i}
	}

	g.Groups, err = api.GetGroups(true)
	if err != nil {
		log.Printf("Error getting groups: %v", err)
		return
	}

	for i := 0; i < len(g.Groups); i++ {
		group := g.Groups[i]
		log.Printf("Group Name: %s, ID: %s, IsChannel: %t\n", group.Name, group.ID, false)
		g.nameIndex[group.Name] = idxGrpChan{t: "g", i: i}
		g.idIndex[group.ID] = idxGrpChan{t: "g", i: i}
	}

	g.IMs, err = api.GetIMChannels()
	if err != nil {
		log.Printf("Error getting IM channels: %v", err)
		return
	}

	for i := 0; i < len(g.IMs); i++ {
		im := g.IMs[i]
		log.Printf("IM Name: %s, ID: %s, IsChannel: %t\n", im.User, im.ID, false)
		g.nameIndex[im.User] = idxGrpChan{t: "d", i: i}
		g.idIndex[im.ID] = idxGrpChan{t: "d", i: i}
	}

}

// NameToID - translate Slack name (without @) to Slack ID
func (g *GlobalChannels) NameToID(name string) string {
	var id string

	idxStruct, ok := g.nameIndex[name]
	if !ok {
		log.Printf("Unknown input name: %s", name)
		// return translation result as below
		return fmt.Sprintf("UNKNOWN_ID_FROM_NAME_%s", name)
	}

	if idxStruct.t == "c" {
		id = g.Channels[idxStruct.i].ID
	} else if idxStruct.t == "g" {
		id = g.Groups[idxStruct.i].ID
	} else if idxStruct.t == "d" {
		id = g.IMs[idxStruct.i].ID
	}

	return id
}

// IDToName - translate Slack ID  to Slack name (without @)
func (g *GlobalChannels) IDToName(id string) string {
	var name string

	idxStruct, ok := g.idIndex[id]
	if !ok {
		log.Printf("Unknown input ID: %s", id)
		// return translation result as below
		return fmt.Sprintf("UNKNOWN_NAME_FROM_ID_%s", id)
	}
	if idxStruct.t == "c" {
		name = g.Channels[idxStruct.i].Name
	} else if idxStruct.t == "g" {
		name = g.Groups[idxStruct.i].Name
	} else if idxStruct.t == "d" {
		id = g.IMs[idxStruct.i].User
	}
	return name
}

// StructPrettyPrint - JSON like
func StructPrettyPrint(s interface{}) string {
	bytesStruct, _ := json.MarshalIndent(s, "", "  ")
	return string(bytesStruct)
}
