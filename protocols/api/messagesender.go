package api

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"net"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/enum"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/libraries/logging"
)

func sendNotificationMessage(conn net.Conn, msg enum.NotificationMsg, logger *logging.Logger) {
	msgBytes, err := json.Marshal(msg)
	var writeBuffer bytes.Buffer

	_ = binary.Write(&writeBuffer, binary.BigEndian, uint16(len(msgBytes)+4))
	_ = binary.Write(&writeBuffer, binary.BigEndian, enum.GossipNotification)
	_ = binary.Write(&writeBuffer, binary.BigEndian, msg.MessageID)
	_ = binary.Write(&writeBuffer, binary.BigEndian, msg.DataType)
	_ = binary.Write(&writeBuffer, binary.BigEndian, []byte(msg.Data))

	_, err = conn.Write(writeBuffer.Bytes())
	if err != nil {
		logger.InfoF("Failed to notification message: %v\n", err)
		return
	} else {
		logger.InfoF("Sent notification message to %s with data %s.\n", conn.RemoteAddr().String(), msg.Data)
	}
}
