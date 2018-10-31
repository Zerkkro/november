package compress

import (
	"bufio"
	"errors"
	// "fmt"
	"gitee.com/johng/gf/g/encoding/gbinary"
	"gitee.com/johng/gf/g/encoding/gjson"
	"gitee.com/johng/gf/g/os/glog"
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

// LocalFileHeader ...
type LocalFileHeader struct {
	Signature        uint32 `json:"signature"`
	MinVersion       uint16 `json:"min_ersion"`
	Flag             uint16 `json:"flag"`
	Method           uint16 `json:"method"`
	LastModifyTime   uint16 `json:"last_modify_time"`
	LastModifyDate   uint16 `json:"last_modify_date"`
	Crc32            uint32 `json:"crc32"`
	CompressedSize   uint32 `json:"compressed_size"`
	UncompressedSize uint32 `json:"uncompressed_size"`
	FileNameLength   uint16 `json:"file_name_length"`
	ExtraFieldLength uint16 `json:"extra_field_length"`
	FileName         string `json:"file_name"`
}

type apkParser struct {
	filePath string
	fs       *os.File
	w        io.Writer
	j        *gjson.Json
}

// ParseApkFile ...
func ParseApkFile(path string, w io.Writer) error {
	ap := &apkParser{
		filePath: path,
		w:        w,
	}
	ap.j, _ = gjson.DecodeToJson([]byte(`{
	"apk": {
		"filename": "xxx",
		"zip_entries":[]
	}
}`))
	ap.j.Set("apk.filename", path)
	return ap.parseAPKFile()
}

func (ap *apkParser) parseAPKFile() error {
	var err error
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}
	dir = strings.Replace(dir, "\\", "/", -1)
	dir = dir + "/"
	ap.filePath = dir + ap.filePath
	if ap.fs, err = os.Open(ap.filePath); err != nil {
		return err
	}
	defer ap.fs.Close()

	r := bufio.NewReader(ap.fs)

	// entries
	if err = ap.parseZipEntries(r); err != nil {
		return err
	}

	// apk sign block
	if err = ap.parseApkSignBlock(r); err != nil {
		return err
	}

	c, _ := ap.j.ToJson()
	_, err = ap.w.Write(c)
	return err
}

func (ap *apkParser) parseZipEntries(r *bufio.Reader) error {
	var err error
	// entriesStrings := []string{}
	// h := LocalFileHeader{}
	var b []byte
	hs := []LocalFileHeader{}
	for i := 1; ; i++ {
		h := LocalFileHeader{}
		if b, err = r.Peek(30); err != nil {
			return err
		}
		h.Signature = gbinary.DecodeToUint32(b[0:4])
		if h.Signature != 0x04034b50 {
			break
		}
		h.MinVersion = gbinary.DecodeToUint16(b[4:6])
		h.Flag = gbinary.DecodeToUint16(b[6:8])
		flagBits := gbinary.DecodeBytesToBits(b[6:8])
		Flag := gbinary.DecodeBits(flagBits[3:4])
		if Flag > 0 {
			return errors.New("not supported")
		}
		h.Method = gbinary.DecodeToUint16(b[8:10])
		h.LastModifyTime = gbinary.DecodeToUint16(b[10:12])
		h.LastModifyDate = gbinary.DecodeToUint16(b[12:14])
		h.Crc32 = gbinary.DecodeToUint32(b[14:18])
		h.CompressedSize = gbinary.DecodeToUint32(b[18:22])
		h.UncompressedSize = gbinary.DecodeToUint32(b[22:26])
		h.FileNameLength = gbinary.DecodeToUint16(b[26:28])
		h.ExtraFieldLength = gbinary.DecodeToUint16(b[28:30])
		r.Discard(30)
		if h.FileNameLength > 512 {
			return errors.New("file name is to long")
		}
		if b, err = r.Peek(int(h.FileNameLength)); err != nil {
			return err
		}
		h.FileName = string(b)
		r.Discard(int(h.FileNameLength + h.ExtraFieldLength))

		// compressed file
		r.Discard(int(h.CompressedSize))
		// c, _ := ap.j.ToJson(F
		// fmt.Printf(">>> %s\nE, string(c))
		// entriesStrings = Fppend(entriesStrings, string(c))
		// if len(hs) < 10 {
		hs = append(hs, h)
		// }
	}
	// if len(entriesStrings) > 0 {
	// 	ap.j.Set("zip_entries", entriesStrings)
	// }
	ap.j.Set("apk.zip_entries", hs)
	// c, _ := ap.j.ToJson()
	// fmt.Printf(">>> %s\n", string(c))
	return err
}

func (ap *apkParser) parseApkSignBlock(r *bufio.Reader) error {
	var err error
	var b []byte
	if b, err = r.Peek(8); err != nil {
		return err
	}
	blockSize1 := gbinary.DecodeToUint64(b)
	var total uint64
	r.Discard(8)
	pos := 0
	for {
		if b, err = r.Peek(8); err != nil {
			return err
		}
		partSize := gbinary.DecodeToUint64(b)
		r.Discard(8)
		if b, err = r.Peek(int(partSize)); err != nil {
			return err
		}
		id := gbinary.DecodeToUint32(b[0:4])
		// payload := string(b[4:])
		r.Discard(int(partSize))
		glog.Infof("id: 0x%02x, part size: %d\n", id, partSize)
		pos += 8 + int(partSize)
		total += 8 + partSize
		if pos+24 >= int(blockSize1) {
			break
		}
	}
	if b, err = r.Peek(8); err != nil {
		return err
	}
	blockSize2 := gbinary.DecodeToUint64(b)
	total += 8
	glog.Infof("blockSize2: %d, total: %d\n", blockSize2, total)

	r.Discard(8)
	if b, err = r.Peek(16); err != nil {
		return err
	}
	glog.Infof("magic part: %s\n", string(b))

	return err
}
