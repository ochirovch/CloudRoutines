package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ochirovch/CollyRoutines/server"
)

var Keeper server.Keeper

// Dashboard show info about vps and statistics for calculations
func Dashboard(w http.ResponseWriter, r *http.Request) {

	for _, vps := range Keeper.VPS {
		switch x := vps.(type) {
		case *server.VPSDigitalOcean:

			for _, droplet := range x.Droplets {
				fmt.Fprintf(w, "Droplet names: %s\n", droplet.Name)

			}
		case *server.VPSGoogleComputeEngine:
		default:
			fmt.Printf("Unsupported type: %T\n", x)
		}
	}
}

func AddNode(w http.ResponseWriter, r *http.Request) {

	Keeper.Launch(
		server.VPSsettings{Cloud: server.GoogleComputeEngine, Type: "mid-1", Payload: server.SourceCodePayload},
		server.VPSsettings{Cloud: server.DigitalOcean, Type: "mid-2", Payload: server.BinaryPayload},
	)
}

func DeleteNode(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)
}

func Download(w http.ResponseWriter, r *http.Request) {

}

func SourceCodePayload(w http.ResponseWriter, r *http.Request) {

}

func BinaryCodePayload(w http.ResponseWriter, r *http.Request) {

}

func main() {
	var err error
	Keeper, err = server.LoadKeeper("settings.json")
	if err != nil {
		log.Println(err.Error())
		return
	}
	http.HandleFunc("/", Dashboard)
	http.HandleFunc("/node/add", AddNode)
	http.HandleFunc("/node/delete", DeleteNode)
	http.HandleFunc("/payload/sourcecode", SourceCodePayload)
	http.HandleFunc("/payload/binarycode", BinaryCodePayload)
	http.HandleFunc("/payload/download", Download)

	http.ListenAndServe(":8099", nil)

}
