package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
)

const ipFile = "./ips.txt"
const urlFile = "./urls.txt"

func main() {
	ipFName := flag.String("f", ipFile, "The file which holds the previous assigned IP addresses.")
	urlFName := flag.String("c", urlFile, "JSON file with the URLs to call.")
	flag.Parse()

	var checkIps []string
	var urls []string
	ips := make(map[string]bool)
	confChanged := false

	decode(*ipFName, &ips)
	err := decode(*urlFName, &urls)
	if err != nil {
		log.Fatal(err)
		return
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
		return
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Fatal(err)
			return
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPAddr:
				ip = v.IP
			case *net.IPNet:
				ip = v.IP
			}
			if ip != nil && ip.To4() == nil && ip.IsGlobalUnicast() {
				checkIps = append(checkIps, ip.String())
			}
		}
	}

	confChanged = len(checkIps) != len(ips)
	if !confChanged {
		for _, ip := range checkIps {
			if !ips[ip] {
				confChanged = true
				break
			}
		}
	}

	if confChanged {
		fmt.Println("Network configuration changed")
		fmt.Println("Old IPs:", ips)
		newIps := make(map[string]bool)
		for _, ip := range checkIps {
			newIps[ip] = true
		}
		fmt.Println("New IPs", newIps)
		err = encode(*ipFName, newIps)
		if err != nil {
			fmt.Println(err)
		}

		for _, url := range urls {
			resp, err := update(url)
			if err != nil {
				log.Println(err)
			} else {
				fmt.Println(resp)
			}
		}
	}
}

func update(url string) (body string, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	fmt.Println(resp.Status)
	bbody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return string(bbody), err
}

func decode(filename string, s interface{}) error {
	f, err := os.Open(filename)
	defer f.Close()
	if err != nil {
		return err
	}
	err = json.NewDecoder(f).Decode(s)
	return err
}

func encode(filename string, s interface{}) error {
	f, err := os.Create(filename)
	defer f.Close()
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	return enc.Encode(s)
}
