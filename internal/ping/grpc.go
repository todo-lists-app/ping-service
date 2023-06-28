package ping

import (
	"context"
	"github.com/hashicorp/vault/sdk/helper/pointerutil"
	"github.com/todo-lists-app/ping-service/internal/config"
	pb "github.com/todo-lists-app/protobufs/generated/ping/v1"
)

type Server struct {
	pb.UnimplementedPingServiceServer
	*config.Config
}

func (s *Server) Ping(ctx context.Context, r *pb.LastUserPingRequest) (*pb.PingResponse, error) {
	p := NewPingService(ctx, *s.Config, r.GetUserId())
	ping, err := p.GetPing()
	if err != nil {
		return &pb.PingResponse{
			UserId:   r.GetUserId(),
			LastPing: 0,
			Status:   pointerutil.StringPtr(err.Error()),
		}, nil
	}

	return &pb.PingResponse{
		UserId:   r.GetUserId(),
		LastPing: ping.Time.Unix(),
	}, nil
}
