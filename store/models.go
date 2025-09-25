package store

import (
	"database/sql"
	"time"
)

// this data is loaded from db as list of bank's account and their parameters
type Account struct {
	Direction    int
	Account      string
	Bank         string
	Token        sql.NullString
	BankRegistr  string
	BalanceState []DataCheckBalance
}

type DataCheckBalance struct {
	Dpd         time.Time // date
	BankRegistr string    // registr of bank account in the capital2010
	FBAD        bool      // state  of bank account's saldo in the capital2010
}

// some banks can return state of exchange in own format and this universal state we can save to db
type DataState struct {
	Status             string
	Type               string
	Phase              string
	Today              string
	LastDay            string
	WorkBalance        string
	ServerDateTime     string
	DateFinalStatement string
}

// info about one transaction which ready to save to db
type DataTransaction struct {
	Direction   int    // `json:"Capital_Direction"`
	BankRegistr string // registr of bank account in the capital2010
	CntrCode    string // OKPO, INN...
	CntrName    string // Nmae
	CntrAcc     string //  Account
	DateTran    string // date in our account
	Comment     string // comment
	SumTran     int64  // SUMMA in national currency
	ID          string // unique number
	TranType    string // D, C
	NumDoc      string // numdoc for define our payments
}

type DataBalance struct {
	Direction      int    // `json:"Capital_Direction"`
	BankRegistr    string // registr of bank account in the capital2010
	Acc            string // account number
	Currency       string // UAH
	BalanceIn      int64  // balance on begin of day
	BalanceInEq    int64  // in national currence
	BalanceOut     int64
	BalanceOutEq   int64
	TurnoverDebt   int64
	TurnoverDebtEq int64
	TurnoverCred   int64
	TurnoverCredEq int64
	// BgfIBrnm       string
	// Brnm           string
	Dpd     string // date
	NameACC string // name of account
	// State          string // 1
	// Atp            string // D
	// Flmn           string // DN
	// DateOpenAccReg string
	// DateOpenAccSys string
	// DateCloseAcc   string
	IsFinalBal bool   // !!!
	Source     string // C- Capital; B- Bank
}

type Payment struct {
	DocumentNumber                 string         `json:"document_number"`
	PayerAccount                   string         `json:"payer_account"`
	RecipientAccount               string         `json:"recipient_account"`
	RecipientNceo                  string         `json:"recipient_nceo"`
	PaymentNaming                  string         `json:"payment_naming"`
	PaymentAmount                  string         `json:"payment_amount"`
	PaymentDestination             string         `json:"payment_destination"`
	DocumentType                   string         `json:"document_type,omitempty"`
	PaymentDate                    string         `json:"payment_date,omitempty"`
	PaymentAcceptDate              string         `json:"payment_accept_date,omitempty"`
	PaymentCbRef                   string         `json:"payment_cb_ref,omitempty"`
	CopyFromRef                    string         `json:"copy_from_ref,omitempty"`
	Attach                         string         `json:"attach,omitempty"`
	SignerMsg                      string         `json:"signer_msg,omitempty"`
	OdbMsg                         string         `json:"odb_msg,omitempty"`
	RecipientIfi                   string         `json:"recipient_ifi,omitempty"`
	RecipientIfiText               string         `json:"recipient_ifi_text,omitempty"`
	PayerUltmtNceo                 string         `json:"payer_ultmt_nceo,omitempty"`
	PayerUltmtDocumentSeries       string         `json:"payer_ultmt_document_series,omitempty"`
	PayerUltmtDocumentNumber       string         `json:"payer_ultmt_document_number,omitempty"`
	PayerUltmtDocumentIdNumber     string         `json:"payer_ultmt_document_id_number,omitempty"`
	PayerUltmtName                 string         `json:"payer_ultmt_name,omitempty"`
	RecipientUltmtNceo             string         `json:"recipient_ultmt_nceo,omitempty"`
	RecipientUltmtDocumentSeries   string         `json:"recipient_ultmt_document_series,omitempty"`
	RecipientUltmtDocumentNumber   string         `json:"recipient_ultmt_document_number,omitempty"`
	RecipientUltmtDocumentIdNumber string         `json:"recipient_ultmt_document_id_number,omitempty"`
	RecipientUltmtName             string         `json:"recipient_ultmt_name,omitempty"`
	StructCode                     string         `json:"struct_code,omitempty"`
	StructCategory                 string         `json:"struct_category,omitempty"`
	StructType                     string         `json:"struct_type,omitempty"`
	Token                          sql.NullString `json:"-"`
	ID_Key                         int64          `json:"-"`
}

type PaymentResponse struct {
	ResponseStatus      string `json:"status"`           // ERROR
	ResponseCode        int64  `json:"code"`             // 201 or 400
	PaymentRef          string `json:"payment_ref"`      // "референс створеного платежу"
	PaymentPackRef      string `json:"payment_pack_ref"` // "запакований референс створеного платежу"
	ResponseMessage     string `json:"message"`          //"invalid document number",
	ResponseRequestId   string `json:"requestId"`        // "20240223_131617_286f",
	ResponseServiceCode string `json:"serviceCode"`      // "PMTSRV0112"
}
