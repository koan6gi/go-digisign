package app

import (
	"github.com/koan6gi/go-digisign/internal/crypto"
	"github.com/koan6gi/go-digisign/internal/network"
)

func Run() error {
	signer := crypto.NewRSASigner()
	router := network.NewRouter(signer)
	return router.Run(":8080")
}
