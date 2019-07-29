package main

import (
	"log"
	"net/http"
	"os"

	"github.com/geckoboard/slash-infra/slackutil"
	"github.com/joho/godotenv"
)

var SyrupBaseURL string

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)

	err := godotenv.Load()
	if err != nil {
		log.Println("could not load .env file", err)
	}

	server := makeHttpHandler()

	handler := slackutil.VerifyRequestSignature(os.Getenv("SLACK_SIGNING_SECRET"))(server)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	SyrupBaseURL = os.Getenv("SYRUP_BASE_URL")
	if SyrupBaseURL == "" {
		SyrupBaseURL = "http://localhost:8000"
	}

	loggingMiddleware := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())

		handler.ServeHTTP(w, r)
	})
	log.Fatal(http.ListenAndServe(":"+port, loggingMiddleware))
}
