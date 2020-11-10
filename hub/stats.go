package hub

func MpsFromNs(ns float64) float64 {
	return 1 / (ns * 1e-9)
}

/*
func (c *Client) report() Report {

	g
	defer func() {
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.WithField("error", err).Error("statsReporter ReadMessage")
			}
			break
		}

		for _, topic := range c.hub.clients {
			for client, _ := range topic {

				//c.conn.SetWriteDeadline(time.Now().Add(writeWait))
				//
				//w, err := c.conn.NextWriter(websocket.TextMessage)
				//if err != nil {
				//	return
				//}

				var tx ReportStats

				if client.stats.tx.size.Count() > 0 {
					tx = ReportStats{
						Last: time.Since(client.stats.tx.last).String(),
						Size: math.Round(client.stats.tx.size.Mean()),
						Fps:  fpsFromNs(client.stats.tx.ns.Mean()),
					}
				} else {
					tx = ReportStats{
						Last: "Never",
						Size: 0,
						Fps:  0,
					}
				}

				var rx ReportStats

				if client.stats.rx.size.Count() > 0 {
					rx = ReportStats{
						Last: time.Since(client.stats.rx.last).String(),
						Size: math.Round(client.stats.rx.size.Mean()),
						Fps:  fpsFromNs(client.stats.rx.ns.Mean()),
					}
				} else {
					rx = ReportStats{
						Last: "Never",
						Size: 0,
						Fps:  0,
					}
				}

				report := &ClientReport{
					Topic:       client.topic,
					Broadcaster: client.broadcaster,
					Connected:   client.stats.connectedAt.String(),
					RemoteAddr:  client.remoteAddr,
					UserAgent:   client.userAgent,
					Stats: RxTx{
						Tx: tx,
						Rx: rx,
					},
				}

				b, err := json.Marshal(report)

				if err != nil {
					log.WithField("error", err).Error("statsReporter marshalling JSON")
					return
				} else {
					c.send <- message{data: b, mt: websocket.TextMessage}
					//w.Write(b)
				}

				//if err := w.Close(); err != nil {
				//	return
				//}
			}
		}

	}
}
*/
