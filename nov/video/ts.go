package video

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ParserWriter ...
type ParserWriter interface {
	// Write  write out data
	Write([]byte) (int, error)
}

// Packet ...
type Packet struct {
	Payload string
}

type tsHeader struct {
	syncByte                   uint32
	transportErrorIndicator    uint32
	payloadUnitStartIndicator  uint32
	transportPriority          uint32
	pid                        uint32
	transportScramblingControl uint32
	adaptionFieldControl       uint32
	continuityCounter          uint32
}

type tsParser struct {
	filePath string
	fs       *os.File
	w        io.Writer

	pmtPids          []uint32
	sis              []streamInfo
	firstPacket      bool
	lastPacket       bool
	firstFramePacket bool
	framePos         uint32
	frameLength      uint32
	frameByte        []byte

	tps []Packet
}

// ParseTsFile ...
func ParseTsFile(path string, w io.Writer) error {
	tp := &tsParser{
		filePath: path,
		w:        w,
		tps:      make([]Packet, 1),
	}
	return tp.parseTSFile()
}

func (tp *tsParser) parseTSFile() error {
	var err error
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}
	dir = strings.Replace(dir, "\\", "/", -1)
	dir = dir + "/"
	tp.filePath = dir + tp.filePath
	if tp.fs, err = os.Open(tp.filePath); err != nil {
		return err
	}
	defer tp.fs.Close()

	reader := bufio.NewReader(tp.fs)

	tp.write("<html>\n")
	tmpString := fmt.Sprintf("<h1>%s</h1>\n", tp.filePath)
	tp.write(tmpString)
	// b := make([]byte, 188)
	for i := 0; i < 500; i++ {
		var n []byte
		if n, err = reader.Peek(188); err != nil {
			return err
		}
		b := []byte{}
		b = append(b, n...)
		reader.Discard(188)
		headerByte, err := reader.Peek(4)
		if err != nil {
			return err
		}
		nextHeader := tp.parseTSHeader(headerByte)
		if err = tp.parsePacket(b, nextHeader); err != nil {
			return err
		}
	}
	tp.write("</html>")
	return nil
}

func (tp *tsParser) parseTSHeader(headerByte []byte) *tsHeader {
	b := &BitReader{
		pos:    0,
		offs:   0,
		buffer: headerByte,
		EOF:    false,
	}
	h := &tsHeader{}
	h.syncByte = uint32(b.Read(8))
	h.transportErrorIndicator = uint32(b.Read(1))
	h.payloadUnitStartIndicator = uint32(b.Read(1))
	h.transportPriority = uint32(b.Read(1))
	h.pid = uint32(b.Read(13))
	h.transportScramblingControl = uint32(b.Read(2))
	h.adaptionFieldControl = uint32(b.Read(2))
	h.continuityCounter = uint32(b.Read(4))

	return h
}

func (tp *tsParser) parsePacket(pkt []byte, nextHeader *tsHeader) error {
	b := &BitReader{
		pos:    0,
		offs:   0,
		buffer: pkt,
		EOF:    false,
	}
	h := tp.parseTSHeader(b.buffer[0:4])
	b.SetPos(4)

	// for i := 0; i < 8; i++ {
	// 	fmt.Printf("%02x ", b.buffer[i])
	// }
	// fmt.Printf("\n")
	// fmt.Printf("adaption_field_control: %d\n", h.adaptionFieldControl)

	var err error
	parseContinue := true
	if h.pid == 0 {
		// PAT
		tp.write("<hr />\n")
		err = tp.parsePAT(b)
		parseContinue = false
	}

	if parseContinue {
		for _, v := range tp.pmtPids {
			if v == h.pid {
				tp.write("<hr />\n")
				err = tp.parsePMT(b)
				parseContinue = false
			}
			tp.firstFramePacket = true
			break
		}
	}

	if parseContinue {
		for _, v := range tp.sis {
			if v.elementaryPID == h.pid {
				parseContinue = false
				tp.firstPacket = false
				tp.lastPacket = false
				if h.payloadUnitStartIndicator == 1 {
					tp.write("<hr />\n")
					tp.firstPacket = true
					tp.framePos = 0
					tp.write(fmt.Sprintf("<h2>Frame Pes Header (PID:%d, %s)</h2>\n", h.pid, v.typeString))
				}
				if nextHeader.payloadUnitStartIndicator == 1 {
					tp.lastPacket = true
				}

				err = tp.parseFrame(b, v)

				if tp.lastPacket {
					if v.streamType == 0x1b {
						tp.parseH264Frame()
					}
					tp.firstFramePacket = false
					if h.pid != nextHeader.pid {
						tp.firstFramePacket = true
					}
				}
				break
			}
		}
	}

	if parseContinue {
		tp.write(fmt.Sprintf("<hr />\n<h2>PID: %d</h2>\n", h.pid))
	}

	return err
}

