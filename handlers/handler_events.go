package handlers

import (
	"github.com/irisnet/ibc-explorer-sync/models"
	"github.com/irisnet/ibc-explorer-sync/utils/constant"
)

type HandleEvents func(event models.Event) (models.Event, bool)

// Template function for handling events
func hookEvents(events []models.Event, handlefunc HandleEvents) []models.Event {
	var change bool
	for i, e := range events {
		if handlefunc == nil {
			return events
		}
		e, change = handlefunc(e)
		if change {
			one := models.Event{
				Type:       e.Type,
				Attributes: e.Attributes,
			}
			events[i] = one
		}
	}
	return events
}

// Remove header from update_client events
func removeHeaderOfUpdateClientEvents(e models.Event) (models.Event, bool) {
	var change bool
	if e.Type != constant.IbcUpdateClientEventTypeUpdateClient {
		return e, change
	}
	if len(e.Attributes) > 0 {
		for j, v := range e.Attributes {
			if v.Key == constant.IbcUpdateClientEventAttriKeyHeader {
				change = true
				attr := models.KvPair{
					Key:   v.Key,
					Value: "ignore ibc header info",
				}
				e.Attributes[j] = attr
			}
		}
	}
	return e, true
}

// Remove packet_data_hex from IBC events
func removePacketDataHexOfIbcTxEvents(e models.Event) (models.Event, bool) {
	var change bool
	if e.Type != constant.IbcRecvPacketEventTypeRecvPacket &&
		e.Type != constant.IbcRecvPacketEventTypeWriteAcknowledge &&
		e.Type != constant.IbcTransferEventTypeSendPacket {
		return e, change
	}
	if len(e.Attributes) > 0 {
		for j, v := range e.Attributes {
			if v.Key == constant.IbcRecvPacketEventAttriKeyPacketDataHex {
				change = true
				attr := models.KvPair{
					Key:   v.Key,
					Value: "ignore ibc packet_data_hex info",
				}
				e.Attributes[j] = attr
			}
		}
	}
	return e, true
}
