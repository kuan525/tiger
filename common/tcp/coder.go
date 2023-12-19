package tcp

import (
	"bytes"
	"encoding/binary"
)

type DataPkg struct {
	Len  uint32
	Data []byte
}

func (d *DataPkg) Marshal() []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	// 大端字节顺序类似于我们眼睛适宜看的顺序，使用binary也是为了让报文跨平台，确定字节顺序
	// 这里将长度放在前面，这是一个固定的4字节长度，前期报文结构暂时简化处理
	binary.Write(bytesBuffer, binary.BigEndian, d.Len)
	return append(bytesBuffer.Bytes(), d.Data...)
}
