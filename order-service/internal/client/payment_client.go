package client

import (
	"context"
	"fmt"
	"log"

	pb "github.com/RendersC/ap2-gen/payment"
	"github.com/crewsrenders/order-service/internal/usecase"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// PaymentGRPCClient wraps the generated gRPC client and satisfies usecase.PaymentClient.
type PaymentGRPCClient struct {
	client pb.PaymentServiceClient
	conn   *grpc.ClientConn
}

// NewPaymentGRPCClient dials the Payment Service using the address from env.
func NewPaymentGRPCClient(addr string) (*PaymentGRPCClient, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("dial payment service at %s: %w", addr, err)
	}
	log.Printf("Connected to Payment gRPC service at %s", addr)
	return &PaymentGRPCClient{
		client: pb.NewPaymentServiceClient(conn),
		conn:   conn,
	}, nil
}

// Close releases the underlying gRPC connection.
func (c *PaymentGRPCClient) Close() error {
	return c.conn.Close()
}

// GetPaymentStats calls the Payment Service to retrieve aggregated payment statistics.
func (c *PaymentGRPCClient) GetPaymentStats(ctx context.Context) (*usecase.PaymentStats, error) {
	resp, err := c.client.GetPaymentStats(ctx, &pb.GetPaymentStatsRequest{})
	if err != nil {
		return nil, fmt.Errorf("get payment stats: %w", err)
	}
	return &usecase.PaymentStats{
		TotalCount:      resp.TotalCount,
		AuthorizedCount: resp.AuthorizedCount,
		DeclinedCount:   resp.DeclinedCount,
		TotalAmount:     resp.TotalAmount,
	}, nil
}

// ProcessPayment calls the Payment Service and returns (paymentID, status, error).
func (c *PaymentGRPCClient) ProcessPayment(ctx context.Context, orderID, userID, currency string, amount float64) (string, string, error) {
	resp, err := c.client.ProcessPayment(ctx, &pb.PaymentRequest{
		OrderId:  orderID,
		UserId:   userID,
		Currency: currency,
		Amount:   amount,
	})
	if err != nil {
		st, _ := status.FromError(err)
		switch st.Code() {
		case codes.InvalidArgument:
			return "", "", fmt.Errorf("invalid payment request: %s", st.Message())
		case codes.Unavailable:
			return "", "", fmt.Errorf("payment service unavailable")
		default:
			return "", "", fmt.Errorf("payment service error: %s", st.Message())
		}
	}
	return resp.PaymentId, resp.Status, nil
}
