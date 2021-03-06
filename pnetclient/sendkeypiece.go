package pnetclient

import (
	"errors"
	"fmt"

	"golang.org/x/net/context"

	pb "github.com/pp2p/paranoid/proto/paranoidnetwork"
	raftpb "github.com/pp2p/paranoid/proto/raft"
	"github.com/pp2p/pfsd/globals"
	"github.com/pp2p/pfsd/keyman"
)

// SendKeyPiece to the node specified by the UUID
func SendKeyPiece(uuid string, generation int64, piece *keyman.KeyPiece, addElement bool) error {
	node, err := globals.Nodes.GetNode(uuid)
	if err != nil {
		return errors.New("could not find node details")
	}

	conn, err := Dial(node)
	if err != nil {
		Log.Error("SendKeyPiece: failed to dial ", node)
		return fmt.Errorf("failed to dial: %s", node)
	}
	defer conn.Close()

	client := pb.NewParanoidNetworkClient(conn)

	thisNodeProto := &pb.Node{
		Ip:         globals.ThisNode.IP,
		Port:       globals.ThisNode.Port,
		CommonName: globals.ThisNode.CommonName,
		Uuid:       globals.ThisNode.UUID,
	}
	keyProto := &pb.KeyPiece{
		Data:              piece.Data,
		ParentFingerprint: piece.ParentFingerprint[:],
		Prime:             piece.Prime.Bytes(),
		Seq:               piece.Seq,
		Generation:        generation,
		OwnerNode:         thisNodeProto,
	}

	resp, err := client.SendKeyPiece(context.Background(), &pb.KeyPieceSend{
		Key:        keyProto,
		AddElement: addElement,
	})
	if err != nil {
		Log.Error("Failed sending KeyPiece to", node, "Error:", err)
		return fmt.Errorf("Failed sending key piece to %s, Error: %s", node, err)
	}

	if resp.ClientMustCommit && addElement {
		raftThisNodeProto := &raftpb.Node{
			Ip:         globals.ThisNode.IP,
			Port:       globals.ThisNode.Port,
			CommonName: globals.ThisNode.CommonName,
			NodeId:     globals.ThisNode.UUID,
		}
		raftOwnerNode := &raftpb.Node{
			Ip:         keyProto.GetOwnerNode().Ip,
			Port:       keyProto.GetOwnerNode().Port,
			CommonName: keyProto.GetOwnerNode().CommonName,
			NodeId:     keyProto.GetOwnerNode().Uuid,
		}
		err := globals.RaftNetworkServer.RequestKeyStateUpdate(raftThisNodeProto, raftOwnerNode, generation)
		if err != nil {
			Log.Errorf("failed to commit to Raft: %s", err)
			return fmt.Errorf("failed to commit to Raft: %s", err)
		}
	}

	return nil
}
