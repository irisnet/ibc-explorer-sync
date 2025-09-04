package handlers

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cometbft/cometbft/types"
	"github.com/irisnet/ibc-explorer-sync/config"
	"github.com/irisnet/ibc-explorer-sync/libs/logger"
	"github.com/irisnet/ibc-explorer-sync/libs/msgparser"
	"github.com/irisnet/ibc-explorer-sync/libs/msgparser/codec"
	. "github.com/irisnet/ibc-explorer-sync/libs/msgparser/modules"
	"github.com/irisnet/ibc-explorer-sync/libs/msgparser/modules/ibc"
	msgsdktypes "github.com/irisnet/ibc-explorer-sync/libs/msgparser/types"
	"github.com/irisnet/ibc-explorer-sync/libs/pool"
	"github.com/irisnet/ibc-explorer-sync/models"
	"github.com/irisnet/ibc-explorer-sync/utils"
	"github.com/irisnet/ibc-explorer-sync/utils/constant"
	"golang.org/x/net/context"
)

var (
	_parser    msgparser.MsgParser
	_conf      *config.Config
	_filterMap map[string]string
)

func InitRouter(conf *config.Config) {
	_conf = conf
	router := msgparser.RegisteRouter()
	_parser = msgparser.NewMsgParser(router)

	if conf.Server.Bech32AccPrefix != "" {
		initBech32Prefix(conf.Server.Bech32AccPrefix)
	}
	//ibc-zone
	if filterMsgType := models.GetSrvConf().SupportTypes; filterMsgType != "" {
		msgTypes := strings.Split(filterMsgType, ",")
		_filterMap = make(map[string]string, len(msgTypes))
		for _, val := range msgTypes {
			_filterMap[val] = val
		}
	}
}

func ParseBlockAndTxs(b int64, client *pool.Client) (*models.Block, []*models.Tx, error) {
	var (
		blockDoc models.Block
		block    *ctypes.ResultBlock
	)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if v, err := client.Block(ctx, &b); err != nil {
		time.Sleep(1 * time.Second)
		if v2, err := client.Block(ctx, &b); err != nil {
			return &blockDoc, nil, utils.ConvertErr(b, "", "ParseBlock", err)
		} else {
			block = v2
		}
	} else {
		block = v
	}
	blockDoc = models.Block{
		Height:   block.Block.Height,
		Time:     block.Block.Time.Unix(),
		Hash:     block.Block.Header.Hash().String(),
		Txn:      int64(len(block.Block.Data.Txs)),
		Proposer: block.Block.ProposerAddress.String(),
	}

	txResultMap := handleTxResult(client, block.Block)

	txDocs := make([]*models.Tx, 0, len(block.Block.Txs))
	if len(block.Block.Txs) > 0 {
		for i, v := range block.Block.Txs {
			if !includeIbcTxs(v, block.Block.Height) {
				continue
			}
			txHash := utils.BuildHex(v.Hash())
			txResult, ok := txResultMap[txHash]
			if !ok {
				return &blockDoc, txDocs, utils.ConvertErr(block.Block.Height, txHash, "TxResult",
					fmt.Errorf("no found"))
			}
			if txResult.Err != nil {
				return &blockDoc, txDocs, utils.ConvertErr(block.Block.Height, txHash, "TxResult",
					txResult.Err)
			}
			txDoc, err := parseTx(v, txResult.TxResult, block.Block, i)
			if err != nil {
				return &blockDoc, txDocs, err
			}
			if txDoc.TxHash != "" && len(txDoc.Type) > 0 {
				txDocs = append(txDocs, &txDoc)
			}
		}
	}

	return &blockDoc, txDocs, nil
}

