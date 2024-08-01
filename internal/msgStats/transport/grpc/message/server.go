package msgstatsgrpc

import (
	"context"
	"github.com/hyphypnotic/messagio-tk/internal/msgStats/app"
	"github.com/hyphypnotic/messagio-tk/internal/msgStats/entity"
	"github.com/hyphypnotic/messagio-tk/internal/msgStats/services"
	pb "github.com/hyphypnotic/messagio-tk/protos/gen/go/msgstats"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MsgStats interface {
	GetStats(ctx context.Context, req entity.Message) (entity.MsgStats, error)
}

type serverAPI struct {
	app *app.Application
	pb.UnimplementedMsgStatsServer
	msgstats MsgStats
	service  services.MsgStats
}

func Register(gRPCServer *grpc.Server, msgStats MsgStats, service services.MsgStats, app *app.Application) {
	pb.RegisterMsgStatsServer(gRPCServer, &serverAPI{msgstats: msgStats, service: service, app: app})
}

// GetStats retrieves message statistics for a given time range
func (s *serverAPI) GetStats(req *pb.MsgStatsRequest) (*pb.MsgStatsResponse, error) {
	s.app.Logger.Info("Received GetStats request",
		zap.String("startTime", req.StartTime.AsTime().String()),
		zap.String("endTime", req.EndTime.AsTime().String()),
	)

	startTime := req.StartTime.AsTime()
	endTime := req.EndTime.AsTime()

	// Delegate to the service layer
	msgStats, err := s.service.GetMsgStats(startTime, endTime)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get message statistics: %v", err)
	}

	return &pb.MsgStatsResponse{
		Count: msgStats.Count,
		StatusStats: &pb.MsgStatsResponse_StatusStats{
			SuccessCount: msgStats.Status.SuccessCount,
			ErrorCount:   msgStats.Status.ErrorCount,
		},
	}, nil
}