type sectionHeader struct {
	tableID                uint32
	sectionSyntaxIndicator uint32
	zero                   uint32
	reserved1              uint32
	sectionLength          uint32
	tranportStreamID       uint32
	reserved2              uint32
	versionNumber          uint32
	currentNextIndicator   uint32
	sectionNumber          uint32
	lastSectionNumber      uint32
}

type pmtNumber struct {
	programNumber uint32
	reserved      uint32
	pid           uint32
}

type patHeader struct {
	sectionHeader
	pmts  []pmtNumber
	crc32 uint32
}

func (tp *tsParser) write(s string) {
	tp.w.Write([]byte(s))
}

func (tp *tsParser) parsePAT(b *BitReader) error {
	var err error
	adaptionLength := b.Read(8)
	fmt.Printf("PAT adaption length: %d\n", adaptionLength)
	var tmpString string
	tp.write("<h2>PAT</h2>")

	b.SetPos(5 + int(adaptionLength))
	h := &patHeader{}
	h.tableID = uint32(b.Read(8))
	h.sectionSyntaxIndicator = uint32(b.Read(1))
	h.zero = uint32(b.Read(1))
	h.reserved1 = uint32(b.Read(2))
	h.sectionLength = uint32(b.Read(12))
	h.tranportStreamID = uint32(b.Read(16))
	h.reserved2 = uint32(b.Read(2))
	h.versionNumber = uint32(b.Read(5))
	h.currentNextIndicator = uint32(b.Read(1))
	h.sectionNumber = uint32(b.Read(8))
	h.lastSectionNumber = uint32(b.Read(8))
	// fmt.Printf("table id: %d\nstream id: %d\nsection number: %d\n",
	// 	h.tableID, h.tranportStreamID, h.sectionNumber)
	h.pmts = []pmtNumber{}
	pt := pmtNumber{}
	tp.write(`<h4>PMT List</h4>
	<table border="1">
	<tr>
	<td>No.</td>
	<td>Program Number</td>
	<td>PID</td>
	<td>type</td>
	</tr>
	`)
	tp.pmtPids = []uint32{}
	for i := 0; i < int(h.sectionLength-9)/4; i++ {
		pt.programNumber = uint32(b.Read(16))
		pt.reserved = uint32(b.Read(3))
		pt.pid = uint32(b.Read(13))
		h.pmts = append(h.pmts, pt)
		tmpString = fmt.Sprintf("<tr><td>%d</td><td>%d</td><td>%d</td>",
			i, pt.programNumber, pt.pid)
		tp.write(tmpString)
		if pt.programNumber == 1 {
			tp.pmtPids = append(tp.pmtPids, pt.pid)
			tp.write("<td>PMT</td>")
		} else if pt.programNumber == 0 {
			tp.write("<td>NIT</td>")
		}
		tp.write("</tr>")
	}
	tp.write("\n</table>")
	h.crc32 = uint32(b.Read(32))

	return err
}

type streamInfo struct {
	streamType    uint32
	typeString    string
	reserved1     uint32
	elementaryPID uint32
	reserved2     uint32
	ESInfoLength  uint32
}

type pmtHeader struct {
	sectionHeader
	reserved3         uint32
	pcrPID            uint32
	reserved4         uint32
	programInfoLength uint32
	sis               []streamInfo
	crc32             uint32
}

