package handler

import (
	"strconv"
	
	"payment/internal/domain/entity"
)

// Helper functions for parsing query parameters
// These are shared across multiple handlers

func parsePaymentStatus(status string) *entity.PaymentStatus {
	switch status {
	case "pending":
		s := entity.PaymentStatusPending
		return &s
	case "processing":
		s := entity.PaymentStatusProcessing
		return &s
	case "completed":
		s := entity.PaymentStatusCompleted
		return &s
	case "failed":
		s := entity.PaymentStatusFailed
		return &s
	case "refunded":
		s := entity.PaymentStatusRefunded
		return &s
	case "cancelled":
		s := entity.PaymentStatusCancelled
		return &s
	}
	return nil
}

func parsePaymentMethod(method string) *entity.PaymentMethod {
	switch method {
	case "cash":
		m := entity.PaymentMethodCash
		return &m
	case "bank_transfer":
		m := entity.PaymentMethodBankTransfer
		return &m
	case "cod_cash":
		m := entity.PaymentMethodCODCash
		return &m
	case "cod_transfer":
		m := entity.PaymentMethodCODTransfer
		return &m
	case "digital_wallet":
		m := entity.PaymentMethodDigitalWallet
		return &m
	}
	return nil
}

func parsePaymentChannel(channel string) *entity.PaymentChannel {
	switch channel {
	case "loyverse_pos":
		c := entity.PaymentChannelLoyversePOS
		return &c
	case "saan_app":
		c := entity.PaymentChannelSAANApp
		return &c
	case "saan_chat":
		c := entity.PaymentChannelSAANChat
		return &c
	case "delivery":
		c := entity.PaymentChannelDelivery
		return &c
	case "web_portal":
		c := entity.PaymentChannelWebPortal
		return &c
	}
	return nil
}

func parsePaymentTiming(timing string) *entity.PaymentTiming {
	switch timing {
	case "prepaid":
		t := entity.PaymentTimingPrepaid
		return &t
	case "cod":
		t := entity.PaymentTimingCOD
		return &t
	}
	return nil
}

func parseInt(str string) (int, error) {
	return strconv.Atoi(str)
}
