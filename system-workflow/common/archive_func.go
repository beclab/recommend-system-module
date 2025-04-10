package common

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"io"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

func DoZlibCompress(src []byte) []byte {
	var in bytes.Buffer
	w := zlib.NewWriter(&in)
	w.Write(src)
	w.Close()
	return in.Bytes()
}

func DoZlibUnCompress(compressSrc []byte) []byte {
	b := bytes.NewReader(compressSrc)
	var out bytes.Buffer
	r, _ := zlib.NewReader(b)
	io.Copy(&out, r)
	return out.Bytes()
}

func UncompressTarGz(dstDirectory, srcTarGzPath string) (dstFiles []string, err error) {
	fileSrcTarGz, openFileSrcTarGzErr := os.Open(srcTarGzPath)

	if openFileSrcTarGzErr != nil {
		Logger.Error("open file error", zap.String("srcTarGz", srcTarGzPath), zap.Error(openFileSrcTarGzErr))
		return
	}
	defer fileSrcTarGz.Close()

	gzipReader, newReaderErr := gzip.NewReader(fileSrcTarGz)
	if newReaderErr != nil {
		Logger.Error("new gzip reader error", zap.Error(newReaderErr))
		return
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	uncompressDstFiles := make([]string, 32)
	for {
		tarHeader, tarReadNexErr := tarReader.Next()
		if tarReadNexErr == io.EOF {
			break
		}

		switch {
		case tarReadNexErr != nil:
			return nil, err
		case tarHeader == nil:
			continue
		}

		dstFileDir := filepath.Join(dstDirectory, tarHeader.Name)
		Logger.Info("three related uncompress info", zap.String("dstDirectory", dstDirectory), zap.String("tarHeader.Name", tarHeader.Name), zap.String("dstFileDir", dstFileDir))
		switch tarHeader.Typeflag {
		case tar.TypeDir:
			if b := ExistDir(dstFileDir); !b {
				if createDirErr := os.MkdirAll(dstFileDir, 0775); createDirErr != nil {
					Logger.Error("create dir error", zap.String("createdDir", dstFileDir), zap.Error(createDirErr))
					return nil, err
				}
			}
		case tar.TypeReg:
			uncompressArchiveSingleFileErr := uncompressArchiveSingleFile(dstFileDir, tarHeader, tarReader)
			if uncompressArchiveSingleFileErr != nil {
				Logger.Error("uncompress single file err", zap.String("dstFileDir", dstFileDir), zap.Error(uncompressArchiveSingleFileErr))
				return nil, uncompressArchiveSingleFileErr
			}
			uncompressDstFiles = append(uncompressDstFiles, dstFileDir)
		}
	}

	return uncompressDstFiles, nil
}

func uncompressArchiveSingleFile(dstFileDir string, tarHeader *tar.Header, tarReader *tar.Reader) error {
	openFile, openFileErr := os.OpenFile(dstFileDir, os.O_CREATE|os.O_RDWR, os.FileMode(tarHeader.Mode))
	if openFileErr != nil {
		Logger.Error("open file error", zap.String("dstFileDir", dstFileDir), zap.Error(openFileErr))
		return openFileErr
	}
	defer openFile.Close()
	copyFileSize, copyErr := io.Copy(openFile, tarReader)
	if copyErr != nil {
		Logger.Error("copy file err", zap.String("dstFileDir", dstFileDir), zap.Error(copyErr))
		return copyErr
	}
	Logger.Info("copy file size", zap.String("copyedFile", dstFileDir), zap.Int64("size", copyFileSize))

	return nil
}
