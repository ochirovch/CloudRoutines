package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/digitalocean/godo"
	"github.com/ochirovch/cloudroutines/server"
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

	BinaryPayloadText, err := ioutil.ReadFile(server.BinaryPayload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	payload := fmt.Sprintf(string(BinaryPayloadText), Keeper.IPserver, Keeper.Name, Keeper.Name, Keeper.Name)

	log.Println(payload)
	instances := Keeper.Launch(
		server.VPSsettings{ProjectName: Keeper.Name, Cloud: server.DigitalOcean, Payload: payload},
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
	t := template.New("SourceCode")                    // Create a template.
	t, err := t.ParseFiles("html/payloads/sourcecode") // Parse template file.
	if err != nil {
		log.Println(err)
	}

	err = t.Execute(w, Keeper.IPserver) // merge.
	if err != nil {
		log.Println(err)
	}

}

func BinaryCodePayload(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "html/payloads/binary/")
}

//ChannelReceive - get results from vm
// vm, bundle, status
func ChannelReceive(w http.ResponseWriter, r *http.Request) {

}

//ChannelSend - send tasks
// vm, bundle, status
func ChannelSend(w http.ResponseWriter, r *http.Request) {
	var tasks = []server.Task{}
	var lastbundleTask = server.Task{}
	var lastbundle int
	// TODO: get last bucket
	Keeper.DB.Order("bundle desc").First(&lastbundleTask)
	lastbundle = lastbundleTask.Bundle
	lastbundle++

	Keeper.DB.Where("status = ?", server.NotProcessed).Find(&tasks)
	fmt.Fprintln(w, lastbundle)
	for _, task := range tasks {
		fmt.Fprintln(w, task.Url)
	}
}

//ChannelReceiveTask - receive urls for handling
// vm, bundle, status
func ChannelReceiveTask(w http.ResponseWriter, r *http.Request) {
	log.Println("!!!Uragsha!!!")
}

func refreshdata(k server.Keeper) {

	c := time.Tick(5 * time.Minute)
	for range c {
		k.LoadVPSes()
	}
}

func main() {
	var err error
	Keeper, err = server.LoadKeeper("settings.json")
	if err != nil {
		log.Println(err.Error())
		return
	}
	go refreshdata(Keeper)
	http.Handle("/payloads/", http.StripPrefix("/payloads/", http.FileServer(http.Dir("html/payloads/binary/"))))
	http.HandleFunc("/", Dashboard)
	http.HandleFunc("/node/add", AddNode)
	http.HandleFunc("/channel/sendresult", ChannelReceive)
	http.HandleFunc("/channel/gettask", ChannelSend)
	http.HandleFunc("/channel/addtask", ChannelReceiveTask)

	http.ListenAndServe(":8099", nil)
}
