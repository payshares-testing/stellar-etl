package utils

import (
	"encoding/hex"
	"errors"
	"math/big"
	"time"

	"github.com/spf13/pflag"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/support/log"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/go/xdr"
)

//PanicOnError is a function that panics if the provided error is not nil
func PanicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

//HashToHexString is utility function that converts and xdr.Hash type to a hex string
func HashToHexString(inputHash xdr.Hash) string {
	sliceHash := inputHash[:]
	hexString := hex.EncodeToString(sliceHash)
	return hexString
}

//TimePointToUTCTimeStamp takes in an xdr TimePoint and converts it to a time.Time struct in UTC. It returns an error for negative timepoints
func TimePointToUTCTimeStamp(providedTime xdr.TimePoint) (time.Time, error) {
	intTime := int64(providedTime)
	if intTime < 0 {
		return time.Now(), errors.New("The timepoint is negative")
	}
	return time.Unix(intTime, 0).UTC(), nil
}

//GetAccountAddressFromMuxedAccount takes in a muxed account and returns the address of the account
func GetAccountAddressFromMuxedAccount(account xdr.MuxedAccount) (string, error) {
	providedID := account.ToAccountId()
	pointerToID := &providedID
	return pointerToID.GetAddress()
}

//CreateSampleTx creates a transaction with a single operation (BumpSequence), the min base fee, and infinite timebounds
func CreateSampleTx(sequence int64) xdr.TransactionEnvelope {
	kp, err := keypair.Random()
	PanicOnError(err)

	sourceAccount := txnbuild.NewSimpleAccount(kp.Address(), int64(0))
	tx, err := txnbuild.NewTransaction(
		txnbuild.TransactionParams{
			SourceAccount: &sourceAccount,
			Operations: []txnbuild.Operation{
				&txnbuild.BumpSequence{
					BumpTo: int64(sequence),
				},
			},
			BaseFee:    txnbuild.MinBaseFee,
			Timebounds: txnbuild.NewInfiniteTimeout(),
		},
	)
	PanicOnError(err)

	env, err := tx.TxEnvelope()
	PanicOnError(err)
	return env
}

//ConvertStroopValueToReal converts a value in stroops, the smallest amount unit, into real units
func ConvertStroopValueToReal(input xdr.Int64) float64 {
	output, _ := big.NewRat(int64(input), int64(10000000)).Float64()
	return output
}

//CreateSampleResultMeta creates Transaction results with the desired success flag and number of sub operation results
func CreateSampleResultMeta(successful bool, subOperationCount int) xdr.TransactionResultMeta {
	resultCode := xdr.TransactionResultCodeTxFailed
	if successful {
		resultCode = xdr.TransactionResultCodeTxSuccess
	}
	operationResults := []xdr.OperationResult{}
	for i := 0; i < subOperationCount; i++ {
		operationResults = append(operationResults, xdr.OperationResult{
			Code: xdr.OperationResultCodeOpInner,
			Tr:   &xdr.OperationResultTr{},
		})
	}
	return xdr.TransactionResultMeta{
		Result: xdr.TransactionResultPair{
			Result: xdr.TransactionResult{
				Result: xdr.TransactionResultResult{
					Code:    resultCode,
					Results: &operationResults,
				},
			},
		},
	}
}

// AddBasicFlags adds the start-ledger, end-ledger, limit, output, and stdout flags to the provided flagset
func AddBasicFlags(objectName string, flags *pflag.FlagSet) {
	flags.Uint32P("start-ledger", "s", 0, "The ledger sequence number for the beginning of the export period")
	flags.Uint32P("end-ledger", "e", 0, "The ledger sequence number for the end of the export range (required)")
	flags.Int64P("limit", "l", -1, "Maximum number of "+objectName+" to export. If the limit is set to a negative number, all the objects in the provided range are exported")
	flags.StringP("output", "o", "exported_"+objectName+".txt", "Filename of the output file")
	flags.Bool("stdout", false, "If set, the output will be printed to stdout instead of to a file")
}

// MustBasicFlags gets the values of the start-ledger, end-ledger, limit, output, and stdout flags from the flag set. If any do not exist, it stops the program fatally using the logger
func MustBasicFlags(flags *pflag.FlagSet, logger *log.Entry) (startNum, endNum uint32, limit int64, path string, useStdOut bool) {
	startNum, err := flags.GetUint32("start-ledger")
	if err != nil {
		logger.Fatal("could not get start sequence number: ", err)
	}

	endNum, err = flags.GetUint32("end-ledger")
	if err != nil {
		logger.Fatal("could not get end sequence number: ", err)
	}

	limit, err = flags.GetInt64("limit")
	if err != nil {
		logger.Fatal("could not get limit: ", err)
	}

	path, err = flags.GetString("output")
	if err != nil {
		logger.Fatal("could not get output filename: ", err)
	}

	useStdOut, err = flags.GetBool("stdout")
	if err != nil {
		logger.Fatal("could not get stdout boolean: ", err)
	}

	return
}