func parseTx(txBytes types.Tx, txResult *ctypes.ResultTx, block *types.Block, index int) (models.Tx, error) {
	var (
		docTx          models.Tx
		docTxMsgs      []msgsdktypes.TxMsg
		includeCfgType bool
	)
	txHash := utils.BuildHex(txBytes.Hash())

	docTx.Time = block.Time.Unix()
	docTx.Height = block.Height
	docTx.TxHash = txHash
	docTx.Status = parseTxStatus(txResult.TxResult.Code)
	if docTx.Status == constant.TxStatusFail {
		docTx.Log = txResult.TxResult.Log
	}

	docTx.TxIndex = uint32(index)
	docTx.TxId = block.Height*100000 + int64(index)
	eventsIndexMap, docTxEventsNew := parseABCILogsFromTxResult(txResult)

	authTx, err := codec.GetSigningTx(txBytes)
	if err != nil {
		logger.Warn(err.Error(),
			logger.String("errTag", "TxDecoder"),
			logger.String("txhash", txHash),
			logger.Int64("height", block.Height))
		return docTx, nil
	}
	docTx.GasUsed = txResult.TxResult.GasUsed
	docTx.Fee = msgsdktypes.BuildFee(authTx.GetFee(), authTx.GetGas())
	docTx.Memo = authTx.GetMemo()

	msgs := authTx.GetMsgs()
	if len(msgs) == 0 {
		return docTx, nil
	}

	for msgIndex, v := range msgs {
		msgDocInfo := _parser.HandleTxMsg(v)
		if len(msgDocInfo.Addrs) == 0 {
			if len(_filterMap) > 0 {
				for id, one := range docTxEventsNew {
					if one.MsgIndex == uint32(msgIndex) {
						docTxEventsNew[id].Events = []models.Event{}
						break
					}
				}
				//add empty msg for msgIndex match
				docTxMsgs = append(docTxMsgs, msgsdktypes.TxMsg{Type: "no setting type"})
			}
			continue
		}
		if len(_filterMap) > 0 {
			_, ok := _filterMap[msgDocInfo.DocTxMsg.Type]
			if ok && !includeCfgType {
				includeCfgType = true
			}
			if !ok {
				for id, one := range docTxEventsNew {
					if one.MsgIndex == uint32(msgIndex) {
						docTxEventsNew[id].Events = []models.Event{}
						break
					}
				}
				docTxMsgs = append(docTxMsgs, msgsdktypes.TxMsg{Type: msgDocInfo.DocTxMsg.Type})
				continue
			}
		}
		var dsPort string
		switch msgDocInfo.DocTxMsg.Type {
		case MsgTypeIBCTransfer:
			if ibcTranferMsg, ok := msgDocInfo.DocTxMsg.Msg.(*ibc.DocMsgTransfer); ok {
				if val, exist := eventsIndexMap[uint32(msgIndex)]; exist {
					ibcTranferMsg.PacketId = buildPacketId(val.Events)
					msgDocInfo.DocTxMsg.Msg = ibcTranferMsg
				}
				if _conf.Server.IgnoreIbcHeader {
					for id, one := range docTxEventsNew {
						if one.MsgIndex == uint32(msgIndex) {
							docTxEventsNew[id].Events = hookEvents(docTxEventsNew[id].Events, removePacketDataHexOfIbcTxEvents)
							break
						}
					}
				}

			} else {
				logger.Warn("ibc transfer handler packet_id failed", logger.String("errTag", "TxMsg"),
					logger.String("txhash", txHash),
					logger.Int("msg_index", msgIndex),
					logger.Int64("height", block.Height))
			}
		case MsgTypeRecvPacket:
			for id, one := range docTxEventsNew {
				if one.MsgIndex == uint32(msgIndex) {
					docTxEventsNew[id].Events = updateEvents(docTxEventsNew[id].Events, UnmarshalAcknowledgement)
					break
				}
			}
			recvPacketMsg, ok := msgDocInfo.DocTxMsg.Msg.(*ibc.DocMsgRecvPacket)
			if ok {
				dsPort = recvPacketMsg.Packet.DestinationPort
			}
			if _conf.Server.IgnoreIbcHeader {
				if ok {
					recvPacketMsg.ProofCommitment = "ignore ibc ProofCommitment info"
					msgDocInfo.DocTxMsg.Msg = recvPacketMsg
				}
				for id, one := range docTxEventsNew {
					if one.MsgIndex == uint32(msgIndex) {
						docTxEventsNew[id].Events = hookEvents(docTxEventsNew[id].Events, removePacketDataHexOfIbcTxEvents)
						break
					}
				}
			}
		case MsgTypeUpdateClient:
			if _conf.Server.IgnoreIbcHeader {
				updateClientMsg, ok := msgDocInfo.DocTxMsg.Msg.(*ibc.DocMsgUpdateClient)
				if ok {
					updateClientMsg.Header = "ignore ibc header info"
					msgDocInfo.DocTxMsg.Msg = updateClientMsg
				}
				for id, one := range docTxEventsNew {
					if one.MsgIndex == uint32(msgIndex) {
						docTxEventsNew[id].Events = hookEvents(docTxEventsNew[id].Events, removeHeaderOfUpdateClientEvents)
						break
					}
				}
			}
		case MsgTypeTimeout:
			timeOutMsg, ok := msgDocInfo.DocTxMsg.Msg.(*ibc.DocMsgTimeout)
			if ok {
				dsPort = timeOutMsg.Packet.DestinationPort
			}
			if _conf.Server.IgnoreIbcHeader {
				if ok {
					timeOutMsg.ProofUnreceived = "ignore ibc ProofUnreceived info"
					msgDocInfo.DocTxMsg.Msg = timeOutMsg
				}
			}
		case MsgTypeAcknowledgement:
			ackMsg, ok := msgDocInfo.DocTxMsg.Msg.(*ibc.DocMsgAcknowledgement)
			if ok {
				dsPort = ackMsg.Packet.DestinationPort
			}
		}
		if msgIndex == 0 {
			docTx.Type = msgDocInfo.DocTxMsg.Type
		}
		if docTx.Type == "" {
			docTx.Type = msgDocInfo.DocTxMsg.Type
		}

		//If this is ICA (Inter-Chain Account) cross-chain data, do not save events and msg content
		if dsPort == IcaHostPortID {
			for id, one := range docTxEventsNew {
				if one.MsgIndex == uint32(msgIndex) {
					docTxEventsNew[id].Events = []models.Event{}
					break
				}
			}
			docTxMsgs = append(docTxMsgs, msgsdktypes.TxMsg{Type: msgDocInfo.DocTxMsg.Type})
			continue
		}

		docTx.Signers = append(docTx.Signers, removeDuplicatesFromSlice(msgDocInfo.Signers)...)
		docTx.Addrs = append(docTx.Addrs, removeDuplicatesFromSlice(msgDocInfo.Addrs)...)
		docTxMsgs = append(docTxMsgs, msgDocInfo.DocTxMsg)
		docTx.Types = append(docTx.Types, msgDocInfo.DocTxMsg.Type)
	}

	docTx.Addrs = removeDuplicatesFromSlice(docTx.Addrs)
	docTx.Types = removeDuplicatesFromSlice(docTx.Types)
	docTx.Signers = removeDuplicatesFromSlice(docTx.Signers)
	docTx.DocTxMsgs = docTxMsgs
	docTx.EventsNew = docTxEventsNew

	//setting type but not included in tx,skip this tx
	if len(_filterMap) > 0 && !includeCfgType {
		logger.Warn("skip tx for no include setting types",
			logger.String("types", strings.Join(docTx.Types, ",")),
			logger.String("txhash", txHash),
			logger.Int64("height", block.Height))
		return models.Tx{}, nil
	}

	// don't save txs which have not parsed
	if docTx.Type == "" {
		logger.Warn(constant.NoSupportMsgTypeTag,
			logger.String("errTag", "TxMsg"),
			logger.String("txhash", txHash),
			logger.Int64("height", block.Height))
		return models.Tx{}, nil
	}

	return docTx, nil
}

