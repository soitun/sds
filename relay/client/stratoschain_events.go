package client

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	setting "github.com/stratosnet/sds/cmd/relayd/config"
	"github.com/stratosnet/sds/msg/protos"
	"github.com/stratosnet/sds/relay/stratoschain"
	"github.com/stratosnet/sds/utils"
	"github.com/stratosnet/sds/utils/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"net/http"
)

func (m *MultiClient) SubscribeToStratosChainEvents() error {
	err := m.SubscribeToStratosChain("message.action='create_resource_node'", m.CreateResourceNodeMsgHandler())
	if err != nil {
		return err
	}
	err = m.SubscribeToStratosChain("message.action='update_resource_node_stake'", m.UpdateResourceNodeStakeMsgHandler())
	if err != nil {
		return err
	}
	err = m.SubscribeToStratosChain("message.action='unbonding_resource_node'", m.UnbondingResourceNodeMsgHandler())
	if err != nil {
		return err
	}
	err = m.SubscribeToStratosChain("message.action='remove_resource_node'", m.RemoveResourceNodeMsgHandler())
	if err != nil {
		return err
	}
	err = m.SubscribeToStratosChain("message.action='complete_unbonding_resource_node'", m.CompleteUnbondingResourceNodeMsgHandler())
	if err != nil {
		return err
	}
	err = m.SubscribeToStratosChain("message.action='create_indexing_node'", m.CreateIndexingNodeMsgHandler())
	if err != nil {
		return err
	}
	err = m.SubscribeToStratosChain("message.action='update_indexing_node_stake'", m.UpdateIndexingNodeStakeMsgHandler())
	if err != nil {
		return err
	}
	err = m.SubscribeToStratosChain("message.action='unbonding_indexing_node'", m.UnbondingIndexingNodeMsgHandler())
	if err != nil {
		return err
	}
	err = m.SubscribeToStratosChain("message.action='remove_indexing_node'", m.RemoveIndexingNodeMsgHandler())
	if err != nil {
		return err
	}
	err = m.SubscribeToStratosChain("message.action='complete_unbonding_indexing_node'", m.CompleteUnbondingIndexingNodeMsgHandler())
	if err != nil {
		return err
	}
	err = m.SubscribeToStratosChain("message.action='indexing_node_reg_vote'", m.IndexingNodeVoteMsgHandler())
	if err != nil {
		return err
	}
	err = m.SubscribeToStratosChain("message.action='SdsPrepayTx'", m.PrepayMsgHandler())
	if err != nil {
		return err
	}
	err = m.SubscribeToStratosChain("message.action='FileUploadTx'", m.FileUploadMsgHandler())
	if err != nil {
		return err
	}
	err = m.SubscribeToStratosChain("message.action='volume_report'", m.VolumeReportHandler())
	return err
}

func (m *MultiClient) CreateResourceNodeMsgHandler() func(event coretypes.ResultEvent) {
	return func(result coretypes.ResultEvent) {
		networkAddressList := result.Events["create_resource_node.network_address"]
		if len(networkAddressList) < 1 {
			utils.ErrorLog("No network address was specified in the create_resource_node message from stratos-chain")
			return
		}

		p2pAddress, err := types.BechToAddress(networkAddressList[0])
		if err != nil {
			utils.ErrorLog("Error when trying to convert P2P address to bytes", err)
			return
		}
		p2pAddressString, err := p2pAddress.ToBech(setting.Config.BlockchainInfo.P2PAddressPrefix)
		if err != nil {
			utils.ErrorLog("Error when trying to convert P2P address to bech32", err)
			return
		}

		nodePubkeyList := result.Events["create_resource_node.pub_key"]
		if len(nodePubkeyList) < 1 {
			utils.ErrorLog("No node pubkey was specified in the create_resource_node message from stratos-chain")
			return
		}
		p2pPubkeyRaw, err := hex.DecodeString(nodePubkeyList[0])
		if err != nil {
			utils.ErrorLog("Error when trying to decode P2P pubkey hex", err)
			return
		}
		p2pPubkey := ed25519.PubKeyEd25519{}
		err = stratoschain.Cdc.UnmarshalBinaryBare(p2pPubkeyRaw, &p2pPubkey)
		if err != nil {
			utils.ErrorLog("Error when trying to read P2P pubkey ed25519 binary", err)
			return
		}

		ozoneLimitChangeStr := result.Events["create_resource_node.ozone_limit_changes"]

		activatedMsg := &protos.ReqActivatedPP{
			P2PAddress:        p2pAddressString,
			P2PPubkey:         hex.EncodeToString(p2pPubkey[:]),
			OzoneLimitChanges: ozoneLimitChangeStr[0],
		}

		err = postToSP("/pp/activated", activatedMsg)
		if err != nil {
			utils.ErrorLog(err)
			return
		}
	}
}

