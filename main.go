package main

import (
	"fmt"
	"github.com/cpssd/paranoid/pfsd/globals"
	"github.com/cpssd/paranoid/pfsd/icserver"
	"github.com/cpssd/paranoid/pfsd/pnetclient"
	"github.com/cpssd/paranoid/pfsd/pnetserver"
	pb "github.com/cpssd/paranoid/proto/paranoidnetwork"
	"github.com/prestonTao/upnp"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Print("Usage:\n\tpfsd <paranoid_directory> <Discovery Server> <Discovery Port>\n")
		os.Exit(1)
	}
	discoveryPort, err := strconv.Atoi(os.Args[3])
	if err != nil || discoveryPort < 1 || discoveryPort > 65535 {
		log.Fatalln("FATAL: Discovery port must be a number between 1 and 65535, inclusive.")
	}
	pnetserver.ParanoidDir = os.Args[1]
	go startIcAndListen(pnetserver.ParanoidDir)
	globals.UpnpMapping = new(upnp.Upnp)

	globals.Server, err = pnetclient.GetIP()
	if err != nil {
		log.Fatalln("FATAL: Cant get external IP. Error : ", err)
	}

	if _, err := os.Stat(pnetserver.ParanoidDir); os.IsNotExist(err) {
		log.Fatalln("FATAL: path", pnetserver.ParanoidDir, "does not exist.")
	}
	if _, err := os.Stat(path.Join(pnetserver.ParanoidDir, "meta")); os.IsNotExist(err) {
		log.Fatalln("FATAL: path", pnetserver.ParanoidDir, "is not valid PFS root.")
	}

	//Asking for port 0 requests a random free port from the OS.
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("FATAL: Failed to start listening : %v.\n", err)
	}

	splits := strings.Split(lis.Addr().String(), ":")
	port, err := strconv.Atoi(splits[len(splits)-1])
	if err != nil {
		log.Fatalln("Could not parse port", splits[len(splits)-1], " Error :", err)
	}

	if globals.UpnpMapping.Active {
		log.Println("Upnp mapping active")
		err = globals.UpnpMapping.AddPortMapping(port, port, "TCP")
		if err != nil {
			log.Fatalln("Could not add Upnp port mapping. Error :", err)
		}
	} else {
		log.Println("Upnp mapping not active")
	}

	pnetserver.SetDiscovery(os.Args[2], os.Args[3], strconv.Itoa(port))
	globals.Port = port

	pnetserver.JoinDiscovery("_")
	srv := grpc.NewServer()
	pb.RegisterParanoidNetworkServer(srv, &pnetserver.ParanoidServer{})
	srv.Serve(lis)
}

func startIcAndListen(pfsDir string) {
	go icserver.RunServer(pfsDir, true)

	for {
		select {
		case message := <-icserver.MessageChan:
			pnetclient.SendRequest(message)
		}
	}
}
