/*
 * Copyright (c) 2018. Abstrium SAS <team (at) pydio.com>
 * This file is part of Pydio Cells.
 *
 * Pydio Cells is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Pydio Cells is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Pydio Cells.  If not, see <http://www.gnu.org/licenses/>.
 *
 * The latest code can be found at <https://pydio.com>.
 */

package views

import (
	"context"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/errors"
	"github.com/pborman/uuid"
	"github.com/pydio/minio-go"
	"go.uber.org/zap"

	"github.com/pydio/cells/common"
	"github.com/pydio/cells/common/log"
	"github.com/pydio/cells/common/proto/tree"
)

type Executor struct {
	AbstractHandler
}

func (a *Executor) ExecuteWrapped(inputFilter NodeFilter, outputFilter NodeFilter, provider NodesCallback) error {

	return provider(inputFilter, outputFilter)

}

func (e *Executor) ReadNode(ctx context.Context, in *tree.ReadNodeRequest, opts ...client.CallOption) (*tree.ReadNodeResponse, error) {

	resp, err := e.clientsPool.GetTreeClient().ReadNode(ctx, in, opts...)
	if err != nil {
		if errors.Parse(err.Error()).Code != 404 {
			log.Logger(ctx).Error("Failed to read node", zap.Any("in", in), zap.Error(err))
		}
	}

	return resp, err
}

func (e *Executor) ListNodes(ctx context.Context, in *tree.ListNodesRequest, opts ...client.CallOption) (tree.NodeProvider_ListNodesClient, error) {
	log.Logger(ctx).Debug("ROUTER LISTING WITH TREE CLIENT", zap.String("path", in.Node.Path))
	return e.clientsPool.GetTreeClient().ListNodes(ctx, in, opts...)
}

func (e *Executor) CreateNode(ctx context.Context, in *tree.CreateNodeRequest, opts ...client.CallOption) (*tree.CreateNodeResponse, error) {
	node := in.Node
	if !node.IsLeaf() {
		dsPath := node.GetStringMeta(common.META_NAMESPACE_DATASOURCE_PATH)
		newNode := &tree.Node{
			Path: strings.TrimRight(node.Path, "/") + "/" + common.PYDIO_SYNC_HIDDEN_FILE_META,
		}
		newNode.SetMeta(common.META_NAMESPACE_DATASOURCE_PATH, dsPath+"/"+common.PYDIO_SYNC_HIDDEN_FILE_META)
		meta := make(map[string]string)
		if session := in.IndexationSession; session != "" {
			meta["X-Pydio-Session"] = session
		}
		if !in.UpdateIfExists {
			if read, er := e.GetObject(ctx, newNode, &GetRequestData{StartOffset: 0, Length: 36}); er == nil {
				bytes, _ := ioutil.ReadAll(read)
				read.Close()
				node.Uuid = string(bytes)
				node.MTime = time.Now().Unix()
				node.Size = 36
				log.Logger(ctx).Debug("[handlerExec.CreateNode] Hidden file already created", node.ZapUuid(), zap.Any("in", in))
				return &tree.CreateNodeResponse{Node: node}, nil
			}
		}
		// Create new Node
		nodeUuid := uuid.New()
		_, err := e.PutObject(ctx, newNode, strings.NewReader(nodeUuid), &PutRequestData{Metadata: meta, Size: int64(len(nodeUuid))})
		if err != nil {
			return nil, err
		}
		node.Uuid = nodeUuid
		node.MTime = time.Now().Unix()
		node.Size = 36
		log.Logger(ctx).Debug("[handlerExec.CreateNode] Created A Hidden .pydio for folder", node.Zap())
		return &tree.CreateNodeResponse{Node: node}, nil
	}
	log.Logger(ctx).Debug("Exec.CreateNode", zap.String("p", in.Node.Path))
	return e.clientsPool.GetTreeClientWrite().CreateNode(ctx, in, opts...)
}

