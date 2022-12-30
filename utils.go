package mqttproxy

func decodeMsgType(header byte) MsgType {
	mType := (header & 0xF0) >> 4
	return MsgType(mType)
}