// Parse events for chains like Osmosis where successful transaction logs may be empty
func parseABCILogsFromTxResult(txResult *ctypes.ResultTx) (map[uint32]models.EventNew, []models.EventNew) {
	if txResult.TxResult.Code == 0 {
		if txResult.TxResult.Log != "" {
			return splitEvents(txResult.TxResult.Log)
		} else {
			return parseABCIEvents(txResult)
		}
	}
	return nil, nil
}

func buildPacketId(events []models.Event) string {
	if len(events) > 0 {
		var mapKeyValue map[string]string
		for _, e := range events {
			if len(e.Attributes) > 0 && e.Type == constant.IbcTransferEventTypeSendPacket {
				mapKeyValue = make(map[string]string, len(e.Attributes))
				for _, v := range e.Attributes {
					mapKeyValue[string(v.Key)] = string(v.Value)
				}
				break
			}
		}

		if len(mapKeyValue) > 0 {
			scPort := mapKeyValue[constant.IbcTransferEventAttriKeyPacketScPort]
			scChannel := mapKeyValue[constant.IbcTransferEventAttriKeyPacketScChannel]
			dcPort := mapKeyValue[constant.IbcTransferEventAttriKeyPacketDcPort]
			dcChannel := mapKeyValue[constant.IbcTransferEventAttriKeyPacketDcChannels]
			sequence := mapKeyValue[constant.IbcTransferEventAttriKeyPacketSequence]
			return fmt.Sprintf("%v%v%v%v%v", scPort, scChannel, dcPort, dcChannel, sequence)
		}
	}
	return ""
}

