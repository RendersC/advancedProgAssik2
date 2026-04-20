package grpc

import (
	"context"
	"log"

	pb "github.com/RendersC/ap2-gen/payment"
	"github.com/crewsrenders/payment-service/internal/domain"
	"github.com/crewsrenders/payment-service/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// PaymentGRPCServer implements the generated PaymentServiceServer interface.
// It delegates all business logic to the use case layer.
type PaymentGRPCServer struct {
	pb.UnimplementedPaymentServiceServer
	useCase *usecase.PaymentUseCase
}

func NewPaymentGRPCServer(uc *usecase.PaymentUseCase) *PaymentGRPCServer {
	return &PaymentGRPCServer{useCase: uc}
}

// GetPaymentStats handles an incoming gRPC GetPaymentStats RPC call.
func (s *PaymentGRPCServer) GetPaymentStats(ctx context.Context, _ *pb.GetPaymentStatsRequest) (*pb.PaymentStats, error) {
	stats, err := s.useCase.GetPaymentStats(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get payment stats: %v", err)
	}
	return &pb.PaymentStats{
		TotalCount:      stats.TotalCount,
		AuthorizedCount: stats.AuthorizedCount,
		DeclinedCount:   stats.DeclinedCount,
		TotalAmount:     stats.TotalAmount,
	}, nil
}

// ProcessPayment handles an incoming gRPC ProcessPayment RPC call.
func (s *PaymentGRPCServer) ProcessPayment(ctx context.Context, req *pb.PaymentRequest) (*pb.PaymentResponse, error) {
	log.Printf("ProcessPayment called for order_id=%s amount=%.2f %s", req.OrderId, req.Amount, req.Currency)

	if req.OrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}
	if req.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be greater than zero")
	}

	domainReq := &domain.Payment{
		OrderID:  req.OrderId,
		UserID:   req.UserId,
		Amount:   req.Amount,
		Currency: req.Currency,
	}

	result, err := s.useCase.ProcessPayment(ctx, domainReq)
	if err != nil {
		log.Printf("ProcessPayment error: %v", err)
		return nil, status.Errorf(codes.Internal, "payment processing failed: %v", err)
	}

	return &pb.PaymentResponse{
		PaymentId:   result.ID,
		Status:      result.Status,
		ProcessedAt: timestamppb.New(result.ProcessedAt),
		Message:     result.Message,
	}, nil
}
