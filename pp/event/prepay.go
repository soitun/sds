package event

import (
	"context"

	"github.com/stratosnet/sds/framework/core"
	"github.com/stratosnet/sds/msg/header"
	"github.com/stratosnet/sds/msg/protos"
	"github.com/stratosnet/sds/pp"
	"github.com/stratosnet/sds/pp/peers"
	"github.com/stratosnet/sds/pp/requests"
	"github.com/stratosnet/sds/relay/stratoschain"
)

// Prepay PP node sends a prepay transaction
func Prepay(ctx context.Context, amount, fee, gas int64) error {
	prepayReq, err := reqPrepayData(amount, fee, gas)
	if err != nil {
		pp.ErrorLog(ctx, "Couldn't build PP prepay request: "+err.Error())
		return err
	}
	pp.Log(ctx, "Sending prepay message to SP! "+prepayReq.WalletAddress)
	peers.SendMessageToSPServer(ctx, prepayReq, header.ReqPrepay)
	return nil
}

// RspPrepay. Response to asking the SP node to send a prepay transaction
func RspPrepay(ctx context.Context, conn core.WriteCloser) {
	var target protos.RspPrepay
	success := requests.UnmarshalData(ctx, &target)
	if !success {
		return
	}

	pp.Log(ctx, "get RspPrepay", target.Result.State, target.Result.Msg)
	if target.Result.State != protos.ResultState_RES_SUCCESS {
		return
	}

	err := stratoschain.BroadcastTxBytes(target.Tx)
	if err != nil {
		pp.ErrorLog(ctx, "The prepay transaction couldn't be broadcast", err)
	} else {
		pp.Log(ctx, "The prepay transaction was broadcast")
	}
}

// RspPrepaid. Response when this PP node's prepay transaction was successful
func RspPrepaid(ctx context.Context, conn core.WriteCloser) {
	pp.Log(ctx, "The prepay transaction has been executed")
}
