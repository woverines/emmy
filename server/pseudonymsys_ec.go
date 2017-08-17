package server

import (
	"github.com/xlab-si/emmy/config"
	"github.com/xlab-si/emmy/crypto/pseudonymsys"
	pb "github.com/xlab-si/emmy/protobuf"
	"github.com/xlab-si/emmy/types"
	"math/big"
)

func (s *Server) PseudonymsysGenerateNymEC(req *pb.Message, stream pb.Protocol_RunServer) error {
	org := pseudonymsys.NewOrgNymGenEC()

	proofRandData := req.GetPseudonymsysNymGenProofRandomDataEc()
	x1 := types.ToECGroupElement(proofRandData.X1)
	nymA := types.ToECGroupElement(proofRandData.A1)
	nymB := types.ToECGroupElement(proofRandData.B1)
	x2 := types.ToECGroupElement(proofRandData.X2)
	blindedA := types.ToECGroupElement(proofRandData.A2)
	blindedB := types.ToECGroupElement(proofRandData.B2)
	signatureR := new(big.Int).SetBytes(proofRandData.R)
	signatureS := new(big.Int).SetBytes(proofRandData.S)

	challenge, err := org.GetChallenge(nymA, blindedA, nymB, blindedB, x1, x2, signatureR, signatureS)
	var resp *pb.Message
	if err != nil {
		resp = &pb.Message{
			Content: &pb.Message_PedersenDecommitment{
				&pb.PedersenDecommitment{},
			},
			ProtocolError: err.Error(),
		}
	} else {
		resp = &pb.Message{
			Content: &pb.Message_PedersenDecommitment{
				&pb.PedersenDecommitment{
					X: challenge.Bytes(),
				},
			},
		}
	}

	if err := s.send(resp, stream); err != nil {
		return err
	}

	req, err = s.receive(stream)
	if err != nil {
		return err
	}

	proofData := req.GetSchnorrProofData() // SchnorrProofData is used in DLog equality proof as well
	z := new(big.Int).SetBytes(proofData.Z)
	valid := org.Verify(z)

	resp = &pb.Message{
		Content: &pb.Message_Status{&pb.Status{Success: valid}},
	}

	if err = s.send(resp, stream); err != nil {
		return err
	}

	return nil
}

func (s *Server) PseudonymsysIssueCredentialEC(req *pb.Message, stream pb.Protocol_RunServer) error {
	proofRandData := req.GetSchnorrEcProofRandomData()
	x := types.ToECGroupElement(proofRandData.X)
	a := types.ToECGroupElement(proofRandData.A)
	b := types.ToECGroupElement(proofRandData.B)

	org := pseudonymsys.NewOrgCredentialIssuerEC()
	challenge := org.GetAuthenticationChallenge(a, b, x)

	resp := &pb.Message{
		Content: &pb.Message_Bigint{
			&pb.BigInt{
				X1: challenge.Bytes(),
			},
		},
	}

	if err := s.send(resp, stream); err != nil {
		return err
	}

	req, err := s.receive(stream)
	if err != nil {
		return err
	}

	proofData := req.GetBigint()
	z := new(big.Int).SetBytes(proofData.X1)

	x11, x12, x21, x22, A, B, err := org.VerifyAuthentication(z)

	if err != nil {
		resp = &pb.Message{
			Content: &pb.Message_PseudonymsysIssueProofRandomDataEc{
				&pb.PseudonymsysIssueProofRandomDataEC{},
			},
			ProtocolError: err.Error(),
		}
	} else {
		resp = &pb.Message{
			Content: &pb.Message_PseudonymsysIssueProofRandomDataEc{
				&pb.PseudonymsysIssueProofRandomDataEC{
					X11: types.ToPbECGroupElement(x11),
					X12: types.ToPbECGroupElement(x12),
					X21: types.ToPbECGroupElement(x21),
					X22: types.ToPbECGroupElement(x22),
					A:   types.ToPbECGroupElement(A),
					B:   types.ToPbECGroupElement(B),
				},
			},
		}
	}

	if err := s.send(resp, stream); err != nil {
		return err
	}

	req, err = s.receive(stream)
	if err != nil {
		return err
	}

	challenges := req.GetDoubleBigint()
	challenge1 := new(big.Int).SetBytes(challenges.X1)
	challenge2 := new(big.Int).SetBytes(challenges.X2)

	z1, z2 := org.GetEqualityProofData(challenge1, challenge2)
	resp = &pb.Message{
		Content: &pb.Message_DoubleBigint{
			&pb.DoubleBigInt{
				X1: z1.Bytes(),
				X2: z2.Bytes(),
			},
		},
	}

	if err := s.send(resp, stream); err != nil {
		return err
	}

	return nil
}

