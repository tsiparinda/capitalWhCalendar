package logic

import (
	"capitalbank/logger"
	"capitalbank/pbapi"

	// "capitalbank/logger"

	"capitalbank/store"

	"github.com/sirupsen/logrus"
)

func StartExchangePayments() {

	payments := []store.Payment{}

	// send privat
	err := store.LoadPaymentsPrivat(&payments)
	if err != nil {
		logger.Log.Info("StartExchangePayments: Error from LoadPaymentsPrivat:", err.Error())
		return
	}
	logger.Log.WithFields(logrus.Fields{
		"payments": payments,
	}).Trace("StartExchangePayments: Payment loaded by LoadPaymentsPrivat")

	for _, p := range payments {
		if p.Token.Valid {
			privat := pbapi.PrivatBankAPI{
				UserAgent:   "Додаток API",
				Token:       p.Token.String,
				ContentType: "application/json;charset=utf8",
			}
			rsp, err := privat.SendPayment(p)
			if err == nil {
				logger.Log.WithFields(logrus.Fields{
					"payment":  p,
					"response": rsp,
				}).Trace("StartExchangePayments: Payment sent successfully!!!")
			} else {
				logger.Log.WithFields(logrus.Fields{
					"payment": p,
				}).Info("StartExchangePayments: Error SendPayment:", err.Error())
			}
			
			store.UpdatePayment(p, rsp, err)
		}
	}

	// send IBankCSV
	//fmt.Println("StartExchangePayments IBankCSV")
	// payments := make([]store.Payment, 0)
	// err := store.LoadPaymentsIBankCSV(&payments)
	// if err != nil {
	// 	return err
	// }
	// var iBankCSV api.BankAPI
	// iBankCSV = ibcsvapi.IBankCSVAPI{
	// 	Account:     a.Account,
	// 	BankRegistr: a.BankRegistr,
	// 	Direction:   a.Direction,
	// 	Records:     csvrecords,
	// }
	// tran, err := iBankCSV.GetTransactions()
	// if err == nil {
	// 	store.SaveTransactions(tran)
	// } else {
	// 	fmt.Println("StartExchangeTran: Error from GetTransactions iBank2UAcsv: ", err)
	// }
}
