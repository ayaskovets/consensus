package net

func (client *Client) Addr() string {
	return client.addr
}

func (server *Server) Addr() string {
	return server.addr
}
