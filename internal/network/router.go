package network

import (
	"github.com/gin-gonic/gin"
	"github.com/koan6gi/go-digisign/internal/crypto"
)

type Router struct {
	engine *gin.Engine
	signer crypto.Signer
}

func NewRouter(signer crypto.Signer) *Router {
	r := &Router{
		engine: gin.Default(),
		signer: signer,
	}

	r.setupRoutes()
	return r
}

func (r *Router) setupRoutes() {
	r.engine.Static("/static", "./static")
	r.engine.StaticFile("/", "./static/index.html")
	r.engine.StaticFile("/generate", "./static/generate.html")
	r.engine.StaticFile("/sign", "./static/sign.html")
	r.engine.StaticFile("/verify", "./static/verify.html")

	r.engine.POST("/api/generate", r.generateHandler)
	r.engine.POST("/api/sign", r.signHandler)
	r.engine.POST("/api/verify", r.verifyHandler)
	r.engine.GET("/download/cert", r.downloadCertHandler)
	r.engine.GET("/download/key", r.downloadKeyHandler)
}

func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
