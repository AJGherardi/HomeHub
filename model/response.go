package model

// ControlResponse is the response to the ListControl query
type ControlResponse struct {
	Devices []Device
	Groups  []Group
}
