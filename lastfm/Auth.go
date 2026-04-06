package lastfm

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
)

func OpenBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Run()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Run()
	}

	if err != nil {
		fmt.Printf("error opening browser: %v\n", err)
	}
}

func StartAuthServer(apiKey string) (string, error) {
	tokenC := make(chan string)
	mux := http.NewServeMux()
	server := &http.Server{Addr: ":8080", Handler: mux}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		t := r.URL.Query().Get("token")
		if t != "" {
			fmt.Fprintf(w, "<h1>Login Successful!</h1><p>You can close this tab now.</p>")
			tokenC <- t
		}
	})

	go server.ListenAndServe()

	authURL := fmt.Sprintf("https://www.last.fm/api/auth/?api_key=%s&cb=http://localhost:8080/callback", apiKey)
	OpenBrowser(authURL)

	token := <-tokenC

	server.Shutdown(context.Background())

	return token, nil
}
