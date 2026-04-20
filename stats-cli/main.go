// stats-cli calls the Payment Service GetPaymentStats RPC and prints results.
// Usage:
//
//	go run main.go -addr localhost:50051
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	pb "github.com/RendersC/ap2-gen/payment"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	addr := flag.String("addr", "localhost:50051", "Payment Service gRPC address")
	flag.Parse()

	conn, err := grpc.NewClient(*addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to Payment Service at %s: %v", *addr, err)
	}
	defer conn.Close()

	client := pb.NewPaymentServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stats, err := client.GetPaymentStats(ctx, &pb.GetPaymentStatsRequest{})
	if err != nil {
		log.Fatalf("GetPaymentStats failed: %v", err)
	}

	fmt.Println("========== Payment Statistics ==========")
	fmt.Printf("  Total payments:      %d\n", stats.TotalCount)
	fmt.Printf("  Authorized (SUCCESS):%d\n", stats.AuthorizedCount)
	fmt.Printf("  Declined (FAILED):   %d\n", stats.DeclinedCount)
	fmt.Printf("  Total amount (cents):%d\n", stats.TotalAmount)
	fmt.Println("========================================")
}
