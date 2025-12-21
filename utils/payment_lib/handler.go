package payment_lib

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/midtrans/midtrans-go"
	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/utils/render"
)

func MidtransCallbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var midtransNotification MidtransNotification

	err := json.NewDecoder(r.Body).Decode(&midtransNotification)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	err = midtransNotification.ValidateSignature(midtrans.ServerKey)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	payment, err := GetByOrderNumber(ctx, midtransNotification.OrderID)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	payment.PaymentType = midtransNotification.PaymentType
	payment.ReferenceStatus = sql.NullString{midtransNotification.TransactionStatus, true}
	payment.ReferenceNumber = sql.NullString{midtransNotification.ReferenceID, midtransNotification.ReferenceID != ""}
	payment.FraudStatus = sql.NullString{midtransNotification.FraudStatus, true}
	payment.SyncStatus()

	err = Update(ctx, nil, payment)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	// TODO: decouple this process
	err = ProcessOrderBenefit(ctx, payment.OrderNumber)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		render.Error(w, r, err, "")
		return
	}

	render.Response(w, r, http.StatusOK, nil)
}