func (e *Executor) UpdateNode(ctx context.Context, in *tree.UpdateNodeRequest, opts ...client.CallOption) (*tree.UpdateNodeResponse, error) {
	return e.clientsPool.GetTreeClientWrite().UpdateNode(ctx, in, opts...)
}

func (e *Executor) DeleteNode(ctx context.Context, in *tree.DeleteNodeRequest, opts ...client.CallOption) (*tree.DeleteNodeResponse, error) {
	info, ok := GetBranchInfo(ctx, "in")
	if !ok {
		return nil, errors.BadRequest(VIEWS_LIBRARY_NAME, "Cannot find S3 client, did you insert a resolver middleware?")
	}
	writer := info.Client
	if meta, mOk := MinioMetaFromContext(ctx); mOk {
		if session := in.IndexationSession; session != "" {
			meta["X-Pydio-Session"] = session
		}
		writer.PrepareMetadata(meta)
		defer writer.ClearMetadata()
	}
	log.Logger(ctx).Debug("Exec.DeleteNode", in.Node.Zap(), zap.Any("bucket", info.ObjectsBucket))

	s3Path := e.buildS3Path(info, in.Node)
	err := writer.RemoveObject(info.ObjectsBucket, s3Path)
	success := true
	if err != nil {
		log.Logger(ctx).Error("Error while deleting node", zap.Error(err))
		success = false
	}
	return &tree.DeleteNodeResponse{Success: success}, err
}

func (e *Executor) GetObject(ctx context.Context, node *tree.Node, requestData *GetRequestData) (io.ReadCloser, error) {
	info, ok := GetBranchInfo(ctx, "in")
	if !ok {
		return nil, errors.BadRequest(VIEWS_LIBRARY_NAME, "Cannot find S3 client, did you insert a resolver middleware?")
	}
	writer := info.Client
	if meta, mOk := MinioMetaFromContext(ctx); mOk {
		writer.PrepareMetadata(meta)
		defer writer.ClearMetadata()
	} else {
		log.Logger(ctx).Debug("Preparing Meta for GetObject: No Meta Found")
	}

	var reader io.ReadCloser
	var err error

	s3Path := e.buildS3Path(info, node)
	headers := minio.GetObjectOptions{}

	if requestData.EncryptionMaterial != nil {

		var offset = requestData.StartOffset
		var end = offset + requestData.Length - 1

		if requestData.StartOffset >= 0 && requestData.Length >= 0 {
			log.Logger(ctx).Debug("GET RANGE", zap.Int64("From", requestData.StartOffset), zap.Int64("Length", requestData.Length), node.Zap())
			if end >= node.Size-1 {
				end = 0
			}
		}

		headers.Materials = requestData.EncryptionMaterial
		if offset == 0 && end == 0 {
			log.Logger(ctx).Debug("GET DATA WITH NO RANGE ")
			reader, err = writer.GetEncryptedObject(info.ObjectsBucket, s3Path, requestData.EncryptionMaterial)
		} else {
			log.Logger(ctx).Info("Warning, passing a request Length on encrypted data is not supported yet", zap.Int64("offset", requestData.StartOffset), zap.Int64("end", end))
			if err := headers.SetRange(requestData.StartOffset, end); err != nil {
				return nil, err
			}
			reader, _, err = writer.GetObject(info.ObjectsBucket, s3Path, headers)
		}
	} else {
		headers := minio.GetObjectOptions{}
		if requestData.StartOffset >= 0 && requestData.Length >= 0 {
			if err := headers.SetRange(requestData.StartOffset, requestData.StartOffset+requestData.Length-1); err != nil {
				return nil, err
			}
		}
		log.Logger(ctx).Debug("Get Object", zap.String("bucket", info.ObjectsBucket), zap.String("s3path", s3Path), zap.Any("headers", headers), zap.Any("request", requestData))
		reader, _, err = writer.GetObject(info.ObjectsBucket, s3Path, headers)
		if err != nil {
			//log.Logger(ctx).Error("Get Object", zap.Error(err))
		}
	}
	return reader, err
}

