package main

import (
	"context"
	"fmt"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	"golang.org/x/crypto/bcrypt"
)

var hostNode host.Host
var idht *dht.IpfsDHT
var db *badger.DB

//var appPeers <-chan peer.AddrInfo

func main() {

	initHost("bvvinai", "bvvinai@1357")
	fmt.Println("This node : ", hostNode.ID())
	connectToPeer("12D3KooWHeAvNK221WW7heHbrv6sgQf1FoPmucN2gbFTkCd2nt8T")
	connectToPeer("12D3KooWQFeTgsRRyRqeLGVNC76HTpvyJTZKvNSQEnBbxiLNqNGT")
	connectToPeer("12D3KooWJTxJUQwcYaq3n6AeUpD5GiTBohAPLL9KTxgxwWQSiANt")
	select {}
}

func connectToPeer(peerid string) {

	fmt.Println("Finding peer : ", peerid)
	peerID, err := peer.Decode(peerid)
	if err != nil {
		panic(err)
	}
	peerAddrInfo, err := idht.FindPeer(context.Background(), peerID)
	if err != nil {
		fmt.Println(err)
	}
	connerr := hostNode.Connect(context.Background(), peerAddrInfo)
	if connerr != nil {
		fmt.Println("No connection established!")
	} else {
		fmt.Println("Connected to node : ", peerid)
	}
}

func initHost(username string, password string) {

	db, _ = badger.Open(badger.DefaultOptions("./badger"))

	var hostKey crypto.PrivKey
	get_err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("priv"))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			hostKey, err = crypto.UnmarshalPrivateKey(val)
			return nil
		})
		return err
	})

	if get_err != nil || hostKey == nil {

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			panic(err)
		}

		priv, _, err := crypto.GenerateKeyPair(
			crypto.Ed25519,
			-1,
		)
		if err != nil {
			panic(err)
		}

		hostKey = priv

		privBytes, err := crypto.MarshalPrivateKey(priv)
		if err != nil {
			panic(err)
		}

		txn := db.NewTransaction(true)
		defer txn.Discard()

		txn.Set([]byte("username"), []byte(username))
		txn.Set([]byte("password"), hashedPassword)
		txn.Set([]byte("priv"), privBytes)
		txn.Commit()
	}

	hostNode, _ = libp2p.New(
		libp2p.Identity(hostKey),
		libp2p.NATPortMap(),
		libp2p.EnableNATService(),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			idht, _ = dht.New(context.Background(), h)
			return idht, nil
		}),
	)

	for _, addr := range dht.DefaultBootstrapPeers {
		pi, _ := peer.AddrInfoFromP2pAddr(addr)
		fmt.Println(pi)
		err := hostNode.Connect(context.Background(), *pi)
		if err != nil {
			fmt.Println(err)
		}
	}

	//broadcaster := crdt.NewPubsubBroadcaster(hostNode)

	// dhtDiscovery := drouting.NewRoutingDiscovery(idht)
	// go func() {
	// 	for {
	// 		_, err := dhtDiscovery.Advertise(context.Background(), "chatapp-bvvinai")
	// 		if err != nil {
	// 			panic(err)
	// 		}

	// 		appPeers, err = dhtDiscovery.FindPeers(context.Background(), "chatapp-bvvinai")
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 	}
	// }()

}