func (tp *tsParser) parsePMT(b *BitReader) error {
	var err error
	adaptionLength := b.Read(8)
	fmt.Printf("PMT adaption length: %d\n", adaptionLength)
	// var tmpString string
	tp.write("<h2>PMT</h2>")
	b.SetPos(5 + int(adaptionLength))

	h := &pmtHeader{}
	h.tableID = uint32(b.Read(8))
	h.sectionSyntaxIndicator = uint32(b.Read(1))
	h.zero = uint32(b.Read(1))
	h.reserved1 = uint32(b.Read(2))
	h.sectionLength = uint32(b.Read(12))
	h.tranportStreamID = uint32(b.Read(16))
	h.reserved2 = uint32(b.Read(2))
	h.versionNumber = uint32(b.Read(5))
	h.currentNextIndicator = uint32(b.Read(1))
	h.sectionNumber = uint32(b.Read(8))
	h.lastSectionNumber = uint32(b.Read(8))
	// fmt.Printf("table id: %d\nstream id: %d\nsection number: %d\n",
	// 	h.tableID, h.tranportStreamID, h.sectionNumber)
	h.reserved3 = uint32(b.Read(3))
	h.pcrPID = uint32(b.Read(13))
	h.reserved4 = uint32(b.Read(4))
	h.programInfoLength = uint32(b.Read(12))

	tp.write(`<h4>Stream List</h4>
	<table border="1">
	<tr>
	<td>No.</td>
	<td>Stream</td>
	<td>Type</td>
	<td>Elementary PID</td>
	</tr>
	`)

	tp.sis = []streamInfo{}
	si := streamInfo{}
	for i := 0; i < int(h.sectionLength-13)/5; i++ {
		si.streamType = uint32(b.Read(8))
		si.reserved1 = uint32(b.Read(3))
		si.elementaryPID = uint32(b.Read(13))
		si.reserved2 = uint32(b.Read(4))
		si.ESInfoLength = uint32(b.Read(12))
		switch si.streamType {
		case 0x1b:
			si.typeString = "H.264"
		case 0x0f:
			si.typeString = "AAC"
		case 0x03:
			si.typeString = "MP3"
		default:
			si.typeString = "unknown"
		}
		tp.write(fmt.Sprintf(`<tr>
		<td>%d</td>
		<td>0x%02x</td>
		<td>%s</td>
		<td>%d</td>
		</tr>
		`, i, si.streamType, si.typeString, si.elementaryPID))
		tp.sis = append(tp.sis, si)
		h.sis = append(h.sis, si)
	}
	tp.write("</table>")
	return err
}

type pesHeader struct {
	pesStartCode    uint32
	streamID        uint32
	pesPacketLength uint32
	flag1           uint32
	flag2           uint32
	pesDataLength   uint32
	pts             uint64
	dts             uint64
}

func (tp *tsParser) parseFrame(b *BitReader, si streamInfo) error {
	var err error
	if tp.firstPacket {
		// pes
		adaptionLength := b.Read(8)
		// fmt.Printf("first frame packet, adaption length: %d\n", adaptionLength)
		if adaptionLength > 0 {
			b.SetPos(5 + int(adaptionLength))
		} else if b.Read(8) == 0 && b.Read(8) == 1 {
			b.SetPos(4)
		}
		h := &pesHeader{}
		h.pesStartCode = uint32(b.Read(24))
		h.streamID = uint32(b.Read(8))
		h.pesPacketLength = uint32(b.Read(16))
		h.flag1 = uint32(b.Read(8))
		h.flag2 = uint32(b.Read(8))
		h.pesDataLength = uint32(b.Read(8))
		b.Skip(7)
		h.pts = b.Read(33)
		if h.pesDataLength == 10 {
			b.Skip(7)
			h.dts = b.Read(33)
		}
		tp.write(fmt.Sprintf(`<table border="1">
		<tr>
			<td>Start code</td>
			<td>Stream ID</td>
			<td>Pes Packet Size</td>
			<td>PTS</td>
			<td>DTS</td>
		</tr>
		<tr>
			<td>0x%03x</td>
			<td>0x%01x</td>
			<td>%d</td>
			<td>%d</td>
			<td>%d</td>
		</tr>
		</table>
		`, h.pesStartCode, h.streamID, h.pesPacketLength, h.pts, h.dts))
		if h.pesPacketLength != 0 {
			tp.frameLength = h.pesPacketLength - 3 - h.pesDataLength
		} else {
			tp.frameLength = 0
		}
		tp.framePos = tp.framePos + 177 - uint32(adaptionLength)
		tp.frameByte = []byte{}
		tp.frameByte = append(tp.frameByte, b.buffer[14+int(uint32(adaptionLength)+h.pesDataLength):]...)
	} else if tp.lastPacket {
		adaptionLength := b.Read(8)
		if adaptionLength > 182 {
			return errors.New("invalid packet")
		}
		tp.frameByte = append(tp.frameByte, b.buffer[5+adaptionLength:]...)
	} else {
		// tp.write("<h2>Frame data packet</h2>\n")
		tp.frameByte = append(tp.frameByte, b.buffer[4:]...)
		tp.framePos += 184
		// tp.write(fmt.Sprintf("<h2>frame pos: %d</h2>\n", tp.framePos))
	}

	return err
}

