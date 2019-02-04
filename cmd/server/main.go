package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/digitalocean/godo"
	"github.com/ochirovch/CollyRoutines/server"
)

var Keeper server.Keeper

// Dashboard show info about vps and statistics for calculations
func Dashboard(w http.ResponseWriter, r *http.Request) {

	Droplets := []godo.Droplet{}
	t := template.New("index.html")           // Create a template.
	t, err := t.ParseFiles("html/index.html") // Parse template file.
	if err != nil {
		log.Println(err)
	}
	for _, vps := range Keeper.VPS {
		switch x := vps.(type) {
		case *server.VPSDigitalOcean:
			Droplets = x.Droplets
		case *server.VPSGoogleComputeEngine:
		default:
			fmt.Printf("Unsupported type: %T\n", x)
		}
	}

	err = t.Execute(w, Droplets) // merge.
	if err != nil {
		log.Println(err)
	}

}

func AddNode(w http.ResponseWriter, r *http.Request) {

	instances := Keeper.Launch(
		server.VPSsettings{ProjectName: "colly", Cloud: server.DigitalOcean, Payload: server.BinaryPayload},
	)
	jsonInstances, err := json.Marshal(instances)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonInstances)
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

//ChannelReceive - get results from vm
// vm, bundle, status
func ChannelReceive(w http.ResponseWriter, r *http.Request) {

}

//ChannelSend - send tasks
// vm, bundle, status
func ChannelSend(w http.ResponseWriter, r *http.Request) {

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
	http.HandleFunc("/channel/receive", ChannelReceive)
	http.HandleFunc("/channel/send", ChannelSend)

	http.ListenAndServe(":8099", nil)

}