func (m *MultiClient) UpdateResourceNodeStakeMsgHandler() func(event coretypes.ResultEvent) {
	return func(result coretypes.ResultEvent) {
		networkAddressList := result.Events["update_resource_node_stake.network_address"]
		if len(networkAddressList) < 1 {
			utils.ErrorLog("No network address was specified in the update_resource_node_stake message from stratos-chain")
			return
		}

		p2pAddress, err := types.BechToAddress(networkAddressList[0])
		if err != nil {
			utils.ErrorLog("Error when trying to convert P2P address to bytes", err)
			return
		}
		p2pAddressString, err := p2pAddress.ToBech(setting.Config.BlockchainInfo.P2PAddressPrefix)
		if err != nil {
			utils.ErrorLog("Error when trying to convert P2P address to bech32", err)
			return
		}

		ozoneLimitChangeStr := result.Events["update_resource_node_stake.ozone_limit_changes"]

		incrStakeBoolList := result.Events["update_resource_node_stake.incr_stake"]
		if len(incrStakeBoolList) < 1 {
			utils.ErrorLog("No incr stake status was specified in the update_resource_node_stake message from stratos-chain")
			return
		}

		updatedStakeMsg := &protos.ReqUpdatedStakePP{
			P2PAddress:        p2pAddressString,
			OzoneLimitChanges: ozoneLimitChangeStr[0],
			IncrStake:         incrStakeBoolList[0],
		}

		err = postToSP("/pp/updatedStake", updatedStakeMsg)
		if err != nil {
			utils.ErrorLog(err)
			return
		}
	}
}

func (m *MultiClient) UnbondingResourceNodeMsgHandler() func(event coretypes.ResultEvent) {
	return func(result coretypes.ResultEvent) {
		nodeAddressList := result.Events["unbonding_resource_node.resource_node"]
		if len(nodeAddressList) < 1 {
			fmt.Println("No node address was specified in the remove_resource_node message from stratos-chain")
			return
		}

		p2pAddress, err := types.BechToAddress(nodeAddressList[0])
		if err != nil {
			fmt.Println("Error when trying to convert P2P address to bytes: " + err.Error())
			return
		}
		p2pAddressString, err := p2pAddress.ToBech(setting.Config.BlockchainInfo.P2PAddressPrefix)
		if err != nil {
			fmt.Println("Error when trying to convert P2P address to bech32: " + err.Error())
			return
		}

		// get ozone limit change
		ozoneLimitChange := result.Events["unbonding_resource_node.ozone_limit_changes"]
		ozoneLimitChangeStr := ozoneLimitChange[0]
		// get mature time
		ubdMatureTime := result.Events["unbonding_resource_node.unbonding_mature_time"]
		ubdMatureTimeStr := ubdMatureTime[0]
		ubdMsg := &protos.ReqUnbondingPP{
			P2PAddress:          p2pAddressString,
			OzoneLimitChanges:   ozoneLimitChangeStr,
			UnbondingMatureTime: ubdMatureTimeStr,
		}

		err = postToSP("/pp/unbonding", ubdMsg)
		if err != nil {
			utils.ErrorLog(err)
			return
		}
	}
}

