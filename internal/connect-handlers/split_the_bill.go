package connect_handlers

import (
	"context"

	"github.com/SlamJam/dolgovnya-backend/internal/app/models"
	"github.com/SlamJam/dolgovnya-backend/internal/app/services"
	"github.com/SlamJam/dolgovnya-backend/internal/pb"
	"github.com/SlamJam/dolgovnya-backend/internal/pb/pbconnect"
	"github.com/bufbuild/connect-go"
)

type SplitTheBillServiceHandler struct {
	pbconnect.UnimplementedSplitTheBillServiceHandler
	service *services.SplitTheBillService
}

func NewSplitTheBillServiceHandler() *SplitTheBillServiceHandler {
	return &SplitTheBillServiceHandler{}
}

func (h *SplitTheBillServiceHandler) NewBillSplit(ctx context.Context, req *connect.Request[pb.SplitRequest]) (*connect.Response[pb.SplitResponse], error) {
	userID, err := userIDFromRequest(req)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	// checkPermissions()

	bill := models.Bill{}

	bill.Items = make([]models.BillItem, 0, len(req.Msg.Items))
	for _, item := range req.Msg.Items {
		bill.Items = append(bill.Items, models.BillItem{
			Title: item.Title,
		})
	}

	billID, err := h.service.SaveBill(ctx, userID, bill)

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := connect.NewResponse(&pb.SplitResponse{
		Id: uint64(billID),
	})

	return resp, nil
}
