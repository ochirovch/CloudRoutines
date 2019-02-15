package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/digitalocean/godo"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"golang.org/x/oauth2"
)

type Keeper struct {
	Name     string
	Tokens   map[string]string
	IPserver string
	VPS      []VPS
	DB       *gorm.DB
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
	Name     string `json:"Name"`
	IPserver string `json:"IPserver"`
	Tokens   []struct {
		Cloud string `json:"Cloud"`
		Token string `json:"Token"`
	} `json:"Tokens"`
	DB struct {
		Name     string `json:"Name"`
		Password string `json:"Password"`
		Username string `json:"Username"`
	} `json:"DB"`
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

type Task struct {
	gorm.Model
	Url string
	Bundle int
	Status int
}

const (
	GoogleComputeEngine = "GoogleComputeEngine"
	DigitalOcean        = "DigitalOcean"
	//not safe without secure connection
	SourceCodePayload = "html/payloads/SourceCode"
	BinaryPayload     = "html/payloads/BinaryCode"
)

//statuses for processing data
const (
	NotProcessed = iota
	Sent 
	Processed
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
		Tags: []string{setting.ProjectName, "CloudRoutines"},
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
	conf := configuration{}
	err = decoder.Decode(&conf)
	if err != nil {
		log.Println("error decoding json:", err)
	}
	k.Name = conf.Name
	k.VPS = append(k.VPS, &VPSDigitalOcean{Name: DigitalOcean})
	k.VPS = append(k.VPS, &VPSGoogleComputeEngine{Name: GoogleComputeEngine})

	k.IPserver = conf.IPserver
	k.Tokens = make(map[string]string)
	for _, conf := range conf.Tokens {
		log.Println(conf.Cloud)
		k.Tokens[conf.Cloud] = conf.Token
	}
	defer db.Close()
	k.DB, err = gorm.Open("postgres", "host="+conf.DB.Name+" port=5432 user="+conf.DB.Username+" dbname=cloudroutines password="+conf.DB.Password+) +":"+conf.DB.Password+"@tcp("+conf.DB.Name+":5432)/cloudroutines")
	if err != nil {
		log.Fatal(err)
	}
	defer k.DB.Close()

	k.loadIPserver()
	k.LoadVPSes()
	return k, nil
}

func (k *Keeper) loadIPserver() {
	if k.IPserver != "" {
		return
	}
	for _, vps := range k.VPS {
		log.Printf("%+v\n", vps)

		switch x := vps.(type) {
		case *VPSDigitalOcean:
			resp, err := http.Get("http://169.254.169.254/metadata/v1/interfaces/private/0/ipv4/address")
			if err != nil {
				// handle error
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			k.IPserver = string(body)
		case *VPSGoogleComputeEngine:
		default:
			fmt.Printf("Unsupported type: %T\n", x)
		}
	}

}

func (k *Keeper) LoadVPSes() {
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