func (m *MultiClient) RemoveResourceNodeMsgHandler() func(event coretypes.ResultEvent) {
	return func(result coretypes.ResultEvent) {
		nodeAddressList := result.Events["remove_resource_node.resource_node"]
		if len(nodeAddressList) < 1 {
			utils.ErrorLog("No node address was specified in the remove_resource_node message from stratos-chain")
			return
		}

		p2pAddress, err := types.BechToAddress(nodeAddressList[0])
		if err != nil {
			utils.ErrorLog("Error when trying to convert P2P address to bytes", err)
			return
		}
		p2pAddressString, err := p2pAddress.ToBech(setting.Config.BlockchainInfo.P2PAddressPrefix)
		if err != nil {
			utils.ErrorLog("Error when trying to convert P2P address to bech32", err)
			return
		}

		deactivatedMsg := &protos.ReqDeactivatedPP{P2PAddress: p2pAddressString}

		err = postToSP("/pp/deactivated", deactivatedMsg)
		if err != nil {
			utils.ErrorLog(err)
			return
		}
	}
}

func (m *MultiClient) CompleteUnbondingResourceNodeMsgHandler() func(event coretypes.ResultEvent) {
	return m.RemoveResourceNodeMsgHandler()
}

func (m *MultiClient) CreateIndexingNodeMsgHandler() func(event coretypes.ResultEvent) {
	return func(result coretypes.ResultEvent) {
		// TODO
		utils.Log(fmt.Sprintf("%+v", result))
	}
}

func (m *MultiClient) UpdateIndexingNodeStakeMsgHandler() func(event coretypes.ResultEvent) {
	return func(result coretypes.ResultEvent) {
		networkAddressList := result.Events["update_indexing_node_stake.network_address"]
		if len(networkAddressList) < 1 {
			utils.ErrorLog("No network address was specified in the update_indexing_node_stake message from stratos-chain")
			return
		}

		p2pAddress, err := types.BechToAddress(networkAddressList[0])
		if err != nil {
			utils.ErrorLog("Error when trying to convert P2P address to bytes", err)
			return
		}
		p2pAddressString, err := p2pAddress.ToBech(setting.Config.BlockchainInfo.P2PAddressPrefix)
		if err != nil {
			utils.ErrorLog("Error when trying to convert P2P address to bech32", err)
			return
		}

		ozoneLimitChangeStr := result.Events["update_indexing_node_stake.ozone_limit_changes"]

		incrStakeBoolList := result.Events["update_indexing_node_stake.incr_stake"]
		if len(incrStakeBoolList) < 1 {
			utils.ErrorLog("No incr stake status was specified in the update_indexing_node_stake message from stratos-chain")
			return
		}

		updatedStakeMsg := &protos.ReqUpdatedStakeSP{
			P2PAddress:        p2pAddressString,
			OzoneLimitChanges: ozoneLimitChangeStr[0],
			IncrStake:         incrStakeBoolList[0],
		}
		err = postToSP("/chain/updatedStake", updatedStakeMsg)
		if err != nil {
			utils.ErrorLog(err)
			return
		}
	}
}

func (m *MultiClient) UnbondingIndexingNodeMsgHandler() func(event coretypes.ResultEvent) {
	return func(result coretypes.ResultEvent) {
		// TODO
		fmt.Printf("%+v\n", result)
	}
}
func (m *MultiClient) RemoveIndexingNodeMsgHandler() func(event coretypes.ResultEvent) {
	return func(result coretypes.ResultEvent) {
		// TODO
		utils.Log(fmt.Sprintf("%+v", result))
	}
}
func (m *MultiClient) CompleteUnbondingIndexingNodeMsgHandler() func(event coretypes.ResultEvent) {
	return m.RemoveIndexingNodeMsgHandler()
}

