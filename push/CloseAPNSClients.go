package push

import (
	"log"

	"golang.org/x/net/http2"
)

// CloseAPNSClients 关闭所有APNS客户端资源
func CloseAPNSClients() {
	// 关闭channel并清理资源
	if CLIENTS != nil {
		// 尝试关闭所有客户端连接
		clientCount := len(CLIENTS)
		for i := 0; i < clientCount; i++ {
			select {
			case client := <-CLIENTS:
				// 如果客户端有需要特别关闭的资源，可以在这里处理
				// 例如关闭HTTP客户端的连接池等
				if client != nil && client.HTTPClient != nil && client.HTTPClient.Transport != nil {
					// 尝试关闭transport
					if transport, ok := client.HTTPClient.Transport.(*http2.Transport); ok && transport != nil {
						transport.CloseIdleConnections()
					}
				}
			default:
				// channel已空
				break
			}
		}

		// 记录关闭信息
		log.Println("All APNS clients have been closed")
	}
}
