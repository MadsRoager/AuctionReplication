package main

import (
	context "context"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/MadsRoager/AuctionReplication/proto"
	"google.golang.org/grpc"
)

type Node struct {
	proto.UnimplementedServerServer
	serverNodePorts []int32
}

var highestBid int32 = 0
var highestBidder = "No bidder yet"
var highestBidderID int32 = 0
var isAuctionOver = false

var mutex = sync.Mutex{}
var nodePort = flag.Int("port", 8081, "server port number")

func main() {
	go setupNode()
	defer runAuction()
}

func setupNode() {
	flag.Parse()
	node := &Node{
		serverNodePorts: make([]int32, 4),
	}
	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(*nodePort))
	if err != nil {
		log.Fatalln("could not start listener")
	}
	proto.RegisterServerServer(grpcServer, node)
	serverError := grpcServer.Serve(listener)
	if serverError != nil {
		log.Fatalln("could not start server")
	}
}

func runAuction() {
	for {
		log.Println("An auction has started!")
		startNewAuction()
		time.Sleep(time.Second * 20)
		endAuction()
		time.Sleep(time.Second * 5)
	}
}

func startNewAuction() {
	highestBid = 0
	highestBidder = "No bidder yet"
	highestBidderID = 0
	isAuctionOver = false
}

func endAuction() {
	isAuctionOver = true
}

func (node *Node) UpdateHighestBid(ctx context.Context, bid *proto.BidRequest) (*proto.Ack, error) {
	fmt.Println("it gets into update highest bid in node")
	if isAuctionOver {
		fmt.Println("auction is over")
		return fail(), nil
	}
	if isWinningBet(bid) {
		fmt.Println("is winning bid")
		updateHighestBid(bid)
		return success(), nil
	}
	fmt.Println("faillll")
	return fail(), nil
}

func isWinningBet(bid *proto.BidRequest) bool {
	if bid.Amount > highestBid {
		return true
	}
	if bid.Amount == highestBid && bid.ProcessID < highestBidderID {
		return true
	}
	return false
}

func updateHighestBid(bid *proto.BidRequest) {
	mutex.Lock()
	highestBid = bid.Amount
	highestBidder = bid.Name
	mutex.Unlock()
}

func (node *Node) GetHighestBid(ctx context.Context, in *proto.Void) (*proto.BidResult, error) {
	result := &proto.BidResult{
		Amount:        highestBid,
		Name:          highestBidder,
		AuctionStatus: getAuctionStatus(),
	}
	return result, nil
}

func getAuctionStatus() string {
	if isAuctionOver {
		return "Auction has ended."
	}
	return "Auction is ongoing."
}

func success() *proto.Ack {
	return &proto.Ack{
		Ack: "Success",
	}
}

func fail() *proto.Ack {
	return &proto.Ack{
		Ack: "Fail",
	}
}
