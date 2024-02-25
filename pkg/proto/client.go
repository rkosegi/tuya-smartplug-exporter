/*
Copyright 2022 Richard Kosegi

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package proto

import (
	"crypto/aes"
	"encoding/binary"
	"encoding/json"
	"hash/crc32"
	"net"

	"github.com/pkg/errors"
)

var (
	prefix = []byte{0x00, 0x00, 0x55, 0xaa, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	infix  = []byte{0x0a, 0x00, 0x00, 0x00}
	suffix = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xaa, 0x55}
)

type Dps struct {
	SwitchOn bool `json:"1"`
	Current  int  `json:"18"`
	Power    int  `json:"19"`
	Voltage  int  `json:"20"`
}

type Response struct {
	Dps Dps `json:"dps"`
}

type Req struct {
	GwId  string `json:"gwId,omitempty"`
	DevId string `json:"devId,omitempty"`
}

type Proto interface {
	Status() (*Response, error)
}

type proto struct {
	key []byte
	ip  string
	id  string
}

func (p *proto) Status() (*Response, error) {
	var resp Response
	data, err := p.exchange()
	if err != nil {
		return nil, err
	}
	data, err = p.decryptResponse(data[20 : len(data)-8])
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(unpad(data), &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func NewClient(ip string, id string, key []byte) Proto {
	return &proto{
		key: key,
		ip:  ip,
		id:  id,
	}
}

func (p *proto) decryptResponse(data []byte) ([]byte, error) {
	cipher, err := aes.NewCipher(p.key)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(data))
	for i, j := 0, 16; i < len(data); i, j = i+16, j+16 {
		cipher.Decrypt(out[i:j], data[i:j])
	}
	return out, nil
}

func pad(text []byte) []byte {
	pad := 16 - (len(text) % 16)
	if pad == 0 {
		pad = 16
	}
	for i := 0; i < pad; i++ {
		text = append(text, byte(pad))
	}
	return text
}

func unpad(text []byte) []byte {
	padding := text[len(text)-1]
	return text[:len(text)-int(padding)]
}

func (p *proto) encryptRequest() ([]byte, error) {
	r := &Req{
		GwId:  p.id,
		DevId: p.id,
	}
	data, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	data = pad(data)
	block, err := aes.NewCipher(p.key)
	if err != nil {
		return nil, err
	}
	encrypted := make([]byte, len(data))
	for i, j := 0, 16; i < len(data); i, j = i+16, j+16 {
		block.Encrypt(encrypted[i:j], data[i:j])
	}
	encrypted = append(encrypted, suffix...)
	result := make([]byte, 0)
	result = append(result, prefix...)
	result = append(result, infix...)
	result = append(result, byte(len(encrypted)))
	result = append(result, encrypted...)
	crc := crc32.ChecksumIEEE(result[:len(result)-8])
	result2 := result[:len(result)-8]
	crcbytes := make([]byte, 4)
	binary.BigEndian.PutUint32(crcbytes, crc)
	result2 = append(result2, crcbytes...)
	result2 = append(result2, result[len(result)-4:]...)
	return result2, nil
}

func (p *proto) exchange() ([]byte, error) {
	data, err := p.encryptRequest()
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial("tcp", net.JoinHostPort(p.ip, "6668"))
	if err != nil {
		return nil, err
	}
	_, err = conn.Write(data)
	if err != nil {
		return nil, err
	}
	inbuffer := make([]byte, 256)
	read, err := conn.Read(inbuffer)
	if err != nil {
		return nil, err
	}
	if read == 0 {
		return nil, errors.Wrap(err, "no data received from device")
	}
	err = conn.Close()
	if err != nil {
		return nil, err
	}
	return inbuffer[0:read], nil
}