func (e *Executor) PutObject(ctx context.Context, node *tree.Node, reader io.Reader, requestData *PutRequestData) (int64, error) {
	info, ok := GetBranchInfo(ctx, "in")
	if !ok {
		return 0, errors.BadRequest(VIEWS_LIBRARY_NAME, "Cannot find S3 client, did you insert a resolver middleware?")
	}
	writer := info.Client
	if meta, mOk := MinioMetaFromContext(ctx); mOk {
		if requestMeta := requestData.Metadata; len(requestMeta) > 0 {
			for key, val := range requestMeta {
				meta[key] = val
			}
		}
		writer.PrepareMetadata(meta)
		defer writer.ClearMetadata()
	}

	s3Path := e.buildS3Path(info, node)

	if requestData.EncryptionMaterial != nil {
		return writer.PutObjectWithContext(context.Background(), info.ObjectsBucket, s3Path, reader, -1, minio.PutObjectOptions{EncryptMaterials: requestData.EncryptionMaterial, UserMetadata: requestData.Metadata})
		//return writer.PutEncryptedObject(info.ObjectsBucket, s3Path, reader, requestData.EncryptionMaterial)

	} else {
		log.Logger(ctx).Debug("handler exec: put object", zap.Any("info", info), zap.String("s3Path", s3Path), zap.Any("requestData", requestData))
		if requestData.Size <= 0 {
			written, err := writer.PutObjectWithContext(ctx, info.ObjectsBucket, s3Path, reader, -1, minio.PutObjectOptions{UserMetadata: requestData.Metadata})
			if err != nil {
				return 0, err
			} else {
				return written, nil
			}
		} else {
			oi, err := writer.PutObject(info.ObjectsBucket, s3Path, reader, requestData.Size, requestData.Md5Sum, requestData.Sha256Sum, requestData.Metadata)
			if err != nil {
				return 0, err
			} else {
				return oi.Size, nil
			}
		}
	}
}

func (e *Executor) CopyObject(ctx context.Context, from *tree.Node, to *tree.Node, requestData *CopyRequestData) (int64, error) {

	// If DS's are same datasource, simple S3 Copy operation. Otherwise it must copy from one to another.
	destInfo, ok := GetBranchInfo(ctx, "to")
	srcInfo, ok2 := GetBranchInfo(ctx, "from")
	if !ok || !ok2 {
		return 0, errors.InternalServerError(VIEWS_LIBRARY_NAME, "Cannot find Client for src or dest")
	}
	destClient := destInfo.Client
	srcClient := srcInfo.Client
	destBucket := destInfo.ObjectsBucket
	srcBucket := srcInfo.ObjectsBucket
	if meta, mOk := MinioMetaFromContext(ctx); mOk {
		destClient.PrepareMetadata(meta)
		srcClient.PrepareMetadata(meta)
		defer srcClient.ClearMetadata()
		defer destClient.ClearMetadata()
	}

	// var srcSse, destSse minio.SSEInfo
	// if requestData.srcEncryptionMaterial != nil {
	// 	srcSse = minio.NewSSEInfo([]byte(requestData.srcEncryptionMaterial.GetDecrypted()), "")
	// }
	// if requestData.destEncryptionMaterial != nil {
	// 	destSse = minio.NewSSEInfo([]byte(requestData.destEncryptionMaterial.GetDecrypted()), "")
	// }

	fromPath := e.buildS3Path(srcInfo, from)
	toPath := e.buildS3Path(destInfo, to)

	if destClient == srcClient && requestData.SrcVersionId == "" {

		// srcInfo := minio.NewSourceInfo(srcBucket, fromPath, &srcSse)
		// destInfo, err := minio.NewDestinationInfo(destBucket, toPath, &destSse, requestData.Metadata)
		// if err != nil {
		// 	return 0, err
		// }

		oi, err := destClient.CopyObject(srcBucket, fromPath, destBucket, toPath, requestData.Metadata)
		if err != nil {
			return 0, err
		}
		oi, err3 := destClient.StatObject(destBucket, toPath, minio.StatObjectOptions{})
		if err3 != nil {
			return 0, err3
		}
		return oi.Size, nil

	} else {

		var reader io.ReadCloser
		var err error
		srcStat, srcErr := srcClient.StatObject(srcBucket, fromPath, minio.StatObjectOptions{})
		if srcErr != nil {
			return 0, srcErr
		}
		if requestData.srcEncryptionMaterial != nil {
			reader, err = srcClient.GetEncryptedObject(srcBucket, fromPath, requestData.srcEncryptionMaterial)
		} else {
			headers := minio.GetObjectOptions{}
			reader, _, err = srcClient.GetObject(srcBucket, fromPath, headers)
		}
		if err != nil {
			return 0, err
		}

		if requestData.destEncryptionMaterial != nil {
			return destClient.PutEncryptedObject(destBucket, toPath, reader, requestData.destEncryptionMaterial)
		} else {
			oi, err := destClient.PutObject(destBucket, toPath, reader, srcStat.Size, nil, nil, requestData.Metadata)
			if err != nil {
				log.Logger(ctx).Error("CopyObject / Different Clients",
					zap.Error(err),
					zap.Any("srcStat", srcStat),
					zap.Any("srcInfo", srcInfo),
					zap.Any("destInfo", destInfo),
					zap.Any("to", toPath))
			}
			return oi.Size, err
		}

	}

}

