package store

import (
	"capitalbank/db"
	"capitalbank/logger"
	"database/sql"
	"encoding/json"
	"os"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/sirupsen/logrus"
)

// ...

func UpdatePayment(p Payment, rsp PaymentResponse, rsperr error) {

	// Insert data into bank_payments_history and update payment's status
	// Encode payment struct into JSON
	rspJSON, err := json.Marshal(rsp)
	if err != nil {
		logger.Log.Error("UpdatePayment: Error Marshall response: ", err.Error())
		os.Exit(1)
	}

	// logger.Log.WithFields(logrus.Fields{
	// 	"id_payment": p.ID_Key,
	// 	"rsp":        string(rspJSON),
	// 	"err":        utils.NS(rsperr.Error()),
	// 	"ref_num":    rsp.PaymentPackRef,
	// }).Trace("Run sp.[bank_PaymentsResponseSave]")
	var rsperrstr string
	if rsperr != nil {
		rsperrstr = rsperr.Error()
	} else {
		rsperrstr = "OK"
	}

	if _, err := db.DB.Exec("exec bank_PaymentsResponseSave @id_payment, @rsp, @err, @ref_num",
		sql.Named("id_payment", p.ID_Key),
		sql.Named("rsp", string(rspJSON)),
		sql.Named("err", rsperrstr),
		sql.Named("ref_num", rsp.PaymentPackRef)); err != nil {
		logger.Log.WithFields(logrus.Fields{
			"id_key": p.ID_Key,
		}).Info("UpdatePayment: Error run sp.[bank_PaymentsResponseSave]: ", err.Error())
	}

}
