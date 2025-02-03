package tar

import (
	"archive/tar"
	"bytes"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelTarTypes "github.com/Motmedel/utils_go/pkg/tar/types"
	"io"
)

func MakeArchiveFromReader(reader io.Reader) (motmedelTarTypes.Archive, error) {
	if reader == nil {
		return nil, nil
	}

	tarArchive := make(motmedelTarTypes.Archive)

	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, &motmedelErrors.CauseError{
				Message: "An error occurred when obtaining an entry in the tar archive.",
				Cause:   err,
			}
		}
		if header == nil {
			continue
		}

		content, err := io.ReadAll(tarReader)
		if err != nil {
			return nil, &motmedelErrors.CauseError{
				Message: "An error occurred when reading header file content.",
				Cause:   err,
			}
		}

		tarArchive[header.Name] = &motmedelTarTypes.Entry{Header: header, Content: content}
	}

	return tarArchive, nil
}

func MakeTarArchiveFrommData(data []byte) (motmedelTarTypes.Archive, error) {
	if len(data) == 0 {
		return nil, nil
	}

	reader := bytes.NewReader(data)
	tarMap, err := MakeArchiveFromReader(reader)
	if err != nil {
		return nil, &motmedelErrors.InputError{
			Message: "An error occurred when making a tar map from a reader.",
			Cause:   err,
			Input:   reader,
		}
	}

	return tarMap, nil
}

func MakeTarArchive(entries ...*motmedelTarTypes.Entry) motmedelTarTypes.Archive {
	tarArchive := make(motmedelTarTypes.Archive)

	for _, entry := range entries {
		if entry == nil {
			continue
		}

		header := entry.Header
		if header == nil {
			continue
		}

		headerName := header.Name
		if headerName == "" {
			continue
		}

		switch header.Typeflag {
		case tar.TypeReg, tar.TypeDir:
			tarArchive[headerName] = entry
		}
	}

	return tarArchive
}
