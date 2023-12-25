package net

// Returns the address that is used to connect to the client
func (client *Client) Addr() string {
	return client.addr
}

// Returns the address that the server accepts incoming connection on
func (server *Server) Addr() string {
	return server.addr
}
