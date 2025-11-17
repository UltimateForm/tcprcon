package client

import "net"

type RCONClient struct {
	Address string
	con     net.Conn
	count   int32
}

func (src *RCONClient) Id() int32 {
	return src.count
}

func (src *RCONClient) Read(p []byte) (n int, err error) {
	return src.con.Read(p)
}

func (src *RCONClient) Write(p []byte) (n int, err error) {
	defer func() {
		src.count++
	}()
	return src.con.Write(p)
}

func (src *RCONClient) Close() error {
	return src.con.Close()
}

func New(address string) (*RCONClient, error) {
	con, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &RCONClient{
		Address: address,
		con:     con,
		count:   0,
	}, nil
}
