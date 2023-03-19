package connect_handlers

import (
	"context"

	"github.com/SlamJam/dolgovnya-backend/internal/app/models"
	"github.com/SlamJam/dolgovnya-backend/internal/app/services"
	split_the_billv1 "github.com/SlamJam/dolgovnya-backend/internal/pb/dolgovnya/split_the_bill/v1"
	"github.com/SlamJam/dolgovnya-backend/internal/pb/dolgovnya/split_the_bill/v1/split_the_billv1connect"
	"github.com/bufbuild/connect-go"
)

type SplitTheBillServiceHandler struct {
	split_the_billv1connect.UnimplementedSplitTheBillServiceHandler
	service *services.SplitTheBillService
}

func NewSplitTheBillServiceHandler() *SplitTheBillServiceHandler {
	return &SplitTheBillServiceHandler{}
}

// var rules := NewRules(rulez.AUTHZ)
func (h *SplitTheBillServiceHandler) NewBillSplit(ctx context.Context, req *connect.Request[split_the_billv1.NewBillRequest]) (*connect.Response[split_the_billv1.NewBillResponse], error) {
	userID, err := userIDFromRequest(req)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	// TODO:
	// err := checkPermissions(userID, rules)
	// if err != nil {
	// 	return nil, connect.NewError(connect.CodePermissionDenied, err)
	// }

	// DTO -> domain model
	bill := models.Bill{}
	bill.Items = make([]models.BillItem, 0, len(req.Msg.Items))
	for _, item := range req.Msg.Items {
		billItem := models.BillItem{
			Title: item.Title,
			// PricePerOne: ,
			// Quantity: item.Quantity.,
			// Type: uint8(item.Type), // if item.Type > 255 raise
		}
		for _, share := range item.Shares {
			// _ = share
			billItem.Shares = append(billItem.Shares, models.BillShare{
				UserID: models.UserID(share.UserId),
				Share:  uint32(share.Share),
			})
		}

		bill.Items = append(bill.Items, billItem)
	}

	for _, payment := range req.Msg.Payments {
		bill.Payments = append(bill.Payments, models.BillPayment{
			UserID: models.UserID(payment.UserId),
			// Amount: payment.Amount,
		})
	}

	billID, err := h.service.SaveBill(ctx, userID, bill)

	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := connect.NewResponse(&split_the_billv1.NewBillResponse{
		BillId: uint64(billID),
	})

	return resp, nil
}
