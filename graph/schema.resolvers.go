package graph

import (
	"context"
	"crypto/rand"
	"errors"
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

func (r *groupResolver) Addr(ctx context.Context, obj *model.Group) (string, error) {
	return utils.EncodeBase64(obj.Addr), nil
}

func (r *mutationResolver) AddDevice(ctx context.Context, addr string, devUUID string, name string) (*model.Device, error) {
	// Provision device
	uuid := utils.DecodeBase64(devUUID)
	r.Controller.Provision(uuid)
	// Wait for node added
	nodeAddr := <-r.NodeAdded
	// Create device object
	device := model.MakeDevice(
		name,
		"2PowerSwitch",
		nodeAddr,
		r.DB,
	)
	// Get group
	groupAddr := utils.DecodeBase64(addr)
	group := r.DB.GetGroupByAddr(groupAddr)
	// Get model id
	if true {
		// Set type and add elements
		device.AddElem("onoff", r.DB)
		device.AddElem("onoff", r.DB)
	}
	// Configure Device
	r.Controller.ConfigureNode(device.Addr, group.KeyIndex)
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

func (r *mutationResolver) SetState(ctx context.Context, devAddr string, elemNumber int, value string) (*model.State, error) {
	state := utils.DecodeBase64(value)
	address := utils.DecodeBase64(devAddr)
	device := r.DB.GetDeviceByAddr(address)
	// Get appKey from group
	group := r.DB.GetGroupByDevAddr(device.Addr)
	// Send State
	if device.GetState(elemNumber).StateType == "onoff" {
		// Send msg
		r.Controller.SendMessage(
			state[0],
			device.GetElemAddr(elemNumber),
			group.KeyIndex,
		)
	}
	modelState := device.GetState(elemNumber)
	return &modelState, nil
}

func (r *queryResolver) AvailableDevices(ctx context.Context) ([]string, error) {
	uuids := make([]string, 0)
	for _, uuid := range *r.UnprovisionedNodes {
		b64 := utils.EncodeBase64(uuid)
		uuids = append(uuids, b64)
	}
	return uuids, nil
}

func (r *queryResolver) GetState(ctx context.Context, devAddr string, elemNumber int) (*model.State, error) {
	address := utils.DecodeBase64(devAddr)
	device := r.DB.GetDeviceByAddr(address)
	state := device.GetState(elemNumber)
	return &state, nil
}

func (r *queryResolver) ListControl(ctx context.Context) (*model.ControlResponse, error) {
	rsp := model.ControlResponse{
		Groups:  r.DB.GetGroups(),
		Devices: r.DB.GetDevices(),
	}
	return &rsp, nil
}

func (r *stateResolver) State(ctx context.Context, obj *model.State) (string, error) {
	return utils.EncodeBase64(obj.State), nil
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

// State returns generated.StateResolver implementation.
func (r *Resolver) State() generated.StateResolver { return &stateResolver{r} }

type deviceResolver struct{ *Resolver }
type elementResolver struct{ *Resolver }
type groupResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type stateResolver struct{ *Resolver }