func (e *Executor) MultipartCreate(ctx context.Context, target *tree.Node, requestData *MultipartRequestData) (string, error) {
	info, ok := GetBranchInfo(ctx, "in")
	if !ok {
		return "", errors.InternalServerError(VIEWS_LIBRARY_NAME, "Cannot find client")
	}
	if meta, mOk := MinioMetaFromContext(ctx); mOk {
		info.Client.PrepareMetadata(meta)
		defer info.Client.ClearMetadata()
	}
	s3Path := e.buildS3Path(info, target)

	putOptions := minio.PutObjectOptions{}
	putOptions.UserMetadata = requestData.Metadata
	id, err := info.Client.NewMultipartUpload(info.ObjectsBucket, s3Path, putOptions)
	return id, err
}

func (e *Executor) MultipartPutObjectPart(ctx context.Context, target *tree.Node, uploadID string, partNumberMarker int, reader io.Reader, requestData *PutRequestData) (minio.ObjectPart, error) {
	info, ok := GetBranchInfo(ctx, "in")
	if !ok {
		return minio.ObjectPart{PartNumber: partNumberMarker}, errors.BadRequest(VIEWS_LIBRARY_NAME, "Cannot find S3 client, did you insert a resolver middleware?")
	}
	writer := info.Client
	if meta, mOk := MinioMetaFromContext(ctx); mOk {
		writer.PrepareMetadata(meta)
		defer writer.ClearMetadata()
	}
	s3Path := e.buildS3Path(info, target)

	if requestData.EncryptionMaterial != nil {
		return minio.ObjectPart{PartNumber: partNumberMarker},
			errors.BadRequest(VIEWS_LIBRARY_NAME, "Multipart encrypted upload is not implemented")
	} else {
		log.Logger(ctx).Debug("HANDLER-EXEC: before put", zap.Any("requestData", requestData))

		if requestData.Size <= 0 {
			// This should never happen, double check
			return minio.ObjectPart{PartNumber: partNumberMarker},
				errors.BadRequest(VIEWS_LIBRARY_NAME, "trying to upload a part object that has no data. Double check")
		} else {
			cp, err := writer.PutObjectPartWithMetadata(info.ObjectsBucket, s3Path, uploadID, partNumberMarker, reader, requestData.Size, requestData.Md5Sum, requestData.Sha256Sum, requestData.Metadata)
			if err != nil {
				log.Logger(ctx).Error("PutObjectPartWithMetadata has failed", zap.Error(err))
				return minio.ObjectPart{PartNumber: partNumberMarker}, err
			} else {
				return cp, nil
			}
		}
	}
}