var h264TypeMap = map[uint32]string{
	0:  "未使用",
	1:  "非关键帧",
	2:  "片分区A",
	3:  "片分区B",
	4:  "片分区C",
	5:  "关键帧",
	6:  "补充增强信息单元(SEI)",
	7:  "SPS",
	8:  "PPS",
	9:  "分解符",
	10: "序列结束",
	11: "码流结束",
}

func getNaluType(t uint32) string {
	if v, ok := h264TypeMap[t]; ok {
		return v
	}
	return "其他"
}

func (tp *tsParser) parseH264Frame() error {
	var err error
	b := &BitReader{
		pos:    0,
		offs:   0,
		buffer: tp.frameByte,
		EOF:    false,
	}
	// for i := 0; i < 100; i++ {
	// 	fmt.Printf("%d ", uint32(b.Read(8)))
	// }
	tp.write(fmt.Sprintf("<h2>H.264 Frame (size:%d)</h2>\n", len(tp.frameByte)))
	if tp.firstFramePacket {
		firstNalu := uint32(b.Read(32))
		ct := uint32(b.Read(8))
		if firstNalu != 1 || ct != 9 {
			return errors.New("invalid frame")
		}
	}
	tp.write(`<table border="1">
	<tr>
		<td>No</td>
		<td>zero-bit</td>
		<td>NRI</td>
		<td>Type</td>
		<td>Type String</td>
	</tr>
	`)
	index := 0
	for i := 0; i < len(tp.frameByte)-4; i++ {
		if tp.frameByte[i] == 0 && tp.frameByte[i+1] == 0 && tp.frameByte[i+2] == 1 {
			b.SetPos(i + 3)
			f := uint32(b.Read(1))
			nri := uint32(b.Read(2))
			nalType := uint32(b.Read(5))
			tp.write(fmt.Sprintf(`<tr>
				<td>%d</td>
				<td>%d</td>
				<td>%d</td>
				<td>%d</td>
				<td>%s</td>
			</tr>
			`, index, f, nri, nalType, getNaluType(nalType)))
			index++
		}
	}
	tp.write("</table>")

	return err
}

// BitReader ...
type BitReader struct {
	pos    int
	offs   uint32 // 0-7
	buffer []byte
	EOF    bool
}

// SetPos ...
func (r *BitReader) SetPos(n int) {
	if n < len(r.buffer)-1 {
		r.pos = n
		r.offs = 0
	}
}

// Read ...
func (r *BitReader) Read(n uint32) uint64 {
	var d uint32
	var v uint64
	for n > 0 {
		if r.pos >= len(r.buffer) {
			r.EOF = true
			return 0
		}
		if r.offs+n > 8 {
			d = 8 - r.offs
		} else {
			d = n
		}
		v = v << d
		v += (uint64(r.buffer[r.pos] >> (8 - r.offs - d))) & (0xff >> (8 - d))
		r.offs += d
		n -= d

		if r.offs == 8 {
			r.pos++
			r.offs = 0
		}
	}

	return v
}

// Skip ...
func (r *BitReader) Skip(n uint32) {
	var d uint32
	for n > 0 {
		if r.pos >= len(r.buffer) {
			r.EOF = true
			return
		}
		if r.offs+n > 8 {
			d = 8 - r.offs
		} else {
			d = n
		}
		r.offs += d
		n -= d

		if r.offs == 8 {
			r.pos++
			r.offs = 0
		}
	}
}

// GetField ...
func (r *BitReader) GetField(s string, n uint32) uint64 {
	v := r.Read(n)
	fmt.Printf("%s: %d\n", s, int(v))
	return v
}

// GetFieldHex ...
func (r *BitReader) GetFieldHex(s string, n uint32) uint64 {
	v := r.Read(n)
	if n <= 8 {
		fmt.Printf("%s: 0x%02x\n", s, v)
	} else {
		fmt.Printf("%s: 0x%04x\n", s, v)
	}
	return v
}

// GetFieldValue ...
// 转换高低字节序
// n 为字节数
func (r *BitReader) GetFieldValue(s string, n uint32) uint64 {
	var v uint64
	var i uint32
	for ; i < n; i++ {
		v += uint64(uint32(r.Read(8)) << (i * 8))
	}
	fmt.Printf("%s: %d\n", s, int(v))
	return v
}
