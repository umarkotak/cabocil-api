package order_service

import (
	"context"
	"database/sql"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/contract"
	"github.com/umarkotak/ytkidd-api/contract_resp"
	"github.com/umarkotak/ytkidd-api/model"
	"github.com/umarkotak/ytkidd-api/repos/order_repo"
	"github.com/umarkotak/ytkidd-api/repos/product_repo"
	"github.com/umarkotak/ytkidd-api/repos/user_repo"
	"github.com/umarkotak/ytkidd-api/utils/payment_lib"
)

func CreateOrder(ctx context.Context, params contract.CreateOrder) (contract_resp.CreateOrder, error) {
	user, err := user_repo.GetByGuid(ctx, params.UserGuid)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return contract_resp.CreateOrder{}, err
	}

	product, err := product_repo.GetByCode(ctx, params.ProductCode)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return contract_resp.CreateOrder{}, err
	}

	order := model.Order{
		UserID:         user.ID,
		OrderType:      product.BenefitType,
		Description:    product.Name,
		Status:         model.ORDER_STATUS_INITIALIZED,
		PaymentStatus:  payment_lib.STATUS_INITIALIZED,
		BasePrice:      product.BasePrice,
		Price:          product.Price,
		DiscountAmount: 0,
		FinalPrice:     product.Price,
		PaymentNumber:  sql.NullString{},
		Metadata: model.OrderMetadata{
			ProductCode: product.Code,
		},
	}
	order.GenNumber()

	createPaymentData, err := payment_lib.CreatePayment(ctx, nil, payment_lib.CreatePaymentParams{
		PaymentPlatform: payment_lib.PAYMENT_PLATFORM_MIDTRANS,
		PaymentType:     payment_lib.PAYMENT_TYPE_MIDTRANS,
		OrderNumber:     order.Number,
		OrderItems: []payment_lib.OrderItem{
			{
				ID:       product.Code,
				Price:    product.Price,
				Quantity: 1,
				Name:     product.Name,
			},
		},
		CustomerFirstName: user.Name,
		CustomerEmail:     user.Email,
		Amount:            order.FinalPrice,
	})
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return contract_resp.CreateOrder{}, err
	}

	order.PaymentNumber = sql.NullString{createPaymentData.Number, true}

	order.ID, err = order_repo.Insert(ctx, nil, order)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return contract_resp.CreateOrder{}, err
	}

	return contract_resp.CreateOrder{
		OrderNumber:                order.Number,
		MidtransSnapToken:          createPaymentData.MidtransSnapToken,
		MidtransSnapUrl:            createPaymentData.MidtransSnapUrl,
		MidtransSandboxPaymentPage: "https://simulator.sandbox.midtrans.com/v2/qris/index",
		MidtransNotificationPage:   "https://dashboard.sandbox.midtrans.com/settings/vtweb_configuration/history",
	}, nil
}