func (e *Executor) MultipartList(ctx context.Context, prefix string, requestData *MultipartRequestData) (res minio.ListMultipartUploadsResult, err error) {
	info, ok := GetBranchInfo(ctx, "in")
	if !ok {
		return res, errors.InternalServerError(VIEWS_LIBRARY_NAME, "Cannot find client")
	}
	if meta, mOk := MinioMetaFromContext(ctx); mOk {
		info.Client.PrepareMetadata(meta)
		defer info.Client.ClearMetadata()
	}
	return info.Client.ListMultipartUploads(info.ObjectsBucket, prefix, requestData.ListKeyMarker, requestData.ListUploadIDMarker, requestData.ListDelimiter, requestData.ListMaxUploads)
}

func (e *Executor) MultipartAbort(ctx context.Context, target *tree.Node, uploadID string, requestData *MultipartRequestData) error {
	info, ok := GetBranchInfo(ctx, "in")
	if !ok {
		return errors.InternalServerError(VIEWS_LIBRARY_NAME, "Cannot find client")
	}
	if meta, mOk := MinioMetaFromContext(ctx); mOk {
		info.Client.PrepareMetadata(meta)
		defer info.Client.ClearMetadata()
	}
	s3Path := e.buildS3Path(info, target)
	return info.Client.AbortMultipartUpload(info.ObjectsBucket, s3Path, uploadID)
}

func (e *Executor) MultipartComplete(ctx context.Context, target *tree.Node, uploadID string, uploadedParts []minio.CompletePart) (minio.ObjectInfo, error) {
	info, ok := GetBranchInfo(ctx, "in")
	if !ok {
		return minio.ObjectInfo{}, errors.InternalServerError(VIEWS_LIBRARY_NAME, "Cannot find client")
	}
	if meta, mOk := MinioMetaFromContext(ctx); mOk {
		info.Client.PrepareMetadata(meta)
		defer info.Client.ClearMetadata()
	}
	s3Path := e.buildS3Path(info, target)

	log.Logger(ctx).Debug("HANDLER-EXEC - before calling minio.CompleteMultipartUpload", zap.Any("Parts", uploadedParts))
	err := info.Client.CompleteMultipartUpload(info.ObjectsBucket, s3Path, uploadID, uploadedParts)
	if err != nil {
		log.Logger(ctx).Error("fail to complete upload", zap.Error(err))
		return minio.ObjectInfo{}, err
	}
	return info.Client.StatObject(info.ObjectsBucket, target.GetStringMeta(common.META_NAMESPACE_DATASOURCE_PATH), minio.StatObjectOptions{})
}

func (e *Executor) MultipartListObjectParts(ctx context.Context, target *tree.Node, uploadID string, partNumberMarker int, maxParts int) (lpi minio.ListObjectPartsResult, err error) {
	info, ok := GetBranchInfo(ctx, "in")
	if !ok {
		return lpi, errors.InternalServerError(VIEWS_LIBRARY_NAME, "Cannot find client")
	}
	if meta, mOk := MinioMetaFromContext(ctx); mOk {
		info.Client.PrepareMetadata(meta)
		defer info.Client.ClearMetadata()
	}
	s3Path := e.buildS3Path(info, target)
	return info.Client.ListObjectParts(info.ObjectsBucket, s3Path, uploadID, partNumberMarker, maxParts)
}

func (e *Executor) buildS3Path(branchInfo BranchInfo, node *tree.Node) string {

	path := node.GetStringMeta(common.META_NAMESPACE_DATASOURCE_PATH)
	if branchInfo.ObjectsBaseFolder != "" {
		path = strings.TrimLeft(branchInfo.ObjectsBaseFolder, "/") + path
	}
	return path

}
