package server

import (
	"github.com/xlab-si/emmy/crypto/commitments"
	"github.com/xlab-si/emmy/crypto/dlog"
	pb "github.com/xlab-si/emmy/protobuf"
	"github.com/xlab-si/emmy/types"
	"math/big"
)

func (s *Server) PedersenEC(curveType dlog.Curve, stream pb.Protocol_RunServer) error {
	pedersenECReceiver := commitments.NewPedersenECReceiver(curveType)

	h := pedersenECReceiver.GetH()
	ecge := pb.ECGroupElement{
		X: h.X.Bytes(),
		Y: h.Y.Bytes(),
	}
	resp := &pb.Message{Content: &pb.Message_EcGroupElement{&ecge}}

	if err := s.send(resp, stream); err != nil {
		return err
	}

	req, err := s.receive(stream)
	if err != nil {
		return err
	}

	ecgrop := req.GetEcGroupElement()
	if ecgrop == nil {
		logger.Critical("Got a nil EC group element")
		return err
	}

	el := types.ToECGroupElement(ecgrop)
	pedersenECReceiver.SetCommitment(el)
	resp = &pb.Message{Content: &pb.Message_Empty{&pb.EmptyMsg{}}}
	if err = s.send(resp, stream); err != nil {
		return err
	}

	req, err = s.receive(stream)
	if err != nil {
		return err
	}

	pedersenDecommitment := req.GetPedersenDecommitment()
	val := new(big.Int).SetBytes(pedersenDecommitment.X)
	r := new(big.Int).SetBytes(pedersenDecommitment.R)
	valid := pedersenECReceiver.CheckDecommitment(r, val)

	logger.Noticef("Commitment scheme success: **%v**", valid)

	resp = &pb.Message{
		Content: &pb.Message_Status{&pb.Status{Success: valid}},
	}

	if err = s.send(resp, stream); err != nil {
		return err
	}

	return nil
}
