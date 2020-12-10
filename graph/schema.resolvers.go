package graph

import (
	"context"
	"crypto/rand"
	"errors"
	math "math/rand"

	"github.com/AJGherardi/HomeHub/cmd"
	"github.com/AJGherardi/HomeHub/generated"
	"github.com/AJGherardi/HomeHub/model"
	"github.com/AJGherardi/HomeHub/utils"
)

func (r *deviceResolver) Elements(ctx context.Context, obj *model.Device) ([]*generated.ElementResponse, error) {
	return toElementResponseSlice(obj.Elements), nil
}

func (r *elementResolver) State(ctx context.Context, obj *model.Element) (string, error) {
	return utils.EncodeBase64(obj.State), nil
}

func (r *groupResolver) Scenes(ctx context.Context, obj *model.Group) ([]*generated.SceneResponse, error) {
	return toSceneResponseSlice(obj.Scenes), nil
}

func (r *groupResolver) Devices(ctx context.Context, obj *model.Group) ([]*generated.DeviceResponse, error) {
	return toDeviceResponseSlice(obj.Devices), nil
}

func (r *mutationResolver) ConfigHub(ctx context.Context) (string, error) {
	if utils.CheckIfConfigured() {
		return "", errors.New("already configured")
	}
	webKey := cmd.ConfigHub(r.Store, r.Controller)
	return utils.EncodeBase64(webKey), nil
}

func (r *mutationResolver) ResetHub(ctx context.Context) (bool, error) {
	if !utils.CheckIfConfigured() {
		return false, errors.New("not configured")
	}
	cmd.ResetHub(r.Store, r.Controller)
	return true, nil
}

func (r *mutationResolver) AddDevice(ctx context.Context, groupAddr int, devUUID string, name string) (int, error) {
	// Provision device
	uuid := utils.DecodeBase64(devUUID)
	nodeAddr := cmd.AddDevice(r.Store, r.Controller, name, uuid, uint16(groupAddr), r.NodeAdded)
	return int(nodeAddr), nil
}

func (r *mutationResolver) RemoveDevice(ctx context.Context, devAddr int, groupAddr int) (int, error) {
	cmd.RemoveDevice(r.Store, r.Controller, uint16(groupAddr), uint16(devAddr))
	return devAddr, nil
}

func (r *mutationResolver) AddGroup(ctx context.Context, name string) (int, error) {
	groupAddr := cmd.AddGroup(r.Store, name)
	return int(groupAddr), nil
}

func (r *mutationResolver) RemoveGroup(ctx context.Context, addr int) (int, error) {
	cmd.RemoveGroup(r.Store, r.Controller, uint16(addr))
	return addr, nil
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

func (r *mutationResolver) SetState(ctx context.Context, groupAddr int, elemAddr int, value string) (bool, error) {
	state := utils.DecodeBase64(value)
	cmd.SetState(r.Store, r.Controller, state, uint16(groupAddr), uint16(elemAddr))
	return true, nil
}

func (r *mutationResolver) SceneStore(ctx context.Context, name string, groupAddr int) (int, error) {
	sceneNumber := cmd.SceneStore(r.Store, r.Controller, uint16(groupAddr), name)
	return int(sceneNumber), nil
}

func (r *mutationResolver) SceneRecall(ctx context.Context, sceneNumber int, groupAddr int) (int, error) {
	cmd.SceneRecall(r.Store, r.Controller, uint16(groupAddr), uint16(sceneNumber))
	return sceneNumber, nil
}

func (r *mutationResolver) SceneDelete(ctx context.Context, sceneNumber int, groupAddr int) (int, error) {
	cmd.SceneDelete(r.Store, r.Controller, uint16(groupAddr), uint16(sceneNumber))
	return sceneNumber, nil
}

func (r *mutationResolver) EventBind(ctx context.Context, sceneNumber int, groupAddr int, devAddr int, elemAddr int) (int, error) {
	cmd.EventBind(r.Store, r.Controller, uint16(groupAddr), uint16(devAddr), uint16(elemAddr), uint16(sceneNumber))
	return sceneNumber, nil
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

func (r *subscriptionResolver) ListGroup(ctx context.Context, groupAddr int) (<-chan *generated.GroupResponse, error) {
	groupChan := make(chan *generated.GroupResponse, 1)
	// Put initial result in chan
	group := r.Store.Groups[uint16(groupAddr)]
	groupChan <- &generated.GroupResponse{
		Addr:  groupAddr,
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