func (m *MultiClient) IndexingNodeVoteMsgHandler() func(event coretypes.ResultEvent) {
	return func(result coretypes.ResultEvent) {
		candidateNetworkAddressList := result.Events["indexing_node_reg_vote.candidate_network_address"]
		if len(candidateNetworkAddressList) < 1 {
			utils.ErrorLog("No candidate network address was specified in the indexing_node_reg_vote message from stratos-chain")
			return
		}
		p2pAddress, err := types.BechToAddress(candidateNetworkAddressList[0])
		if err != nil {
			utils.ErrorLog("Error when trying to convert P2P address to bytes", err)
			return
		}
		p2pAddressString, err := p2pAddress.ToBech(setting.Config.BlockchainInfo.P2PAddressPrefix)
		if err != nil {
			utils.ErrorLog("Error when trying to convert P2P address to bech32", err)
			return
		}

		candidateStatusList := result.Events["indexing_node_reg_vote.candidate_status"]
		if len(candidateStatusList) < 1 {
			utils.ErrorLog("No candidate status was specified in the indexing_node_reg_vote message from stratos-chain")
			return
		}
		if candidateStatusList[0] != sdkTypes.BondStatusBonded {
			utils.DebugLog("Indexing node vote handler: The candidate needs more votes before being considered active")
			return
		}

		activatedMsg := &protos.ReqActivatedSP{
			P2PAddress: p2pAddressString,
		}

		err = postToSP("/chain/activated", activatedMsg)
		if err != nil {
			utils.ErrorLog(err)
			return
		}
	}
}

func (m *MultiClient) PrepayMsgHandler() func(event coretypes.ResultEvent) {
	return func(result coretypes.ResultEvent) {
		utils.Log(fmt.Sprintf("%+v", result))

		reporterList := result.Events["Prepay.reporter"]
		if len(reporterList) < 1 {
			utils.ErrorLog("No reporter address was specified in the prepay message from stratos-chain")
			return
		}

		purchasedUozList := result.Events["Prepay.purchased"]
		if len(purchasedUozList) < 1 {
			utils.ErrorLog("No purchased ozone amount was specified in the prepay message from stratos-chain")
			return
		}

		prepaidMsg := &protos.ReqPrepaid{
			WalletAddress: reporterList[0],
			PurchasedUoz:  purchasedUozList[0],
		}

		err := postToSP("/pp/prepaid", prepaidMsg)
		if err != nil {
			utils.ErrorLog(err)
			return
		}
	}
}

func (m *MultiClient) FileUploadMsgHandler() func(event coretypes.ResultEvent) {
	return func(result coretypes.ResultEvent) {
		reporterAddressList := result.Events["FileUploadTx.reporter"]
		if len(reporterAddressList) < 1 {
			utils.ErrorLog("No reporter address was specified in the FileUploadTx message from stratos-chain")
			return
		}

		uploaderAddressList := result.Events["FileUploadTx.uploader"]
		if len(uploaderAddressList) < 1 {
			utils.ErrorLog("No uploader address was specified in the FileUploadTx message from stratos-chain")
			return
		}

		fileHashList := result.Events["FileUploadTx.file_hash"]
		if len(fileHashList) < 1 {
			utils.ErrorLog("No file hash was specified in the FileUploadTx message from stratos-chain")
			return
		}

		uploadedMsg := &protos.Uploaded{
			ReporterAddress: reporterAddressList[0],
			UploaderAddress: uploaderAddressList[0],
			FileHash:        fileHashList[0],
		}

		err := postToSP("/pp/uploaded", uploadedMsg)
		if err != nil {
			utils.ErrorLog(err)
			return
		}
	}
}

func (m *MultiClient) VolumeReportHandler() func(event coretypes.ResultEvent) {
	return func(result coretypes.ResultEvent) {
		// TODO
		utils.Log(fmt.Sprintf("%+v", result))
	}
}

func postToSP(endpoint string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return errors.New("Error when trying to marshal data to json: " + err.Error())
	}

	sdsApiUrl := "http://" + setting.Config.SDS.NetworkAddress + ":" + setting.Config.SDS.ApiPort
	resp, err := http.Post(sdsApiUrl+endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return errors.New("Error when calling " + endpoint + " endpoint in SP node: " + err.Error())
	}

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	utils.Log(endpoint+" endpoint response from SP node", resp.StatusCode, res["Msg"])
	return nil
}
