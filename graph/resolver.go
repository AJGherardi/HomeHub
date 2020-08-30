package graph

import (
	mesh "github.com/AJGherardi/GoMeshController"
	"github.com/AJGherardi/HomeHub/generated"
	"github.com/AJGherardi/HomeHub/model"
	"github.com/grandcat/zeroconf"
)

// Resolver is the root of the schema
type Resolver struct {
	DB         model.DB
	Controller mesh.Controller
	NodeAdded  chan []byte
	Mdns       *zeroconf.Server
}

// New returns the servers config
func New(db model.DB, controller mesh.Controller, nodeAdded chan []byte, mdns *zeroconf.Server) generated.Config {
	c := generated.Config{
		Resolvers: &Resolver{DB: db, Controller: controller, NodeAdded: nodeAdded, Mdns: mdns},
	}
	return c
}
