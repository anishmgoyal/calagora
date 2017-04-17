package utils

import (
	"bytes"
	"encoding/binary"
	"os"
)

const (
	orientationTag = 0x0112
)

func findExifBlock(file *os.File) []byte {
	file.Seek(0, 0)

	buff := make([]byte, 6)

	if n, err := file.Read(buff[:3]); err != nil || n != 3 {
		return nil
	}

	if buff[0] != 0xFF ||
		buff[1] != 0xD8 ||
		buff[2] != 0xFF {
		return nil
	}

	for {
		if n, err := file.Read(buff[:1]); err != nil || n != 1 {
			return nil
		}

		var length int16
		if err := binary.Read(file, binary.BigEndian, &length); err != nil {
			return nil
		}

		if buff[0] != 0xE1 {
			if _, err := file.Seek(int64(length-2), os.SEEK_CUR); err != nil {
				return nil
			}
			var b byte
			if err := binary.Read(file, binary.BigEndian, &b); err != nil ||
				b != 0xFF {

				return nil
			}
			continue
		} else {

			if n, err := file.Read(buff[:6]); err != nil || n != 6 ||
				!bytes.Equal(buff[:6], append([]byte("Exif"), 0x00, 0x00)) {

			} else {
				if err != nil {
					return nil
				}

				if length < 2 {
					return nil
				}

				buff = make([]byte, length-2)
				if n, err := file.Read(buff); err != nil || n != int(length)-2 {
					return nil
				}
				return buff
			}
			return nil
		}
	}
}

func findOrientationInExifBlock(block []byte) int {
	r := bytes.NewReader(block)
	var endian binary.ByteOrder
	var buff = make([]byte, 2)
	if n, err := r.Read(buff); err != nil || n != 2 {
		return 1
	}
	if buff[0] == 'I' && buff[1] == 'I' {
		endian = binary.LittleEndian
	} else if buff[0] == 'M' && buff[1] == 'M' {
		endian = binary.BigEndian
	}

	var magic int16
	if err := binary.Read(r, endian, &magic); err != nil || magic != 0x002A {
		return 1
	}

	if _, err := r.Seek(4, os.SEEK_CUR); err != nil {
		return 1
	}

	var numTags int16
	if err := binary.Read(r, endian, &numTags); err != nil {
		return 1
	}

	for i := int16(0); i < numTags; i++ {
		var tagNum int16
		if err := binary.Read(r, endian, &tagNum); err != nil {
			return 1
		}

		if tagNum == orientationTag {
			var valType int16
			var numVals int32
			var value int16
			if err := binary.Read(r, endian, &valType); err != nil || valType != 3 {
				return 1
			}
			if err := binary.Read(r, endian, &numVals); err != nil || numVals != 1 {
				return 1
			}
			if err := binary.Read(r, endian, &value); err != nil {
				return 1
			}
			return int(value)
		}

		// If we're here, that was the wrong tag... skip to the next one
		r.Seek(10, os.SEEK_CUR)
	}

	return 1
}

func getImageExifOrientation(file *os.File) int {
	block := findExifBlock(file)
	if block == nil {
		return 1
	}
	return findOrientationInExifBlock(block)
}
