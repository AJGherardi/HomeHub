package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	math "math/rand"
	"os"
	"reflect"
	"time"

	"github.com/AJGherardi/HomeHub/generated"
	"github.com/AJGherardi/HomeHub/model"
	"github.com/AJGherardi/HomeHub/utils"
	"github.com/grandcat/zeroconf"
)

func (r *deviceResolver) Elements(ctx context.Context, obj *model.Device) ([]*generated.ElementResponse, error) {
	elementResponses := []*generated.ElementResponse{}
	for i := range obj.Elements {
		elementResponses = append(
			elementResponses,
			&generated.ElementResponse{Addr: int(i), Element: obj.Elements[i]},
		)
	}
	return elementResponses, nil
}

func (r *elementResolver) State(ctx context.Context, obj *model.Element) (string, error) {
	return utils.EncodeBase64(obj.State), nil
}

func (r *groupResolver) Scenes(ctx context.Context, obj *model.Group) ([]*generated.SceneResponse, error) {
	sceneResponses := []*generated.SceneResponse{}
	for i := range obj.Scenes {
		sceneResponses = append(
			sceneResponses,
			&generated.SceneResponse{Number: int(i), Scene: obj.Scenes[i]},
		)
	}
	return sceneResponses, nil
}

func (r *groupResolver) Devices(ctx context.Context, obj *model.Group) ([]*generated.DeviceResponse, error) {
	deviceResponses := []*generated.DeviceResponse{}
	for i := range obj.Devices {
		deviceResponses = append(
			deviceResponses,
			&generated.DeviceResponse{Addr: int(i), Device: obj.Devices[i]},
		)
	}
	return deviceResponses, nil
}

func (r *mutationResolver) AddDevice(ctx context.Context, addr int, devUUID string, name string) (int, error) {
	// Provision device
	uuid := utils.DecodeBase64(devUUID)
	r.Controller.Provision(uuid)
	// Wait for node added
	nodeAddr := <-r.NodeAdded
	// Create device object
	device := model.Device{}
	// If device is a 2 plug outlet
	if reflect.DeepEqual(uuid[6:8], []byte{0x00, 0x02}) {
		device = model.MakeDevice(
			"2Outlet",
			nodeAddr,
		)
	}
	if reflect.DeepEqual(uuid[6:8], []byte{0x00, 0x01}) {
		device = model.MakeDevice(
			"4Button",
			nodeAddr,
		)
	}
	// Get group
	group := r.Store.Groups[uint16(addr)]
	// Configure Device
	r.Controller.ConfigureNode(nodeAddr, group.KeyIndex)
	time.Sleep(100 * time.Millisecond)
	// If device is a 2 plug outlet
	if reflect.DeepEqual(uuid[6:8], []byte{0x00, 0x02}) {
		// Set type and add elements
		elemAddr0 := device.AddElem(name+"-0", "onoff", nodeAddr)
		r.Controller.ConfigureElem(uint16(addr), nodeAddr, elemAddr0, group.KeyIndex)
		time.Sleep(100 * time.Millisecond)
		elemAddr1 := device.AddElem(name+"-1", "onoff", nodeAddr)
		r.Controller.ConfigureElem(uint16(addr), nodeAddr, elemAddr1, group.KeyIndex)
	}
	// If device is a button
	if reflect.DeepEqual(uuid[6:8], []byte{0x00, 0x01}) {
		// Set type and add elements
		elemAddr0 := device.AddElem(name+"-0", "event", nodeAddr)
		r.Controller.ConfigureElem(uint16(addr), nodeAddr, elemAddr0, group.KeyIndex)
		time.Sleep(1000 * time.Millisecond)
		elemAddr1 := device.AddElem(name+"-1", "event", nodeAddr)
		r.Controller.ConfigureElem(uint16(addr), nodeAddr, elemAddr1, group.KeyIndex)
		time.Sleep(3000 * time.Millisecond)
		elemAddr2 := device.AddElem(name+"-2", "event", nodeAddr)
		r.Controller.ConfigureElem(uint16(addr), nodeAddr, elemAddr2, group.KeyIndex)
		time.Sleep(3000 * time.Millisecond)
		elemAddr3 := device.AddElem(name+"-3", "event", nodeAddr)
		r.Controller.ConfigureElem(uint16(addr), nodeAddr, elemAddr3, group.KeyIndex)
	}
	// Add device to group
	group.AddDevice(nodeAddr, device)
	return int(nodeAddr), nil
}

