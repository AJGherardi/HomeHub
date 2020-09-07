package graph

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/AJGherardi/HomeHub/generated"
	"github.com/AJGherardi/HomeHub/model"
	"github.com/AJGherardi/HomeHub/utils"
	"github.com/grandcat/zeroconf"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (r *deviceResolver) Addr(ctx context.Context, obj *model.Device) (string, error) {
	return utils.EncodeBase64(obj.Addr), nil
}

func (r *elementResolver) Addr(ctx context.Context, obj *model.Element) (string, error) {
	return utils.EncodeBase64(obj.Addr), nil
}

func (r *elementResolver) State(ctx context.Context, obj *model.Element) (string, error) {
	return utils.EncodeBase64(obj.State), nil
}

func (r *groupResolver) Addr(ctx context.Context, obj *model.Group) (string, error) {
	return utils.EncodeBase64(obj.Addr), nil
}

func (r *groupResolver) DevAddrs(ctx context.Context, obj *model.Group) ([]string, error) {
	addrs := make([]string, 0)
	for _, addr := range obj.GetDevAddrs() {
		b64 := utils.EncodeBase64(addr)
		addrs = append(addrs, b64)
	}
	return addrs, nil
}

func (r *mutationResolver) AddDevice(ctx context.Context, addr string, devUUID string, name string) (*model.Device, error) {
	// Provision device
	uuid := utils.DecodeBase64(devUUID)
	r.Controller.Provision(uuid)
	// Wait for node added
	nodeAddr := <-r.NodeAdded
	// Create device object
	device := model.MakeDevice(
		"2PowerSwitch",
		nodeAddr,
		r.DB,
	)
	// Get group
	groupAddr := utils.DecodeBase64(addr)
	group := r.DB.GetGroupByAddr(groupAddr)
	// Configure Device
	r.Controller.ConfigureNode(device.Addr, group.KeyIndex)
	time.Sleep(100 * time.Millisecond)
	if true {
		// Set type and add elements
		elemAddr0 := device.AddElem(name+"-0", "onoff", r.DB)
		r.Controller.ConfigureElem(group.Addr, device.Addr, elemAddr0, group.KeyIndex)
		time.Sleep(100 * time.Millisecond)
		elemAddr1 := device.AddElem(name+"-1", "onoff", r.DB)
		r.Controller.ConfigureElem(group.Addr, device.Addr, elemAddr1, group.KeyIndex)
	}
	// Add device to group
	group.AddDevice(device.Addr, r.DB)
	return &device, nil
}

func (r *mutationResolver) RemoveDevice(ctx context.Context, addr string) (*model.Device, error) {
	// Get devKey
	devAddr := utils.DecodeBase64(addr)
	device := r.DB.GetDeviceByAddr(devAddr)
	// Send reset payload
	r.Controller.ResetNode(devAddr)
	// Remove device from database
	r.DB.DeleteDevice(devAddr)
	// Remove devAddr from group
	group := r.DB.GetGroupByDevAddr(devAddr)
	group.RemoveDevice(device.Addr, r.DB)
	return &device, nil
}

func (r *mutationResolver) RemoveGroup(ctx context.Context, addr string) (*model.Group, error) {
	// Get groupAddr
	groupAddr := utils.DecodeBase64(addr)
	group := r.DB.GetGroupByAddr(groupAddr)
	// Delete devices
	for _, devAddr := range group.GetDevAddrs() {
		device := r.DB.GetDeviceByAddr(devAddr)
		// Send reset payload
		r.Controller.ResetNode(device.Addr)
		// Remove device from database
		r.DB.DeleteDevice(devAddr)
	}
	// Remove the group
	r.DB.DeleteGroup(groupAddr)
	return &group, nil
}

func (r *mutationResolver) AddGroup(ctx context.Context, name string) (*model.Group, error) {
	netData := r.DB.GetNetData()
	// Get net values
	keyIndex := netData.GetNextAppKeyIndex()
	groupAddr := netData.GetNextGroupAddr()
	// Add an app key
	r.Controller.AddKey(keyIndex)
	// Add a group
	group := model.MakeGroup(name, groupAddr, keyIndex, r.DB)
	// Update net data
	netData.IncrementNextGroupAddr(r.DB)
	netData.IncrementNextAppKeyIndex(r.DB)
	return &group, nil
}

