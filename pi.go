package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	//"time"

	"encoding/json"
	"io/ioutil"

	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"github.com/currantlabs/ble"
	"github.com/currantlabs/ble/linux"

	"gopkg.in/resty.v1"

	"math"
)

var (
	device = flag.String("device", "default", "implementation of ble")
	//du     = flag.Duration("du", 0, "scanning duration")
	dup = flag.Bool("dup", true, "allow duplicate reported")
)

type Packet struct {
	ID   string
	RSSI int
}

var ids = make(chan Packet, 1)
var pid string
var x float64
var y float64

func loop() {
	for {
		pack := <-ids
		id := pack.ID

		resp, err := resty.R().Get("http://omaraa.ddns.net:62027/db/beacons/" + id)
		if err != nil {
			fmt.Println(err)
		}
		if resp.StatusCode() == 200 {
			var b map[string]interface{}
			json.Unmarshal([]byte(resp.String()), &b)
			xx := math.Pow(x-b["xpos"].(float64), 2)
			yy := math.Pow(y-b["ypos"].(float64), 2)
			dis := math.Sqrt(xx + yy)

			offset := math.Log10(dis)*25 + float64(pack.RSSI)
			fmt.Printf("Device:%s, dis:%f, rssi:%i, offset:%f\n", id, dis, pack.RSSI, offset)
			putDevice(id, offset, dis)
		}
	}
}

func putDevice(id string, offset float64, dis float64) {
	_, err := resty.R().
		SetFormData(map[string]string{
			"offset": fmt.Sprintf("%f", offset),
			"distance": fmt.Sprintf("%f", dis),
		}).
		Put("http://omaraa.ddns.net:62027/db/beacons/" + id)

	if err != nil {
		panic(err)
	}

	fmt.Println(map[string]interface{}{
		"_id":    id,
		"offset": offset,
	})
}

func main() {

	file, _ := ioutil.ReadFile("pi.json")

	var data map[string]interface{}

	json.Unmarshal(file, &data)

	pid = data["id"].(string)
	resp, err := resty.R().Get("http://omaraa.ddns.net:62027/db/pies/" + pid)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode() == 200 {
		var b map[string]interface{}
		json.Unmarshal([]byte(resp.String()), &b)
		x = b["xpos"].(float64)
		y = b["ypos"].(float64)
	}

	fmt.Println(data)
	flag.Parse()

	d, err := linux.NewDevice()
	if err != nil {
		log.Fatalf("can't new device : %s", err)
	}
	ble.SetDefaultDevice(d)

	go loop()

	// Scan for specified durantion, or until interrupted by user.
	fmt.Printf("Scanning for infinity...\n")
	//ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), *du))
	chkErr(ble.Scan(context.Background(), *dup, advHandler, nil))
}

var devices = map[string]int{}
var devicesRSum = map[string]int{}

func advHandler(a ble.Advertisement) {
	if len(a.ServiceData()) > 0 {
		data := a.ServiceData()[0].Data

		if len(data) > 0 {
			if data[0]&0x0F == 0x02 {
				id := hex.EncodeToString(data[1:9])
				if _, ok := devices[id]; !ok {
					devices[id] = 1
					devicesRSum[id] = a.RSSI()
				}

				devices[id] = devices[id] + 1
				devicesRSum[id] = devicesRSum[id] + a.RSSI()
				ids <- Packet{ID: id, RSSI: devicesRSum[id] / devices[id]}

			}
		}
	}
}

func chkErr(err error) {
	switch errors.Cause(err) {
	case nil:
	case context.DeadlineExceeded:
		fmt.Printf("done\n")
	case context.Canceled:
		fmt.Println("cancled")
	default:
		log.Fatalf(err.Error())
	}
}
