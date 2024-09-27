package main

import (
	"fmt"
	"net/http"
)

func main() {
	// 注意：Go 为编译型语言，直接修改代码不能直接生效！请在控制台右上角“导出代码”，然后根据 README.md 中的说明编译代码并重新上传。
	// 注意：Go 为编译型语言，直接修改代码不能直接生效！请在控制台右上角“导出代码”，然后根据 README.md 中的说明编译代码并重新上传。
	// 注意：Go 为编译型语言，直接修改代码不能直接生效！请在控制台右上角“导出代码”，然后根据 README.md 中的说明编译代码并重新上传。
	// Notice: You need to complie the code first otherwise the code change will not take effect.
	fmt.Println("start listen port 9000")
	http.HandleFunc("/", HelloServer)
	http.ListenAndServe(":9000", nil)
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	// 注意：Go 为编译型语言，直接修改代码不能直接生效！请在控制台右上角“导出代码”，然后根据 README.md 中的说明编译代码并重新上传。
	// 注意：Go 为编译型语言，直接修改代码不能直接生效！请在控制台右上角“导出代码”，然后根据 README.md 中的说明编译代码并重新上传。
	// 注意：Go 为编译型语言，直接修改代码不能直接生效！请在控制台右上角“导出代码”，然后根据 README.md 中的说明编译代码并重新上传。
	// Notice: You need to complie the code first otherwise the code change will not take effect.
	// requestId := r.Header.Get("x-fc-request-id")
	// fcLogger := gr.GetLogger().WithField("requestId", requestId)
	// fcLogger.Infof("This is a log from golang!")
	// fmt.Fprintf(w, "Hello, World!")

	// 实现类型转换以取得底层的TCP连接
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
		return
	}
	conn, buf, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Manually write the HTTP response
	buf.WriteString("HTTP/1.1 200 OK\r\n")
	buf.WriteString("Content-Type: application/json\r\n")
	// 响应的body数据小于100，触发unexpected EOF报错
	buf.WriteString("Content-Length: 100\r\n")
	buf.WriteString("\r\n")
	buf.WriteString("Hello, ")
	buf.Flush()
}
