package pnetclient

import (
	"github.com/cpssd/paranoid/pfsd/globals"
	"github.com/cpssd/paranoid/pfsd/upnp"
	pb "github.com/cpssd/paranoid/proto/paranoidnetwork"
	"golang.org/x/net/context"
	"strconv"
)

func Ping() {
	ip, err := upnp.GetIP()
	if err != nil {
		Log.Fatal("Can not ping peers: unable to get IP. Error:", err)
	}

	nodes := globals.Nodes.GetAll()
	for _, node := range nodes {
		port := strconv.Itoa(globals.Port)

		conn, err := Dial(node)
		if err != nil {
			Log.Error("Ping: failed to dial ", node)
		}
		defer conn.Close()

		client := pb.NewParanoidNetworkClient(conn)

		_, err = client.Ping(context.Background(), &pb.PingRequest{
			Ip:         ip,
			Port:       port,
			CommonName: globals.CommonName,
			Uuid:       globals.UUID,
		})
		if err != nil {
			Log.Error("Can't ping ", node)
		}
	}
}
