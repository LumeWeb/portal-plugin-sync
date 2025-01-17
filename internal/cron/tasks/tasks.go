package tasks

import (
	"bytes"
	"encoding/hex"
	"errors"
	"go.lumeweb.com/portal-plugin-sync/internal/cron/define"
	"go.lumeweb.com/portal-plugin-sync/internal/metadata"
	"go.lumeweb.com/portal-plugin-sync/types"
	"go.lumeweb.com/portal/bao"
	"go.lumeweb.com/portal/core"
	_event "go.lumeweb.com/portal/event"
	"go.sia.tech/renterd/api"
	"go.sia.tech/renterd/object"
	"go.uber.org/zap"
	"io"
)

const syncBucketName = "sync"

func getSyncProtocol(protocol string) (types.SyncProtocol, error) {
	proto := core.GetProtocol(protocol)

	if proto == nil {
		return nil, errors.New("protocol not found")
	}

	syncProto, ok := proto.(types.SyncProtocol)

	if !ok {
		return nil, errors.New("protocol is not a Sync protocol")
	}

	return syncProto, nil
}

func encodeProtocolFileName(hash []byte, protocol string) (string, error) {
	syncProto, err := getSyncProtocol(protocol)
	if err != nil {
		return "", err
	}

	return syncProto.EncodeFileName(hash), nil
}

func CronTaskVerifyObject(input any, ctx core.Context) error {
	args, ok := input.(*define.CronTaskVerifyObjectArgs)
	if !ok {
		return errors.New("invalid arguments type")
	}
	logger := ctx.Logger()
	renter := ctx.Service(core.RENTER_SERVICE).(core.RenterService)
	cron := ctx.Service(core.CRON_SERVICE).(core.CronService)
	err := renter.CreateBucketIfNotExists(syncBucketName)
	if err != nil {
		return err
	}

	success := false

	var foundObject metadata.FileMeta

	for _, object_ := range args.Object {
		if !bytes.Equal(object_.Hash, args.Hash) {
			logger.Error("hash mismatch", zap.Binary("expected", args.Hash), zap.Binary("actual", object_.Hash))
			continue
		}

		fileName, err := encodeProtocolFileName(object_.Hash, object_.Protocol)
		if err != nil {
			logger.Error("failed to encode protocol file name", zap.Error(err))
			return err
		}

		err = renter.ImportObjectMetadata(ctx, syncBucketName, fileName, object.Object{
			Key:   object_.Key,
			Slabs: object_.Slabs,
		})

		if err != nil {
			logger.Error("failed to import object metadata", zap.Error(err))
			continue
		}

		objectRet, err := renter.GetObject(ctx, syncBucketName, fileName, api.DownloadObjectOptions{})
		if err != nil {
			return err
		}

		verifier := bao.NewVerifier(objectRet.Content, bao.Result{
			Hash:   object_.Hash,
			Proof:  object_.Proof,
			Length: uint(object_.Size),
		}, logger.Logger)

		_, err = io.Copy(io.Discard, verifier)
		if err != nil {
			logger.Error("failed to verify object", zap.Error(err))
			continue
		}

		success = true
		foundObject = object_
	}

	if success {
		err := cron.CreateJobIfNotExists(define.CronTaskUploadObjectName, define.CronTaskUploadObjectArgs{
			Hash:       args.Hash,
			Protocol:   foundObject.Protocol,
			Size:       foundObject.Size,
			UploaderID: args.UploaderID,
		}, []string{hex.EncodeToString(args.Hash)})
		if err != nil {
			return err
		}
	}

	return nil
}

type seekableSiaStream struct {
	rc    io.ReadCloser
	ctx   core.Context
	args  *define.CronTaskUploadObjectArgs
	pos   int64
	reset bool
	size  int64
}

func (r *seekableSiaStream) Read(p []byte) (n int, err error) {
	if r.reset {
		r.reset = false
		err := r.rc.Close()
		if err != nil {
			return 0, err
		}

		fileName, err := encodeProtocolFileName(r.args.Hash, r.args.Protocol)
		if err != nil {
			r.ctx.Logger().Error("failed to encode protocol file name", zap.Error(err))
			return 0, err
		}

		objectRet, err := r.ctx.Service(core.RENTER_SERVICE).(core.RenterService).GetObject(r.ctx, syncBucketName, fileName, api.DownloadObjectOptions{})
		if err != nil {
			return 0, err
		}
		r.rc = objectRet.Content
		r.pos = 0
	}
	n, err = r.rc.Read(p)
	r.pos += int64(n)
	return n, err
}

func (r *seekableSiaStream) Seek(offset int64, whence int) (int64, error) {
	if offset == 0 && whence == io.SeekStart {
		r.reset = true
		return 0, nil
	}

	if offset == 0 && whence == io.SeekEnd {
		return r.size, nil
	}

	return 0, errors.New("seek not supported")
}

func (r *seekableSiaStream) Close() error {
	return r.rc.Close()
}

func CronTaskUploadObject(input any, ctx core.Context) error {
	args, ok := input.(*define.CronTaskUploadObjectArgs)
	if !ok {
		return errors.New("invalid arguments type")
	}

	logger := ctx.Logger()
	renter := ctx.Service(core.RENTER_SERVICE).(core.RenterService)
	storage := ctx.Service(core.STORAGE_SERVICE).(core.StorageService)
	meta := ctx.Service(core.METADATA_SERVICE).(core.MetadataService)
	fileName, err := encodeProtocolFileName(args.Hash, args.Protocol)
	if err != nil {
		logger.Error("failed to encode protocol file name", zap.Error(err))
		return err
	}

	objectRet, err := renter.GetObject(ctx, syncBucketName, fileName, api.DownloadObjectOptions{})
	if err != nil {
		return err
	}

	syncProtocol, err := getSyncProtocol(args.Protocol)
	if err != nil {
		logger.Error("failed to get Sync protocol", zap.Error(err))
		return err
	}

	storeProtocol := syncProtocol.StorageProtocol()

	wrapper := &seekableSiaStream{
		rc:   objectRet.Content,
		ctx:  ctx,
		args: args,
		size: objectRet.Size,
	}

	upload, err := storage.UploadObject(ctx, storeProtocol, wrapper, args.Size, nil, nil)

	if err != nil {
		return err
	}

	upload.UserID = uint(args.UploaderID)

	err = meta.SaveUpload(ctx, *upload, true)
	if err != nil {
		return err
	}

	err = renter.DeleteObjectMetadata(ctx, syncBucketName, fileName)
	if err != nil {
		return err
	}

	err = _event.FireStorageObjectUploadedEvent(ctx, upload)
	if err != nil {
		return err
	}

	return nil
}

func CronTaskScanObjects(_ any, ctx core.Context) error {
	logger := ctx.Logger()
	meta := ctx.Service(core.METADATA_SERVICE).(core.MetadataService)
	_sync := ctx.Service(types.SYNC_SERVICE).(types.SyncService)
	uploads, err := meta.GetAllUploads(ctx)
	if err != nil {
		return err
	}

	for _, upload := range uploads {
		err := _sync.Update(upload)
		if err != nil {
			logger.Error("failed to update upload", zap.Error(err))
		}
	}

	return nil
}
