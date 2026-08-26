package dicom

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"
)

type Parser struct {
	MaxElementBytes uint32
	MaxFileBytes    int64
}

func (p Parser) Parse(r io.Reader) (Dataset, error) {
	data, err := io.ReadAll(io.LimitReader(r, p.MaxFileBytes+1))
	if err != nil {
		return Dataset{}, fmt.Errorf("read Part 10 stream: %w", err)
	}
	if int64(len(data)) > p.MaxFileBytes {
		return Dataset{}, tooLarge(0, "file exceeds limit")
	}
	if len(data) < 132 || string(data[128:132]) != "DICM" {
		return Dataset{}, malformed(0, "missing Part 10 preamble or DICM prefix")
	}
	d := Dataset{TransferSyntax: "1.2.840.10008.1.2.1", ReceivedAt: time.Now().UTC(), Size: len(data)}
	sum := sha256.Sum256(data)
	d.SHA256 = hex.EncodeToString(sum[:])
	off := 132
	for off < len(data) {
		e, n, err := p.element(data, off, true)
		if err != nil {
			return Dataset{}, fmt.Errorf("parse element at offset %d: %w", off, err)
		}
		off = n
		d.Elements = append(d.Elements, e)
		p.populate(&d, e)
	}
	return d, nil
}
func (p Parser) element(data []byte, off int, explicit bool) (Element, int, error) {
	if len(data)-off < 8 {
		return Element{}, off, malformed(off, "truncated element header")
	}
	g := binary.LittleEndian.Uint16(data[off:])
	el := binary.LittleEndian.Uint16(data[off+2:])
	vr := "UN"
	head := 8
	var length uint32
	if explicit {
		vr = string(data[off+4 : off+6])
		if !validVR(vr) {
			return Element{}, off, malformed(off, "invalid value representation")
		}
		if longVR(vr) {
			if len(data)-off < 12 {
				return Element{}, off, malformed(off, "truncated long header")
			}
			length = binary.LittleEndian.Uint32(data[off+8:])
			head = 12
		} else {
			length = uint32(binary.LittleEndian.Uint16(data[off+6 : off+8]))
		}
	} else {
		length = binary.LittleEndian.Uint32(data[off+4 : off+8])
		head = 8
	}
	if length != 0xffffffff && length > p.MaxElementBytes {
		return Element{}, off, tooLarge(off, "element exceeds limit")
	}
	start := off + head
	if length == 0xffffffff {
		return p.fragments(data, off, start, Tag{g, el}, vr)
	}
	end64 := int64(start) + int64(length)
	if end64 > int64(len(data)) {
		return Element{}, off, malformed(off, "element length exceeds input")
	}
	end := int(end64)
	value := append([]byte(nil), data[start:end]...)
	return Element{Tag: Tag{g, el}, VR: vr, Length: length, Value: value}, end, nil
}
func (p Parser) fragments(data []byte, off, start int, tag Tag, vr string) (Element, int, error) {
	if tag.Group != 0x7fe0 || tag.Element != 0x0010 {
		return Element{}, off, malformed(off, "undefined length only supported for pixel data")
	}
	pos := start
	var frags [][]byte
	for {
		if len(data)-pos < 8 {
			return Element{}, off, malformed(pos, "truncated fragment")
		}
		g := binary.LittleEndian.Uint16(data[pos:])
		e := binary.LittleEndian.Uint16(data[pos+2:])
		n := binary.LittleEndian.Uint32(data[pos+4:])
		pos += 8
		if g == 0xfffe && e == 0xe0dd {
			if n != 0 {
				return Element{}, off, malformed(pos, "invalid sequence delimiter")
			}
			break
		}
		if g != 0xfffe || e != 0xe000 {
			return Element{}, off, malformed(pos, "invalid fragment item")
		}
		if n > p.MaxElementBytes {
			return Element{}, off, tooLarge(pos, "fragment exceeds limit")
		}
		if int64(pos)+int64(n) > int64(len(data)) {
			return Element{}, off, malformed(pos, "fragment length exceeds input")
		}
		frags = append(frags, append([]byte(nil), data[pos:pos+int(n)]...))
		pos += int(n)
	}
	return Element{Tag: tag, VR: vr, Length: 0xffffffff, Fragments: frags}, pos, nil
}
func validVR(v string) bool {
	return len(v) == 2 && strings.IndexByte("AEASATCSCSSHDAWLOFLLTOBODOFOWSQURUTUN", v[0]) >= 0 && strings.IndexByte("AEASATCSCSSHDAWLOFLLTOBODOFOWSQURUTUN", v[1]) >= 0
}
func longVR(v string) bool {
	return v == "OB" || v == "OD" || v == "OF" || v == "OL" || v == "OW" || v == "SQ" || v == "UC" || v == "UR" || v == "UT" || v == "UN"
}
func (p Parser) populate(d *Dataset, e Element) {
	switch e.Tag {
	case Tag{0x0020, 0x000d}:
		d.StudyUID = strings.TrimSpace(e.Text())
	case Tag{0x0020, 0x000e}:
		d.SeriesUID = strings.TrimSpace(e.Text())
	case Tag{0x0008, 0x0018}:
		d.SOPInstanceUID = strings.TrimSpace(e.Text())
	case Tag{0x0010, 0x0010}:
		d.PatientName = e.Text()
	case Tag{0x0010, 0x0020}:
		d.PatientID = e.Text()
	case Tag{0x0008, 0x0060}:
		d.Modality = strings.TrimSpace(e.Text())
	}
}
func EncodePart10(d Dataset) ([]byte, error) {
	var b bytes.Buffer
	b.Write(make([]byte, 128))
	b.WriteString("DICM")
	for _, e := range d.Elements {
		if e.Length == 0xffffffff {
			return nil, fmt.Errorf("encoding fragments not implemented")
		}
		if len(e.Value) > 65535 && !longVR(e.VR) {
			return nil, fmt.Errorf("element too long")
		}
		binary.Write(&b, binary.LittleEndian, e.Tag.Group)
		binary.Write(&b, binary.LittleEndian, e.Tag.Element)
		vr := e.VR
		if len(vr) != 2 {
			vr = "UN"
		}
		b.WriteString(vr)
		if longVR(vr) {
			b.Write([]byte{0, 0})
			binary.Write(&b, binary.LittleEndian, uint32(len(e.Value)))
		} else {
			binary.Write(&b, binary.LittleEndian, uint16(len(e.Value)))
		}
		b.Write(e.Value)
	}
	return b.Bytes(), nil
}
