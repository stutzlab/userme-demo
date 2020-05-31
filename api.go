package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	jwtparser "github.com/flaviostutz/gin-jwt-parser"
	cors "github.com/itsjamie/gin-cors"
)

type HTTPServer struct {
	server *http.Server
	router *gin.Engine
}

func NewHTTPServer() *HTTPServer {
	router := gin.Default()

	router.Use(cors.Middleware(cors.Config{
		Origins:         opt.corsAllowedOrigins,
		Methods:         "GET, POST",
		RequestHeaders:  "Origin, Content-Type",
		ExposedHeaders:  "",
		MaxAge:          24 * 3600 * time.Second,
		Credentials:     false,
		ValidateHeaders: false,
	}))

	router.Use(jwtparser.Middleware(jwtparser.Config{
		FromBearer:       "Authorization",
		FromCookie:       "jwt",
		FromQuery:        "t",
		JWTSigningMethod: "ES256",
		JWTVerifyKeyFile: opt.jwtSigningKeyFile,
	}))

	h := &HTTPServer{server: &http.Server{
		Addr:    ":2000",
		Handler: router,
	}, router: router}

	logrus.Infof("Initializing HTTP Handlers...")
	h.setupUserTODO()

	return h
}

//Start the main HTTP Server entry
func (s *HTTPServer) Start() error {
	logrus.Infof("Starting HTTP Server on port 2000")
	return s.server.ListenAndServe()
}
