package graph

import (
	"context"
	"errors"
	math "math/rand"

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
	if r.Network.CheckIfConfigured() {
		return "", errors.New("Already configured")
	}
	webKey, configErr := r.Network.ConfigHub()
	return utils.EncodeBase64(webKey), configErr
}

func (r *mutationResolver) ResetHub(ctx context.Context) (bool, error) {
	if !r.Network.CheckIfConfigured() {
		return false, errors.New("Not configured")
	}
	r.Network.ResetHub()
	return true, nil
}

func (r *mutationResolver) AddDevice(ctx context.Context, groupAddr int, devUUID string, name string) (int, error) {
	// Provision device
	uuid := utils.DecodeBase64(devUUID)
	nodeAddr, addErr := r.Network.AddDevice(name, uuid, uint16(groupAddr))
	return int(nodeAddr), addErr
}

func (r *mutationResolver) RemoveDevice(ctx context.Context, devAddr int, groupAddr int) (int, error) {
	removeErr := r.Network.RemoveDevice(uint16(groupAddr), uint16(devAddr))
	return devAddr, removeErr
}

func (r *mutationResolver) AddGroup(ctx context.Context, name string) (int, error) {
	groupAddr, addErr := r.Network.AddGroup(name)
	return int(groupAddr), addErr
}

func (r *mutationResolver) RemoveGroup(ctx context.Context, groupAddr int) (int, error) {
	removeErr := r.Network.RemoveGroup(uint16(groupAddr))
	return groupAddr, removeErr
}

func (r *mutationResolver) AddUser(ctx context.Context) (string, error) {
	// Remove user pin
	r.UserPin = 000000
	webKey, addErr := r.Network.AddAccessKey()
	return utils.EncodeBase64(webKey), addErr
}

func (r *mutationResolver) SetState(ctx context.Context, groupAddr int, elemAddr int, value string) (bool, error) {
	state := utils.DecodeBase64(value)
	r.Network.SetState(state, uint16(groupAddr), uint16(elemAddr))
	return true, nil
}

func (r *mutationResolver) SceneStore(ctx context.Context, name string, groupAddr int) (int, error) {
	sceneNumber, addErr := r.Network.SceneStore(uint16(groupAddr), name)
	return int(sceneNumber), addErr
}

func (r *mutationResolver) SceneRecall(ctx context.Context, sceneNumber int, groupAddr int) (int, error) {
	recallErr := r.Network.SceneRecall(uint16(groupAddr), uint16(sceneNumber))
	return sceneNumber, recallErr
}

func (r *mutationResolver) SceneDelete(ctx context.Context, sceneNumber int, groupAddr int) (int, error) {
	removeErr := r.Network.SceneDelete(uint16(groupAddr), uint16(sceneNumber))
	return sceneNumber, removeErr
}

func (r *mutationResolver) EventBind(ctx context.Context, sceneNumber int, groupAddr int, devAddr int, elemAddr int) (int, error) {
	bindErr := r.Network.EventBind(uint16(groupAddr), uint16(devAddr), uint16(elemAddr), uint16(sceneNumber))
	return sceneNumber, bindErr
}

func (r *queryResolver) AvailableDevices(ctx context.Context) ([]string, error) {
	uuids := make([]string, 0)
	// Encode uuids of unprovisioned nodes as base64
	for _, uuid := range r.Network.GetUnprovisionedNodes() {
		b64 := utils.EncodeBase64(uuid)
		uuids = append(uuids, b64)
	}
	return uuids, nil
}

func (r *queryResolver) AvailableGroups(ctx context.Context) ([]*generated.GroupResponse, error) {
	groups := r.Network.GetGroups()
	return toGroupResponseSlice(groups), nil
}

func (r *queryResolver) GetUserPin(ctx context.Context) (int, error) {
	// Generate a 6 digit random number
	pin := math.Intn(1000000)
	r.UserPin = pin
	return r.UserPin, nil
}

func (r *subscriptionResolver) WatchGroup(ctx context.Context, groupAddr int) (<-chan *generated.GroupResponse, error) {
	// Make a chan for updates
	groupChan := make(chan *generated.GroupResponse, 1)
	// Put initial result in chan
	group, getErr := r.Network.GetGroup(uint16(groupAddr))
	if getErr != nil {
		return nil, getErr
	}
	groupChan <- toGroupResponse(uint16(groupAddr), group)
	return groupChan, nil
}

func (r *subscriptionResolver) WatchState(ctx context.Context, groupAddr int, devAddr int, elemAddr int) (<-chan string, error) {
	stateChan := make(chan string, 1)
	// Put initial result in chan
	state, readErr := r.Network.ReadState(uint16(groupAddr), uint16(devAddr), uint16(elemAddr))
	if readErr != nil {
		return nil, readErr
	}
	// Add observer
	r.StateObservers = append(
		r.StateObservers,
		stateObserver{
			groupAddr: uint16(groupAddr),
			devAddr:   uint16(devAddr),
			elemAddr:  uint16(elemAddr),
			messages:  stateChan,
			ctx:       ctx,
		},
	)
	stateChan <- utils.EncodeBase64(state)
	return stateChan, nil
}

func (r *subscriptionResolver) WatchEvents(ctx context.Context) (<-chan int, error) {
	eventChan := make(chan int, 1)
	r.EventObservers = append(
		r.EventObservers,
		eventObserver{
			messages: eventChan,
			ctx:      ctx,
		},
	)
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
