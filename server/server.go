package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

type Keeper struct {
	Name   string
	Tokens map[string]string
	VPS    []VPS
}

type VPS interface {
	Launch(VPSsettings)
}

type VPSsettings struct {
	Names   []string
	Cloud   string
	Region  string
	Type    string
	Image   string
	Token   string
	Payload string
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

func (v *VPSGoogleComputeEngine) Launch(VPSsettings) {

}

// Launch use for launching one or few examples of vps
// Count of names is count of instance
// by default use most low cost instance
//
func (v *VPSDigitalOcean) Launch(setting VPSsettings) {
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

	createRequest := &godo.DropletMultiCreateRequest{
		Names:  setting.Names,
		Region: setting.Region,
		Size:   setting.Type,
		Image: godo.DropletCreateImage{
			Slug: setting.Image,
		},
		IPv6: true,
		Tags: []string{"CollyRoutines"},
	}

	droplets, _, err := client.Droplets.CreateMultiple(ctx, createRequest)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("%+v\n", setting)
		return
	}
	v.Droplets = droplets
	fmt.Printf("%+v\n", v.Droplets)
}

func (k *Keeper) Launch(settings ...VPSsettings) {
	for _, setting := range settings {
		if k.Tokens[setting.Cloud] == "" {
			log.Println("set up Token for ", setting.Cloud)
			continue
		}
		setting.Token = k.Tokens[setting.Cloud]

		switch setting.Cloud {
		case GoogleComputeEngine:
			vps := &VPSGoogleComputeEngine{}
			vps.Launch(setting)
			k.VPS = append(k.VPS, vps)
		case DigitalOcean:
			vps := &VPSDigitalOcean{}
			vps.Launch(setting)
			k.VPS = append(k.VPS, vps)
		default:
			vps := &VPSDigitalOcean{}
			vps.Launch(setting)
			k.VPS = append(k.VPS, vps)
		}
	}
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
	log.Println(tokenSource.AccessToken)
	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
	client := godo.NewClient(oauthClient)
	ctx := context.TODO()

	droplets, _, err := client.Droplets.List(ctx, &godo.ListOptions{Page: 1, PerPage: 200}) //.ListByTag(ctx, k.Name, &godo.ListOptions{Page: 1, PerPage: 200})
	if err != nil {
		log.Println(err)
		return
	}

	k.VPS = append(k.VPS, &VPSDigitalOcean{})
	k.VPS = append(k.VPS, &VPSGoogleComputeEngine{})

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
