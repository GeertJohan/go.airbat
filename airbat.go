package airbat

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
)

const airbatURL = `http://airb.at/%s`

// ErrInvalidID is returned by conversion functions when the given id is of zero value (0 or "")
var ErrInvalidID = errors.New("invalid id")

func toAirbat(id uint64) (string, error) {
	// check input value
	if id == 0 {
		return "", ErrInvalidID
	}

	// put unsigned varint into []byte buffer
	buf := make([]byte, 10)
	n := binary.PutUvarint(buf, id)

	// encode with base64 URL encoding
	code := base64.URLEncoding.EncodeToString(buf[:n])

	// all done
	return code, nil
}

func fromAirbat(code string) (uint64, error) {
	// check input value
	if code == "" {
		return 0, ErrInvalidID
	}

	// base64 decode
	buf, err := base64.URLEncoding.DecodeString(code)
	if err != nil {
		return 0, err
	}

	id, _ := binary.Uvarint(buf)
	return id, nil
}

// UintToAirbatCode converts the given notice-id to a airbat-code
func UintToAirbatCode(id uint64) (string, error) {
	return toAirbat(id)
}

// UintToAirbatURL converts the given notice-id to a airbat url
func UintToAirbatURL(id uint64) (string, error) {
	code, err := toAirbat(id)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(airbatURL, code), nil
}

// AirbatCodeToUint converts the given airbat code to an notice-id
func AirbatCodeToUint(code string) (uint64, error) {
	return fromAirbat(code)
}