func (r *mutationResolver) RemoveDevice(ctx context.Context, addr int, groupAddr int) (int, error) {
	// Get device
	group := r.Store.Groups[uint16(groupAddr)]
	// Send reset payload
	r.Controller.ResetNode(uint16(addr))
	// Remove device from group
	group.RemoveDevice(uint16(addr))
	return addr, nil
}

func (r *mutationResolver) RemoveGroup(ctx context.Context, addr int) (int, error) {
	// Get groupAddr
	group := r.Store.Groups[uint16(addr)]
	// Reset devices
	for addr := range group.Devices {
		r.Controller.ResetNode(addr)
	}
	// Remove the group
	delete(r.Store.Groups, uint16(addr))
	return addr, nil
}

func (r *mutationResolver) AddGroup(ctx context.Context, name string) (int, error) {
	netData := r.Store.NetData
	// Get net values
	groupAddr := netData.GetNextGroupAddr()
	// Add a group
	group := model.MakeGroup(name, groupAddr, 0x0000)
	r.Store.Groups[groupAddr] = &group
	// Update net data
	netData.IncrementNextGroupAddr()
	return int(groupAddr), nil
}

func (r *mutationResolver) ConfigHub(ctx context.Context) (string, error) {
	if utils.CheckIfConfigured() {
		return "", errors.New("already configured")
	}
	// Switch the mdns server
	r.Mdns.Shutdown()
	r.Mdns, _ = zeroconf.Register("hub", "_alexandergherardi._tcp", "local.", 8080, nil, nil)
	// Make a web key
	webKey := make([]byte, 16)
	rand.Read(webKey)
	// Start data save threed
	go SaveStore(r.Store)
	// Add and get net data
	netData := model.MakeNetData(webKey)
	r.Store.NetData = &netData
	// Setup controller
	r.Controller.Setup()
	time.Sleep(100 * time.Millisecond)
	// Add an app key
	r.Controller.AddKey(0x0000)
	return utils.EncodeBase64(webKey), nil
}

func (r *mutationResolver) ResetHub(ctx context.Context) (bool, error) {
	if !utils.CheckIfConfigured() {
		return false, errors.New("not configured")
	}
	// Remove all devices
	groups := r.Store.Groups
	for _, group := range groups {
		for addr := range group.Devices {
			// Send reset payload
			r.Controller.ResetNode(addr)
		}
	}
	// TODO: Clean house
	// Reset mesh controller
	r.Controller.Reset()
	time.Sleep(time.Second)
	r.Controller.Reboot()
	go func() {
		time.Sleep(300 * time.Millisecond)
		os.Exit(0)
	}()
	return true, nil
}

func (r *mutationResolver) SetState(ctx context.Context, groupAddr int, addr int, value string) (bool, error) {
	state := utils.DecodeBase64(value)
	// Get appKey from group
	group := r.Store.Groups[uint16(groupAddr)]
	// Send State
	if true {
		// Send msg
		r.Controller.SendMessage(
			state[0],
			uint16(addr),
			group.KeyIndex,
		)
	}
	return true, nil
}

func (r *mutationResolver) SceneStore(ctx context.Context, name string, addr int) (int, error) {
	group := r.Store.Groups[uint16(addr)]
	netData := r.Store.NetData
	// Get and increment next scene number
	sceneNumber := netData.GetNextSceneNumber()
	netData.IncrementNextSceneNumber()
	// Store scene
	group.AddScene(name, sceneNumber)
	r.Controller.SendStoreMessage(sceneNumber, uint16(addr), group.KeyIndex)
	return int(sceneNumber), nil
}

func (r *mutationResolver) SceneRecall(ctx context.Context, sceneNumber int, addr int) (int, error) {
	group := r.Store.Groups[uint16(addr)]
	r.Controller.SendRecallMessage(uint16(sceneNumber), uint16(addr), group.KeyIndex)
	return sceneNumber, nil
}

