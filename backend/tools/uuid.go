package tools

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
)

// UUID contains data for an RFC4122 version 4 UUID
type UUID struct {
	Data [16]byte
}

// UUIDGen generates an RFC4122 version 4 UUID
func UUIDGen() *UUID {
	res := new(UUID)
	rand.Read(res.Data[0:])

	res.Data[8] = res.Data[8] & 63
	res.Data[8] = res.Data[8] | 0x80

	res.Data[6] = res.Data[6] & 0x0F
	res.Data[6] = res.Data[6] | 0x40

	return res
}

// AsSlice returns the UUID value as a byte slice.
func (u *UUID) AsSlice() []byte {
	return u.Data[:]
}

func (u *UUID) String() string {
	return fmt.Sprintf("%X-%X-%X-%X-%X", u.Data[0:4], u.Data[4:6], u.Data[6:8], u.Data[8:10], u.Data[10:])
}

// MarshalJSON allows the JSON serializer to serialize a UUID
func (u *UUID) MarshalJSON() ([]byte, error) {
	text := fmt.Sprintf("\"%s\"", u.String())
	return ([]byte)(text), nil
}

// UnmarshalJSON allows the JSON serializer to deserialize a UUID
func (u *UUID) UnmarshalJSON(data []byte) error {
	text := string(data)

	if len(text) < 2 {
		return fmt.Errorf("UUID too short")
	}

	withoutInvCommas := text[1 : len(text)-1]

	ok := u.Parse(withoutInvCommas)
	if !ok {
		return fmt.Errorf("unable to deserialize UUID")
	}

	return nil
}

// IsEqual if a and rhs denote the same UUID
func (u *UUID) IsEqual(rhs *UUID) bool {
	return u.Data == rhs.Data
}

// Parse parses a uuid and fills the data component accordingly. Returns false in case of a failure.
func (u *UUID) Parse(uuid string) bool {
	hexchars := "[0123456789abcdefABCDEF]"
	expStr := fmt.Sprintf("^(%s{8})-(%s{4})-(%s{4})-(%s{4})-(%s{12})$", hexchars, hexchars, hexchars, hexchars, hexchars)
	uuidRegExp := regexp.MustCompile(expStr)

	matches := uuidRegExp.FindStringSubmatch(uuid)
	if matches == nil {
		return false
	}

	help := u.Data[0:4]
	temp, _ := hex.DecodeString(matches[1])
	copy(help, temp)

	help = u.Data[4:6]
	temp, _ = hex.DecodeString(matches[2])
	copy(help, temp)

	help = u.Data[6:8]
	temp, _ = hex.DecodeString(matches[3])
	copy(help, temp)

	help = u.Data[8:10]
	temp, _ = hex.DecodeString(matches[4])
	copy(help, temp)

	help = u.Data[10:]
	temp, _ = hex.DecodeString(matches[5])
	copy(help, temp)

	return true
}