func parseTxStatus(code uint32) uint32 {
	if code == 0 {
		return constant.TxStatusSuccess
	} else {
		return constant.TxStatusFail
	}
}

func splitEvents(log string) (map[uint32]models.EventNew, []models.EventNew) {
	var eventDocs []models.EventNew
	if log != "" {
		eventDocs = parseABCILogs(log)
	}

	msgIndexMap := make(map[uint32]models.EventNew, len(eventDocs))
	for _, val := range eventDocs {
		msgIndexMap[val.MsgIndex] = val
	}
	return msgIndexMap, eventDocs
}

func updateEvents(events []models.Event, fn func([]byte) string) []models.Event {

	for i, e := range events {
		if e.Type != constant.IbcRecvPacketEventTypeWriteAcknowledge {
			continue
		}
		if len(e.Attributes) > 0 {
			for j, v := range e.Attributes {
				if v.Key == constant.IbcRecvPacketEventAttriKeyPacketAck {
					attr := models.KvPair{
						Key:   string(v.Key),
						Value: string(v.Value),
					}
					attr.Value = fn([]byte(v.Value))
					e.Attributes[j] = attr
				}
			}
		}
		one := models.Event{
			Type:       e.Type,
			Attributes: e.Attributes,
		}
		events[i] = one
	}
	return events
}

// parseABCILogs attempts to parse a stringified ABCI tx log into a slice of
// EventNe types. It ignore error upon JSON decoding failure.
func parseABCILogs(logs string) []models.EventNew {
	var res []models.EventNew
	utils.UnMarshalJsonIgnoreErr(logs, &res)
	return res
}

// 从txResult.events中解析只保留ibc相关的子event
var (
	ibcAbciEventTypesMap = map[string]struct{}{
		"ibc_transfer":                  {},
		"transfer":                      {},
		"fungible_token_packet":         {},
		"update_client":                 {},
		"send_packet":                   {},
		"recv_packet":                   {},
		"acknowledge_packet":            {},
		"timeout_packet":                {},
		"channel_open_confirm":          {},
		"write_acknowledgement":         {},
		"denomination_trace":            {},
		"uptick.erc20.v1.EventIBCERC20": {},
	}
)

func parseABCIEvents(txResult *ctypes.ResultTx) (map[uint32]models.EventNew, []models.EventNew) {
	var res []models.EventNew
	msgIndexMap := make(map[uint32]models.EventNew, 2)
	for i := range txResult.TxResult.Events {
		if _, ok := ibcAbciEventTypesMap[txResult.TxResult.Events[i].Type]; !ok {
			//Skip non-IBC related sub-events
			continue
		}
		var msgIndex string
		lengAttr := len(txResult.TxResult.Events[i].Attributes)
		for index := lengAttr - 1; index >= 0; index-- {
			if txResult.TxResult.Events[i].Attributes[index].Key == constant.IbcTxEventAttriKeyMsgIndex {
				msgIndex = txResult.TxResult.Events[i].Attributes[index].Value
				break
			}
		}
		if msgIndex != "" {
			var event models.Event
			utils.UnMarshalJsonIgnoreErr(utils.MarshalJsonIgnoreErr(txResult.TxResult.Events[i]), &event)
			if value, err := strconv.ParseUint(msgIndex, 10, 64); err == nil {
				if eventsNew, ok := msgIndexMap[uint32(value)]; ok {
					eventsNew.Events = append(eventsNew.Events, event)
					msgIndexMap[uint32(value)] = eventsNew
				} else {
					data := models.EventNew{
						MsgIndex: uint32(value),
					}
					data.Events = make([]models.Event, 0, 10)
					data.Events = append(data.Events, event)
					msgIndexMap[uint32(value)] = data
				}
			}
		}
	}

	for _, val := range msgIndexMap {
		res = append(res, val)
	}
	//Store in order of msg index
	sort.Slice(res, func(i, j int) bool {
		return res[i].MsgIndex < res[j].MsgIndex
	})
	return msgIndexMap, res
}

func removeDuplicatesFromSlice(data []string) (result []string) {
	tempSet := make(map[string]string, len(data))
	for _, val := range data {
		if _, ok := tempSet[val]; ok || val == "" {
			continue
		}
		tempSet[val] = val
	}
	for one := range tempSet {
		result = append(result, one)
	}
	return
}
