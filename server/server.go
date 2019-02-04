package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

type Keeper struct {
	Name   string
	Tokens map[string]string
	VPS    []VPS
}

type VPS interface {
	Launch(VPSsettings) []Instance
	GetName() string
}

type Instance struct {
	Cloud   string
	Project string
	Name    string
}

type VPSsettings struct {
	ProjectName string
	Cloud       string
	Region      string
	Type        string
	Image       string
	Token       string
	Payload     string
	Amount      int
	//
}

type VPSGoogleComputeEngine struct {
	Name     string
	Settings VPSsettings
}

type VPSDigitalOcean struct {
	Name     string
	Settings VPSsettings
	Droplets []godo.Droplet
}
type configuration struct {
	Name   string `json:"Name"`
	Tokens []struct {
		Cloud string `json:"Cloud"`
		Token string `json:"Token"`
	} `json:"Tokens"`
}

type TokenSource struct {
	AccessToken string
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

const (
	GoogleComputeEngine = "GoogleComputeEngine"
	DigitalOcean        = "DigitalOcean"
	//not safe without secure connection
	SourceCodePayload = "/payloads/SourceCode"
	BinaryPayload     = "/payloads/Binary"
)

func (v *VPSGoogleComputeEngine) Launch(VPSsettings) (instances []Instance) {
	return
}

func (v *VPSDigitalOcean) getNewNames(ProjectName string, Amount int, client *godo.Client, ctx context.Context) []string {
	var names = []string{}

	opt := &godo.ListOptions{
		Page:    1,
		PerPage: 200,
	}

	mostBig := 0
	droplets, _, err := client.Droplets.ListByTag(ctx, ProjectName, opt)
	if err != nil {
		log.Println(err)
	}

	v.Droplets = append(v.Droplets, droplets...)

	for _, droplet := range droplets {
		strInt := strings.Replace(droplet.Name, ProjectName, "", -1)
		convInt, err := strconv.Atoi(strInt)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if convInt > mostBig {
			mostBig = convInt
		}
	}

	for i := 1; i <= Amount; i++ {
		names = append(names, ProjectName+strconv.Itoa(mostBig+i))
	}
	return names
}

func (v *VPSDigitalOcean) GetName() string {
	return v.Name
}
func (v *VPSGoogleComputeEngine) GetName() string {
	return v.Name
}

// Launch use for launching one or few examples of vps
// Count of names is count of instance
// by default use most low cost instance
//
func (v *VPSDigitalOcean) Launch(setting VPSsettings) (instances []Instance) {
	tokenSource := &TokenSource{
		AccessToken: setting.Token,
	}
	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	client := godo.NewClient(oauthClient)
	ctx := context.TODO()

	// set default values
	if setting.Region == "" {
		setting.Region = "nyc1"
	}
	if setting.Type == "" {
		setting.Type = "s-1vcpu-1gb"
	}
	if setting.Image == "" {
		setting.Image = "ubuntu-16-04-x64"
	}
	if setting.Amount == 0 {
		setting.Amount = 1
	}
	v.Droplets = v.Droplets[:0]
	names := v.getNewNames(setting.ProjectName, setting.Amount, client, ctx)

	createRequest := &godo.DropletMultiCreateRequest{
		Names:  names,
		Region: setting.Region,
		Size:   setting.Type,
		Image: godo.DropletCreateImage{
			Slug: setting.Image,
		},
		IPv6: true,
		Tags: []string{setting.ProjectName, "ClouRoutines"},
	}

	droplets, _, err := client.Droplets.CreateMultiple(ctx, createRequest)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("%+v\n", setting)
		return
	}

	for _, droplet := range droplets {
		instances = append(instances, Instance{Cloud: setting.Cloud, Project: setting.ProjectName, Name: droplet.Name})
	}

	v.Droplets = append(v.Droplets, droplets...)
	fmt.Printf("%+v\n", v.Droplets)
	return instances
}

func (k *Keeper) getVPS(vpsname string) (VPS, error) {
	for _, vps := range k.VPS {
		if vps.GetName() == vpsname {
			return vps, nil
		}
	}
	return nil, errors.New("can not find this vps: " + vpsname)
}

func (k *Keeper) Launch(settings ...VPSsettings) (instances []Instance) {
	for _, setting := range settings {
		if k.Tokens[setting.Cloud] == "" {
			log.Println("set up Token for ", setting.Cloud)
			continue
		}
		setting.Token = k.Tokens[setting.Cloud]

		vps, err := k.getVPS(setting.Cloud)
		if err != nil {
			fmt.Println(err)
			continue
		}
		instances = append(vps.Launch(setting), instances...)
	}
	return instances
}

func LoadKeeper(path string) (k Keeper, err error) {

	file, _ := os.Open(path)
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := configuration{}
	err = decoder.Decode(&configuration)
	if err != nil {
		log.Println("error decoding json:", err)
	}
	k.Name = configuration.Name
	k.VPS = append(k.VPS, &VPSDigitalOcean{Name: DigitalOcean})
	k.VPS = append(k.VPS, &VPSGoogleComputeEngine{Name: GoogleComputeEngine})

	k.Tokens = make(map[string]string)
	for _, conf := range configuration.Tokens {
		log.Println(conf.Cloud)
		k.Tokens[conf.Cloud] = conf.Token
	}
	k.loadVPSes()
	return k, nil
}

func (k *Keeper) loadVPSes() {
	tokenSource := &TokenSource{
		AccessToken: k.Tokens[DigitalOcean],
	}
	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	client := godo.NewClient(oauthClient)
	ctx := context.TODO()
	//droplets, _, err := client.Droplets.ListByTag(ctx, k.Name, &godo.ListOptions{Page: 1, PerPage: 200})
	droplets, _, err := client.Droplets.List(ctx, &godo.ListOptions{Page: 1, PerPage: 200})
	if err != nil {
		log.Println(err)
		return
	}

	for _, vps := range k.VPS {
		log.Printf("%+v\n", vps)

		switch x := vps.(type) {
		case *VPSDigitalOcean:
			x.Droplets = droplets
		case *VPSGoogleComputeEngine:
		default:
			fmt.Printf("Unsupported type: %T\n", x)
		}

	}

}
