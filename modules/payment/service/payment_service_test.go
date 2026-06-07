package service

import (
	"crypto/sha512"
	"encoding/hex"
	"testing"

	"github.com/Rizal-Nurochman/matchnbuild/modules/payment/dto"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/constants"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifySignature(t *testing.T) {
	serverKey := "SB-Mid-server-test"
	notification := dto.NotificationRequest{
		OrderID:     "MNB-order-1",
		StatusCode:  "200",
		GrossAmount: "150000.00",
	}

	hash := sha512.Sum512([]byte(
		notification.OrderID +
			notification.StatusCode +
			notification.GrossAmount +
			serverKey,
	))
	notification.SignatureKey = hex.EncodeToString(hash[:])

	assert.True(t, verifySignature(notification, serverKey))

	notification.SignatureKey = "invalid"
	assert.False(t, verifySignature(notification, serverKey))
}

func TestAmountToIDR(t *testing.T) {
	amount, err := amountToIDR(decimal.NewFromInt(150000))
	require.NoError(t, err)
	assert.Equal(t, int64(150000), amount)

	_, err = amountToIDR(decimal.RequireFromString("150000.50"))
	assert.ErrorIs(t, err, dto.ErrInvalidAmount)
}

func TestNormalizePaymentStatus(t *testing.T) {
	assert.Equal(t, constants.PAYMENT_STATUS_CAPTURE, normalizePaymentStatus("capture"))
	assert.Equal(t, constants.PAYMENT_STATUS_SETTLEMENT, normalizePaymentStatus("settlement"))
	assert.Equal(t, constants.PAYMENT_STATUS_EXPIRE, normalizePaymentStatus("expire"))
	assert.Equal(t, constants.PAYMENT_STATUS_PENDING, normalizePaymentStatus("unknown"))
}

func TestIsPaid(t *testing.T) {
	assert.True(t, isPaid("settlement", ""))
	assert.True(t, isPaid("capture", "accept"))
	assert.False(t, isPaid("capture", "challenge"))
	assert.False(t, isPaid("pending", ""))
}
