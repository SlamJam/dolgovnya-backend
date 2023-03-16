package connect_handlers

import (
	"context"

	"github.com/SlamJam/dolgovnya-backend/internal/pb"
	"github.com/SlamJam/dolgovnya-backend/internal/pb/pbconnect"
	"github.com/bufbuild/connect-go"
)

type SplitTheBillServiceHandler struct {
	pbconnect.UnimplementedSplitTheBillServiceHandler
}

func (h *SplitTheBillServiceHandler) NewBillSplit(context.Context, *connect.Request[pb.SplitRequest]) (*connect.Response[pb.SplitResponse], error) {
	// return nil, connect.NewError(connect.CodeUnimplemented, errors.New("SplitTheBillService.NewBillSplit is not implemented"))
	resp := connect.NewResponse(&pb.SplitResponse{
		Id: 11334353,
	})

	return resp, nil
}
