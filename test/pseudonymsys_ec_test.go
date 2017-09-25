package test

import (
	"github.com/stretchr/testify/assert"
	"github.com/xlab-si/emmy/client"
	"github.com/xlab-si/emmy/config"
	"github.com/xlab-si/emmy/crypto/dlog"
	"github.com/xlab-si/emmy/crypto/pseudonymsys"
	"github.com/xlab-si/emmy/types"
	"testing"
)

func TestPseudonymsysEC(t *testing.T) {
	curveType := dlog.P256
	ecdlog := dlog.NewECDLog(curveType)
	caClient, err := client.NewPseudonymsysCAClientEC(testGrpcClientConn, curveType)
	if err != nil {
		t.Errorf("Error when initializing NewPseudonymsysCAClientEC")
	}

	userSecret := config.LoadPseudonymsysUserSecret("user1", "ecdlog")

	nymA := types.NewECGroupElement(ecdlog.Curve.Params().Gx, ecdlog.Curve.Params().Gy)
	nymB1, nymB2 := ecdlog.Exponentiate(nymA.X, nymA.Y, userSecret) // this is user's public key
	nymB := types.NewECGroupElement(nymB1, nymB2)

	masterNym := pseudonymsys.NewPseudonymEC(nymA, nymB)
	caCertificate, err := caClient.ObtainCertificate(userSecret, masterNym)
	if err != nil {
		t.Errorf("Error when registering with CA")
	}

	// usually the endpoint is different from the one used for CA:
	c1, err := client.NewPseudonymsysClientEC(testGrpcClientConn, curveType)
	nym1, err := c1.GenerateNym(userSecret, caCertificate)
	if err != nil {
		t.Errorf(err.Error())
	}

	orgName := "org1"
	h1X, h1Y, h2X, h2Y := config.LoadPseudonymsysOrgPubKeysEC(orgName)
	h1 := types.NewECGroupElement(h1X, h1Y)
	h2 := types.NewECGroupElement(h2X, h2Y)
	orgPubKeys := pseudonymsys.NewOrgPubKeysEC(h1, h2)
	credential, err := c1.ObtainCredential(userSecret, nym1, orgPubKeys)
	if err != nil {
		t.Errorf(err.Error())
	}

	// register with org2
	// create a client to communicate with org2
	caClient1, err := client.NewPseudonymsysCAClientEC(testGrpcClientConn, curveType)
	caCertificate1, err := caClient1.ObtainCertificate(userSecret, masterNym)
	if err != nil {
		t.Errorf("Error when registering with CA")
	}

	c2, err := client.NewPseudonymsysClientEC(testGrpcClientConn, curveType)
	nym2, err := c2.GenerateNym(userSecret, caCertificate1)
	if err != nil {
		t.Errorf(err.Error())
	}

	authenticated, err := c2.TransferCredential(orgName, userSecret, nym2, credential)
	if err != nil {
		t.Errorf(err.Error())
	}

	assert.Equal(t, authenticated, true, "Pseudonymsys test failed")
}
