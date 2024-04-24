package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
	"github.com/tmaxmax/go-sse"
)

var sseServer *sse.Server

type Message struct {
	Data  string `json:"data"`
	Event string `json:"event"`
}

func setupConfig() {
	viper.SetDefault("JWT_SECRET_KEY", "supersecret")
	viper.SetDefault("API_KEY", "supersecret")
	viper.SetDefault("PORT", "8088")
	viper.SetDefault("SSE_PATH", "/events")
	viper.SetDefault("PUBLISH_PATH", "/publish")
	viper.SetDefault("TOPICS", []string{})
	viper.AutomaticEnv()
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(viper.GetString("JWT_SECRET_KEY")), nil
	})
}

func validateAPIKey(apiKey string) bool {
	return apiKey == viper.GetString("API_KEY")
}

func publishMessage(m Message) {
	message := &sse.Message{}
	message.AppendData(m.Data)
	message.Type = sse.Type(m.Event)
	sseServer.Publish(message)
}

func publisher(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	apiKey := req.Header.Get("X-API-KEY")
	if !validateAPIKey(apiKey) {
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	var message Message
	if err := json.NewDecoder(req.Body).Decode(&message); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	publishMessage(message)
}

func onSSESession(s *sse.Session) (sse.Subscription, bool) {
	authHeader := s.Req.Header.Get("Authorization")
	if authHeader == "" {
		log.Printf("Unauthorized session: %s", s.Req.RemoteAddr)
		s.Res.WriteHeader(http.StatusUnauthorized)
		s.Res.Write([]byte("Authorization header is missing"))
		return sse.Subscription{}, false
	}
	if _, err := validateJWT(authHeader); err != nil {
		log.Printf("Invalid JWT: %s", err)
		s.Res.WriteHeader(http.StatusUnauthorized)
		s.Res.Write([]byte(err.Error()))
		return sse.Subscription{}, false
	}

	log.Printf("New session: %s", s.Req.RemoteAddr)

	topics := viper.GetStringSlice("TOPICS")

	return sse.Subscription{
		Client:      s,
		LastEventID: s.LastEventID,
		Topics:      append(topics, sse.DefaultTopic),
	}, true
}

func main() {
	setupConfig()

	sseServer = &sse.Server{
		OnSession: onSSESession,
	}

	mux := http.NewServeMux()
	mux.HandleFunc(viper.GetString("PUBLISH_PATH"), publisher)
	mux.Handle(viper.GetString("SSE_PATH"), sseServer)

	port := ":" + viper.GetString("PORT")

	log.Printf("Server started at %s", port)
	log.Fatal(http.ListenAndServe(port, mux))
}
