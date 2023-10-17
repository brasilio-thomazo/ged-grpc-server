package repository

import (
	"context"
	"fmt"
	"io"
	"os"

	"br.dev.optimus/hermes/dao"
	"br.dev.optimus/hermes/model"
	"br.dev.optimus/hermes/pb"
	"br.dev.optimus/hermes/utils"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (r *HermesRepositoryDB) DocumentImageStore(ctx context.Context, stream pb.Hermes_DocumentImageStoreServer) error {
	imageDAO := dao.NewDocumentImageDAO(r.Reader, r.Writer)
	documentDAO := dao.NewDocumentDAO(r.Reader, r.Writer)

	in, err := stream.Recv()
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	documentID, err := uuid.Parse(in.GetInfo().DocumentId)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	if _, err := documentDAO.FindById(ctx, documentID); err != nil {
		return status.Error(codes.InvalidArgument, "document id not exists")
	}

	data := &model.DocumentImage{}
	if err := copier.Copy(data, in); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	uploadPath := fmt.Sprintf("%s/%s", os.Getenv("UPLOAD_IMAGE"), documentID)
	file, err := utils.NewFile(fmt.Sprintf("%s%s", id.String(), in.GetInfo().Extension), uploadPath)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	if err := file.Create(); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	image := &model.DocumentImage{
		DocumentID: documentID,
		Filename:   file.Filename,
		Page:       in.GetInfo().Page,
	}

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if _, err := writeData(file, in.GetContent()); err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}

	result, err := imageDAO.Store(ctx, image)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	reply := &pb.DocumentImageReply{}
	if err := copier.Copy(reply, result); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	if err := stream.SendAndClose(reply); err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

func writeData(file *utils.File, data []byte) (int, error) {
	write, err := file.File.Write(data)
	if err != nil {
		return 0, err
	}
	return write, nil
}
