package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"log"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	mesh "github.com/AJGherardi/GoMeshController"
	"github.com/AJGherardi/HomeHub/generated"
	"github.com/AJGherardi/HomeHub/graph"
	"github.com/AJGherardi/HomeHub/model"
	"github.com/AJGherardi/HomeHub/utils"
	"github.com/gorilla/websocket"
	"github.com/grandcat/zeroconf"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	mdns               *zeroconf.Server
	unprovisionedNodes = new([][]byte)
	nodeAdded          = make(chan []byte)
	stateChanged       = make(chan []byte)
	controller         mesh.Controller
)

func main() {
	// Get db
	db := model.OpenDB()
	// Open Mesh Controller and defer close
	controller = mesh.Open()
	defer controller.Close()
	// Generate a cert
	if _, err := os.Stat("cert.pem"); err != nil {
		if os.IsNotExist(err) {
			writeCert()
		}
	}
	// Check if configured
	if db.GetNetData().ID == primitive.NilObjectID {
		// Setup the mdns service
		mdns, _ = zeroconf.Register("unprovisioned", "_alexandergherardi._tcp", "local.", 8080, nil, nil)
	} else {
		mdns, _ = zeroconf.Register("hub", "_alexandergherardi._tcp", "local.", 8080, nil, nil)
	}
	// Serve the schema
	schema, updateState, publishEvents, resolver := graph.New(db, controller, nodeAdded, mdns, unprovisionedNodes)
	srv := handler.New(
		generated.NewExecutableSchema(
			schema,
		),
	)
	// Register read functions
	go controller.Read(
		// onSetupStatus
		func() {},
		// onAddKeyStatus
		func(appIdx []byte) {},
		// onUnprovisionedBeacon
		func(uuid []byte) {
			*unprovisionedNodes = append(*unprovisionedNodes, uuid)
		},
		// onNodeAdded
		func(addr []byte) {
			nodeAdded <- addr
		},
		// onState
		func(addr []byte, state byte) {
			// Update device state
			device := db.GetDeviceByElemAddr(addr)
			device.UpdateState(addr, []byte{state}, db)
			updateState()
		},
		// onEvent
		func(addr []byte) {
			publishEvents(addr)
		},
	)
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		InitFunc: func(ctx context.Context, initPayload transport.InitPayload) (context.Context, error) {
			netData := db.GetNetData()
			// If not configured auth is not required
			if netData.ID == primitive.NilObjectID {
				return ctx, nil
			}
			// If pin is available check it
			if resolver.UserPin != 000000 {
				if initPayload["pin"] != nil {
					if initPayload["pin"].(float64) == float64(resolver.UserPin) {
						return ctx, nil
					}
				}
			}
			// Check web key
			webKey := utils.DecodeBase64(initPayload["webKey"].(string))
			verify := netData.CheckWebKey(webKey)
			// If valid continue
			if verify {
				return ctx, nil
			}
			// Else error
			return ctx, errors.New("Bad webKey")
		},
	})
	srv.Use(extension.Introspection{})
	http.Handle("/graphql", srv)
	http.ListenAndServeTLS(":2041", "cert.pem", "key.pem", nil)
}

func writeCert() {
	// Make private key
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	// Set key usage
	keyUsage := x509.KeyUsageDigitalSignature
	// Set time limits
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	// Make serial number
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, serialNumberLimit)
	// Create cert template
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Home"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	// Set ca to true
	template.IsCA = true
	template.KeyUsage |= x509.KeyUsageCertSign
	// Create the cert
	derBytes, _ := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	// Write cert to file
	certOut, _ := os.Create("cert.pem")
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()
	log.Print("wrote cert.pem\n")
	// Write key to file
	keyOut, _ := os.OpenFile("key.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	privBytes, _ := x509.MarshalPKCS8PrivateKey(priv)
	pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	keyOut.Close()
	log.Print("wrote key.pem\n")
}
