package app

import (
	"errors"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"pingtool/internal/scan"
	"pingtool/internal/web"
)

func Run() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	scanner := scan.NewService()

	handler, err := web.NewHandler(scanner)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           logRequests(handler),
		ReadHeaderTimeout: 5 * time.Second,
	}

	if cfg.OpenBrowser {
		go openBrowser(cfg.URL())
	}

	log.Printf("Ping 检测台已启动: %s", cfg.URL())

	err = server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

func openBrowser(url string) {
	time.Sleep(300 * time.Millisecond)

	var err error
	switch runtime.GOOS {
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = exec.Command("xdg-open", url).Start()
	}

	if err != nil {
		log.Printf("自动打开浏览器失败，请手动访问 %s", url)
	}
}