func (s *Server) PseudonymsysTransferCredentialEC(req *pb.Message, stream pb.Protocol_RunServer) error {
	org := pseudonymsys.NewOrgCredentialVerifierEC()
	data := req.GetPseudonymsysTransferCredentialDataEc()
	orgName := data.OrgName
	x1 := types.ToECGroupElement(data.X1)
	x2 := types.ToECGroupElement(data.X2)
	nymA := types.ToECGroupElement(data.NymA)
	nymB := types.ToECGroupElement(data.NymB)

	t1 := make([]*big.Int, 6)
	t1[0] = new(big.Int).SetBytes(data.Credential.T1.A.X)
	t1[1] = new(big.Int).SetBytes(data.Credential.T1.A.Y)
	t1[2] = new(big.Int).SetBytes(data.Credential.T1.B.X)
	t1[3] = new(big.Int).SetBytes(data.Credential.T1.B.Y)
	t1[4] = new(big.Int).SetBytes(data.Credential.T1.Hash)
	t1[5] = new(big.Int).SetBytes(data.Credential.T1.ZAlpha)

	t2 := make([]*big.Int, 6)
	t2[0] = new(big.Int).SetBytes(data.Credential.T2.A.X)
	t2[1] = new(big.Int).SetBytes(data.Credential.T2.A.Y)
	t2[2] = new(big.Int).SetBytes(data.Credential.T2.B.X)
	t2[3] = new(big.Int).SetBytes(data.Credential.T2.B.Y)
	t2[4] = new(big.Int).SetBytes(data.Credential.T2.Hash)
	t2[5] = new(big.Int).SetBytes(data.Credential.T2.ZAlpha)

	credential := pseudonymsys.NewCredentialEC(
		types.ToECGroupElement(data.Credential.SmallAToGamma),
		types.ToECGroupElement(data.Credential.SmallBToGamma),
		types.ToECGroupElement(data.Credential.AToGamma),
		types.ToECGroupElement(data.Credential.BToGamma),
		t1, t2,
	)

	challenge := org.GetAuthenticationChallenge(nymA, nymB,
		credential.SmallAToGamma, credential.SmallBToGamma, x1, x2)

	resp := &pb.Message{
		Content: &pb.Message_Bigint{
			&pb.BigInt{
				X1: challenge.Bytes(),
			},
		},
	}

	if err := s.send(resp, stream); err != nil {
		return err
	}

	req, err := s.receive(stream)
	if err != nil {
		return err
	}

	// PubKeys of the organization that issue a credential:
	h1X, h1Y, h2X, h2Y := config.LoadPseudonymsysOrgPubKeysEC(orgName)
	h1 := types.NewECGroupElement(h1X, h1Y)
	h2 := types.NewECGroupElement(h2X, h2Y)
	orgPubKeys := pseudonymsys.NewOrgPubKeysEC(h1, h2)

	proofData := req.GetBigint()
	z := new(big.Int).SetBytes(proofData.X1)

	verified := org.VerifyAuthentication(z, credential, orgPubKeys)

	resp = &pb.Message{
		Content: &pb.Message_Status{&pb.Status{Success: verified}},
	}

	if err = s.send(resp, stream); err != nil {
		return err
	}

	return nil
}
