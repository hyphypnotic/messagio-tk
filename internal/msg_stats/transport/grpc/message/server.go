package msgstatsgrpc

import (
	"context"
	"github.com/hyphypnotic/messagio-tk/internal/msg_stats/services"
	pb "github.com/hyphypnotic/messagio-tk/protos/gen/go/msgstats"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	service services.MsgStats
	pb.UnimplementedMsgStatsServer
}

func Register(grpcServer *grpc.Server, service services.MsgStats) {
	pb.RegisterMsgStatsServer(grpcServer, &serverAPI{service: service})
}

// GetMsgStats retrieves message statistics for a given time range
func (s *serverAPI) GetMsgStats(ctx context.Context, req *pb.MsgStatsRequest) (*pb.MsgStatsResponse, error) {
	startTime := req.GetStartTime().AsTime()
	endTime := req.GetEndTime().AsTime()

	msgStats, err := s.service.GetMsgStats(ctx, startTime, endTime)
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
