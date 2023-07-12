package validate

import (
	"context"
	"github.com/todo-lists-app/ping-service/internal/config"
	pb "github.com/todo-lists-app/protobufs/generated/id_checker/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Validate struct {
	Config *config.Config
	CTX    context.Context
}

func NewValidate(config *config.Config, ctx context.Context) *Validate {
	return &Validate{
		Config: config,
		CTX:    ctx,
	}
}

func (v *Validate) ValidateUser(userId, token string) (bool, error) {
	if v.Config.Development {
		return true, nil
	}

	conn, err := grpc.DialContext(v.CTX, v.Config.Identity.Service, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return false, err
	}
	defer conn.Close()

	g := pb.NewIdCheckerServiceClient(conn)
	resp, err := g.CheckId(v.CTX, &pb.CheckIdRequest{
		Id:          userId,
		AccessToken: accessToken,
	})
	if err != nil {
		return false, err
	}
	if !resp.GetIsValid() {
		return false, nil
	}

	return true, nil
}