func (r *mutationResolver) SceneDelete(ctx context.Context, sceneNumber int, addr int) (int, error) {
	group := r.Store.Groups[uint16(addr)]
	group.DeleteScene(uint16(sceneNumber))
	r.Controller.SendDeleteMessage(uint16(addr), uint16(addr), group.KeyIndex)
	return sceneNumber, nil
}

func (r *mutationResolver) EventBind(ctx context.Context, sceneNumber int, groupAddr int, devAddr int, elemAddr int) (int, error) {
	group := r.Store.Groups[uint16(groupAddr)]
	device := group.Devices[uint16(devAddr)]
	var sceneNumberBytes []byte
	binary.BigEndian.PutUint16(sceneNumberBytes, uint16(sceneNumber))
	device.UpdateState(uint16(elemAddr), sceneNumberBytes)
	r.Controller.SendBindMessage(uint16(sceneNumber), uint16(elemAddr), group.KeyIndex)
	return sceneNumber, nil
}

func (r *mutationResolver) AddUser(ctx context.Context) (string, error) {
	// Remove user pin
	r.UserPin = 000000
	// Make a web key
	webKey := make([]byte, 16)
	rand.Read(webKey)
	// Add webKey to netData
	netData := r.Store.NetData
	netData.AddWebKey(webKey)
	return utils.EncodeBase64(webKey), nil
}

func (r *queryResolver) AvailableDevices(ctx context.Context) ([]string, error) {
	uuids := make([]string, 0)
	for _, uuid := range *r.UnprovisionedNodes {
		b64 := utils.EncodeBase64(uuid)
		uuids = append(uuids, b64)
	}
	return uuids, nil
}

func (r *queryResolver) AvailableGroups(ctx context.Context) ([]*generated.GroupResponse, error) {
	groups := r.Store.Groups
	groupPointers := make([]*generated.GroupResponse, 0)
	for i := range groups {
		groupPointers = append(groupPointers, &generated.GroupResponse{Addr: int(i), Group: groups[i]})
	}
	return groupPointers, nil
}

func (r *queryResolver) GetUserPin(ctx context.Context) (int, error) {
	// Generate a 6 digt random number
	pin := math.Intn(1000000)
	r.UserPin = pin
	return r.UserPin, nil
}

func (r *subscriptionResolver) ListGroup(ctx context.Context, addr int) (<-chan *generated.GroupResponse, error) {
	groupChan := make(chan *generated.GroupResponse, 1)
	// Put initial result in chan
	group := r.Store.Groups[uint16(addr)]
	groupChan <- &generated.GroupResponse{
		Addr:  addr,
		Group: group,
	}
	return groupChan, nil
}

func (r *subscriptionResolver) GetState(ctx context.Context, groupAddr int, devAddr int, elemAddr int) (<-chan string, error) {
	stateChan := make(chan string, 1)
	r.StateObservers = append(r.StateObservers, stateObserver{
		groupAddr: uint16(groupAddr),
		devAddr:   uint16(devAddr),
		elemAddr:  uint16(elemAddr),
		messages:  stateChan,
		ctx:       ctx,
	})
	device := r.Store.Groups[uint16(groupAddr)].Devices[uint16(devAddr)]
	state := device.GetState(uint16(elemAddr))
	stateChan <- utils.EncodeBase64(state)
	return stateChan, nil
}

func (r *subscriptionResolver) GetEvents(ctx context.Context) (<-chan int, error) {
	eventChan := make(chan int, 1)
	r.EventObservers = append(r.EventObservers, eventObserver{
		messages: eventChan,
		ctx:      ctx,
	})
	return eventChan, nil
}

// Device returns generated.DeviceResolver implementation.
func (r *Resolver) Device() generated.DeviceResolver { return &deviceResolver{r} }

// Element returns generated.ElementResolver implementation.
func (r *Resolver) Element() generated.ElementResolver { return &elementResolver{r} }

// Group returns generated.GroupResolver implementation.
func (r *Resolver) Group() generated.GroupResolver { return &groupResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type deviceResolver struct{ *Resolver }
type elementResolver struct{ *Resolver }
type groupResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
