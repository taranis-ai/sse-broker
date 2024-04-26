package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
	"github.com/tmaxmax/go-sse"
)

type App struct {
	Config    Config
	SSEServer *sse.Server
}

type Config struct {
	JWTSecretKey string   `mapstructure:"JWT_SECRET_KEY"`
	APIKey       string   `mapstructure:"API_KEY"`
	Port         string   `mapstructure:"PORT"`
	SSEPath      string   `mapstructure:"SSE_PATH"`
	PublishPath  string   `mapstructure:"PUBLISH_PATH"`
	Topics       []string `mapstructure:"TOPICS"`
}
type Message struct {
	Data  string `json:"data"`
	Event string `json:"event"`
}

func setupConfig() Config {
	viper.SetDefault("JWT_SECRET_KEY", "supersecret")
	viper.SetDefault("API_KEY", "supersecret")
	viper.SetDefault("PORT", "8088")
	viper.SetDefault("SSE_PATH", "/events")
	viper.SetDefault("PUBLISH_PATH", "/publish")
	viper.SetDefault("TOPICS", []string{})
	viper.AutomaticEnv()

	config := Config{}
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to decode config into struct, %v", err)
	}
	return config
}

func validateJWT(tokenString string, jwtKey string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtKey), nil
	})
}

func validateAPIKey(apiKey string, configApiKey string) bool {
	return apiKey == configApiKey
}

func (app *App) publishMessage(m Message) {
	message := &sse.Message{}
	message.AppendData(m.Data)
	message.Type = sse.Type(m.Event)
	app.SSEServer.Publish(message)
}

func (app *App) publisher(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	apiKey := req.Header.Get("X-API-KEY")
	if !validateAPIKey(apiKey, app.Config.APIKey) {
		http.Error(w, "Invalid API key", http.StatusUnauthorized)
		return
	}

	var message Message
	if err := json.NewDecoder(req.Body).Decode(&message); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	app.publishMessage(message)
}

func onSSESession(config Config) func(s *sse.Session) (sse.Subscription, bool) {
	return func(s *sse.Session) (sse.Subscription, bool) {
		authHeader := s.Req.Header.Get("Authorization")
		jwtQuery := s.Req.URL.Query().Get("jwt")
		tokenString := ""
		if authHeader != "" {
			tokenString = authHeader
		} else if jwtQuery != "" {
			tokenString = jwtQuery
		} else {
			log.Printf("Unauthorized session: %s", s.Req.RemoteAddr)
			s.Res.WriteHeader(http.StatusUnauthorized)
			s.Res.Write([]byte("Authorization header is missing"))
			return sse.Subscription{}, false
		}

		if _, err := validateJWT(tokenString, config.JWTSecretKey); err != nil {
			log.Printf("Invalid JWT: %s", err)
			s.Res.WriteHeader(http.StatusUnauthorized)
			s.Res.Write([]byte(err.Error()))
			return sse.Subscription{}, false
		}

		log.Printf("New session: %s", s.Req.RemoteAddr)

		return sse.Subscription{
			Client:      s,
			LastEventID: s.LastEventID,
			Topics:      append(config.Topics, sse.DefaultTopic),
		}, true
	}
}

func (app *App) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(app.Config.PublishPath, app.publisher)
	mux.Handle(app.Config.SSEPath, app.SSEServer)
	return mux
}

func main() {
	config := setupConfig()

	app := &App{
		Config: config,
		SSEServer: &sse.Server{
			OnSession: nil,
		},
	}

	app.SSEServer.OnSession = onSSESession(config)

	port := ":" + config.Port
	log.Printf("Server started at %s", port)
	log.Fatal(http.ListenAndServe(port, app.setupRoutes()))
}