func (r *mutationResolver) ConfigHub(ctx context.Context) (string, error) {
	// Check if configured
	if r.DB.GetNetData().ID != primitive.NilObjectID {
		return "", errors.New("already configured")
	}
	// Stop the mdns server
	r.Mdns.Shutdown()
	// Make a web key
	webKey := make([]byte, 16)
	rand.Read(webKey)
	// Clean house
	r.DB.DeleteAll()
	// Add and get net data
	model.MakeNetData(webKey, r.DB)
	// Setup controller
	r.Controller.Setup()
	return utils.EncodeBase64(webKey), nil
}

func (r *mutationResolver) ResetHub(ctx context.Context) (bool, error) {
	// Check if configured
	if r.DB.GetNetData().ID == primitive.NilObjectID {
		return false, errors.New("not configured")
	}
	// Clean house
	r.DB.DeleteAll()
	// Reset mesh controller
	r.Controller.Reset()
	time.Sleep(time.Second)
	r.Controller.Reboot()
	// Start Mdns
	r.Mdns, _ = zeroconf.Register("hub", "_alexandergherardi._tcp", "local.", 8080, nil, nil)
	return true, nil
}

func (r *mutationResolver) SetState(ctx context.Context, addr string, value string) (bool, error) {
	state := utils.DecodeBase64(value)
	address := utils.DecodeBase64(addr)
	device := r.DB.GetDeviceByElemAddr(address)
	// Get appKey from group
	group := r.DB.GetGroupByDevAddr(device.Addr)
	// Send State
	if true {
		// Send msg
		r.Controller.SendMessage(
			state[0],
			address,
			group.KeyIndex,
		)
	}
	return true, nil
}

func (r *mutationResolver) SceneStore(ctx context.Context, addr string) (string, error) {
	address := utils.DecodeBase64(addr)
	group := r.DB.GetGroupByAddr(address)
	netData := r.DB.GetNetData()
	// Get and incrment next scene number
	sceneNumber := netData.GetNextSceneNumber()
	netData.IncrementNextSceneNumber(r.DB)
	// Store scene
	r.Controller.SendStoreMessage(sceneNumber, address, group.KeyIndex)
	return utils.EncodeBase64(sceneNumber), nil
}

func (r *mutationResolver) SceneRecall(ctx context.Context, sceneNumber string, addr string) (string, error) {
	address := utils.DecodeBase64(addr)
	sceneNumberBytes := utils.DecodeBase64(sceneNumber)
	group := r.DB.GetGroupByAddr(address)
	netData := r.DB.GetNetData()
	// Get and increment next scene number
	netData.IncrementNextSceneNumber(r.DB)
	// Store scene
	r.Controller.SendRecallMessage(sceneNumberBytes, address, group.KeyIndex)
	return sceneNumber, nil
}

func (r *mutationResolver) SceneDelete(ctx context.Context, sceneNumber string, addr string) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) AvailableDevices(ctx context.Context) ([]string, error) {
	uuids := make([]string, 0)
	for _, uuid := range *r.UnprovisionedNodes {
		b64 := utils.EncodeBase64(uuid)
		uuids = append(uuids, b64)
	}
	return uuids, nil
}

func (r *queryResolver) AvailableGroups(ctx context.Context) ([]*model.Group, error) {
	groups := r.DB.GetGroups()
	groupPointers := make([]*model.Group, 0)
	for i := range groups {
		groupPointers = append(groupPointers, &groups[i])
	}
	return groupPointers, nil
}

func (r *subscriptionResolver) ListGroup(ctx context.Context, addr string) (<-chan []*model.Device, error) {
	address := utils.DecodeBase64(addr)
	groupChan := make(chan []*model.Device, 1)
	// Put initial result in chan
	group := r.DB.GetGroupByAddr(address)
	devicePointers := make([]*model.Device, 0)
	devices := r.DB.GetDevices()
	for i, device := range devices {
		for _, devAddr := range group.DevAddrs {
			if reflect.DeepEqual(devAddr, device.Addr) {
				devicePointers = append(devicePointers, &devices[i])
			}
		}
	}
	groupChan <- devicePointers
	return groupChan, nil
}

func (r *subscriptionResolver) GetState(ctx context.Context, addr string) (<-chan string, error) {
	address := utils.DecodeBase64(addr)
	stateChan := make(chan string, 1)
	r.Observers = append(r.Observers, observer{
		addr:     address,
		messages: stateChan,
		ctx:      ctx,
	})
	device := r.DB.GetDeviceByElemAddr(address)
	state := device.GetState(address)
	stateChan <- utils.EncodeBase64(state)
	return stateChan, nil
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
